package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// RotatingProxy holds the list of local IPv6 addresses to rotate through
type RotatingProxy struct {
	anyNet      *net.IPNet // If set, ANY-IP mode (On-the-fly random IPv6 per connection - 0 IP gán card mạng)
	anyIface    string
	IPs         []net.IP
	host        string
	port        int
	counter     uint64
	httpServer  *http.Server
	socksServer net.Listener
	mu          sync.RWMutex
	pool        *RollingPool // If set, delegates IP rotation to dynamic rolling pool
}

// SetupAnyIP configures Linux kernel route and sysctl for on-the-fly nonlocal binding
func SetupAnyIP(iface string, ipNet *net.IPNet) error {
	_ = exec.Command("sysctl", "-w", "net.ipv6.ip_nonlocal_bind=1").Run()
	_ = exec.Command("sysctl", "-w", "net.ipv6.conf.all.forwarding=1").Run()
	_ = exec.Command("ip", "link", "set", iface, "promisc", "on").Run()
	_ = exec.Command("ip", "link", "set", iface, "mtu", "1440").Run()
	prefixStr := ipNet.String()
	// Always bind Any-IP to loopback lo (ndppd on eth0 answers the external router)
	_ = exec.Command("ip", "-6", "route", "replace", "local", prefixStr, "dev", "lo").Run()
	return nil
}

// TeardownAnyIP removes the local route on stop
func TeardownAnyIP(iface string, ipNet *net.IPNet) {
	prefixStr := ipNet.String()
	_ = exec.Command("ip", "-6", "route", "del", "local", prefixStr, "dev", "lo").Run()
}

// NewAnyIPProxy creates a new rotating proxy in ANY-IP mode (0 IPs assigned to card!)
func NewAnyIPProxy(iface string, ipNet *net.IPNet, host string, port int) *RotatingProxy {
	if host == "" {
		host = "0.0.0.0"
	}
	_ = SetupAnyIP(iface, ipNet)
	return &RotatingProxy{
		anyNet:   ipNet,
		anyIface: iface,
		host:     host,
		port:     port,
	}
}

// NewRotatingProxy creates a new rotating proxy instance with static IPs
func NewRotatingProxy(ips []net.IP, host string, port int) *RotatingProxy {
	if host == "" {
		host = "0.0.0.0"
	}
	return &RotatingProxy{
		IPs:  ips,
		host: host,
		port: port,
	}
}

// NewDynamicRotatingProxy creates a new rotating proxy driven by a dynamic rolling pool
func NewDynamicRotatingProxy(pool *RollingPool, host string, port int) *RotatingProxy {
	if host == "" {
		host = "0.0.0.0"
	}
	return &RotatingProxy{
		pool: pool,
		host: host,
		port: port,
	}
}

// PickIPv6 picks an IPv6 from the list or pool using round-robin or ANY-IP
func (p *RotatingProxy) PickIPv6() net.IP {
	if p.anyNet != nil {
		return GenerateSingleRandomIPv6(p.anyNet)
	}
	if p.pool != nil {
		return p.pool.PickIPv6()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.IPs) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&p.counter, 1) % uint64(len(p.IPs))
	return p.IPs[idx]
}

// RandomIPv6 picks a random IPv6 from the list or pool or ANY-IP
func (p *RotatingProxy) RandomIPv6() net.IP {
	if p.anyNet != nil {
		return GenerateSingleRandomIPv6(p.anyNet)
	}
	if p.pool != nil {
		return p.pool.RandomIPv6()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.IPs) == 0 {
		return nil
	}
	return p.IPs[rand.Intn(len(p.IPs))]
}

// Start launches both HTTP(S) proxy and SOCKS5 proxy
func (p *RotatingProxy) Start() error {
	if p.host == "" {
		p.host = "0.0.0.0"
	}

	// Hook termination signals to ensure cleanup of IPs from network card
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Printf("\n🛑 Nhận tín hiệu dừng! Đang tắt Proxy Server...\n")
		p.Stop()
		os.Exit(0)
	}()

	// 1. Start HTTP/HTTPS Proxy on port
	httpAddr := fmt.Sprintf("%s:%d", p.host, p.port)
	p.httpServer = &http.Server{
		Addr: httpAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				p.handleTunneling(w, r)
			} else {
				p.handleHTTP(w, r)
			}
		}),
	}

	// 2. Start SOCKS5 Proxy on port + 1
	socksAddr := fmt.Sprintf("%s:%d", p.host, p.port+1)
	socksListener, err := net.Listen("tcp", socksAddr)
	if err != nil {
		return fmt.Errorf("không thể lắng nghe cổng SOCKS5 %s: %w", socksAddr, err)
	}
	p.socksServer = socksListener

	go func() {
		for {
			conn, err := p.socksServer.Accept()
			if err != nil {
				return
			}
			go p.handleSOCKS5(conn)
		}
	}()

	fmt.Printf("\n🚀 Proxy Server đã khởi động thành công!\n")
	fmt.Printf("   👉 HTTP / HTTPS Proxy:  http://%s:%d\n", p.host, p.port)
	fmt.Printf("   👉 SOCKS5 Proxy:        socks5://%s:%d\n", p.host, p.port+1)
	if p.host == "0.0.0.0" {
		fmt.Printf("   💡 Từ máy Windows, kết nối qua: http://<IP_VPS>:%d hoặc socks5://<IP_VPS>:%d\n", p.port, p.port+1)
	}
	if p.anyNet != nil {
		fmt.Printf("   ⚡ Chế độ: 🚀 ANY-IP KERNEL (0 IP gán card mạng, mỗi request tự sinh 1 IPv6 mới toanh từ %s)\n", p.anyNet.String())
	} else if p.pool != nil {
		fmt.Printf("   ⚡ Chế độ: Dynamic Rolling Pool (%d IPs duy trì, tự động xoay liên tục)\n", p.pool.poolSize)
	} else {
		fmt.Printf("   ⚡ Tổng số IPv6 xoay vòng: %d IPs\n", len(p.IPs))
	}
	fmt.Printf("   Nhấn Ctrl+C để dừng Proxy Server.\n\n")

	return p.httpServer.ListenAndServe()
}

// Stop gracefully stops the proxy servers and any attached rolling pool
func (p *RotatingProxy) Stop() {
	if p.anyNet != nil && p.anyIface != "" {
		TeardownAnyIP(p.anyIface, p.anyNet)
	}
	if p.httpServer != nil {
		p.httpServer.Shutdown(context.Background())
	}
	if p.socksServer != nil {
		p.socksServer.Close()
	}
	if p.pool != nil {
		p.pool.Stop()
	}
}

// dialTarget connects to destination binding to a rotated local IPv6
func (p *RotatingProxy) dialTarget(target string) (net.Conn, net.IP, error) {
	outIP := p.PickIPv6()
	var dialer net.Dialer
	dialer.Timeout = 10 * time.Second

	if outIP != nil {
		dialer.LocalAddr = &net.TCPAddr{
			IP: outIP,
		}
		// 1. Try tcp6 first with rotated IPv6
		conn, err := dialer.Dial("tcp6", target)
		if err == nil {
			log.Printf("[SUCCESS IPv6] %s -> %s", outIP, target)
			return conn, outIP, nil
		}
		log.Printf("[tcp6 failed] outIP=%s, target=%s, err=%v", outIP, target, err)

		// 2. Try generic tcp with rotated IPv6
		conn, err = dialer.Dial("tcp", target)
		if err == nil {
			log.Printf("[SUCCESS generic TCP] %s -> %s", outIP, target)
			return conn, outIP, nil
		}
		log.Printf("[generic tcp failed] outIP=%s, target=%s, err=%v", outIP, target, err)
	}

	// 3. Fallback to default dialer (if destination is IPv4-only)
	log.Printf("[FALLBACK to IPv4] target=%s", target)
	var fallbackDialer net.Dialer
	fallbackDialer.Timeout = 10 * time.Second
	conn, err := fallbackDialer.Dial("tcp", target)
	return conn, nil, err
}

// handleTunneling handles HTTPS CONNECT method
func (p *RotatingProxy) handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, outIP, err := p.dialTarget(r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	if outIP != nil {
		log.Printf("[HTTPS CONNECT] %s -> Outbound IPv6: %s", r.Host, outIP.String())
	}

	// Bidirectional transfer
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}

// handleHTTP handles plain HTTP proxy requests
func (p *RotatingProxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	outIP := p.PickIPv6()
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			dialer.Timeout = 10 * time.Second
			if outIP != nil {
				dialer.LocalAddr = &net.TCPAddr{IP: outIP}
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	outReq.Header = r.Header.Clone()

	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if outIP != nil {
		log.Printf("[HTTP %s] %s -> Outbound IPv6: %s", r.Method, r.URL.Host, outIP.String())
	}

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// handleSOCKS5 implements RFC 1928 SOCKS5 Protocol
func (p *RotatingProxy) handleSOCKS5(client net.Conn) {
	defer client.Close()

	buf := make([]byte, 256)
	// 1. Negotiation version
	if _, err := io.ReadFull(client, buf[:2]); err != nil {
		return
	}
	if buf[0] != 0x05 { // SOCKS5
		return
	}
	numMethods := int(buf[1])
	if _, err := io.ReadFull(client, buf[:numMethods]); err != nil {
		return
	}

	// Respond: 0x05 (SOCKS5), 0x00 (NO AUTH REQUIRED)
	client.Write([]byte{0x05, 0x00})

	// 2. Request details
	if _, err := io.ReadFull(client, buf[:4]); err != nil {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x01 { // 0x01 = CONNECT
		return
	}

	var targetHost string
	switch buf[3] {
	case 0x01: // IPv4
		if _, err := io.ReadFull(client, buf[:4]); err != nil {
			return
		}
		targetHost = net.IP(buf[:4]).String()
	case 0x03: // Domain name
		if _, err := io.ReadFull(client, buf[:1]); err != nil {
			return
		}
		domainLen := int(buf[0])
		if _, err := io.ReadFull(client, buf[:domainLen]); err != nil {
			return
		}
		targetHost = string(buf[:domainLen])
	case 0x04: // IPv6
		if _, err := io.ReadFull(client, buf[:16]); err != nil {
			return
		}
		targetHost = net.IP(buf[:16]).String()
	default:
		return
	}

	// Read Port (2 bytes)
	if _, err := io.ReadFull(client, buf[:2]); err != nil {
		return
	}
	port := (int(buf[0]) << 8) | int(buf[1])
	targetAddr := fmt.Sprintf("%s:%d", targetHost, port)

	// Connect to target with rotated IPv6
	destConn, outIP, err := p.dialTarget(targetAddr)
	if err != nil {
		// Connection refused
		client.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer destConn.Close()

	// Success response: 0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0
	client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	if outIP != nil {
		log.Printf("[SOCKS5 CONNECT] %s -> Outbound IPv6: %s", targetAddr, outIP.String())
	}

	go io.Copy(destConn, client)
	io.Copy(client, destConn)
}
