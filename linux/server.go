package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CountingConn wraps a net.Conn to count bytes read and written
type CountingConn struct {
	net.Conn
	account *ProxyAccount
}

func (c *CountingConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 && c.account != nil {
		c.account.AddTraffic(int64(n))
	}
	return n, err
}

func (c *CountingConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 && c.account != nil {
		c.account.AddTraffic(int64(n))
	}
	return n, err
}

// PortListener manages a TCP listener on a dedicated port for a ProxyAccount
type PortListener struct {
	account  *ProxyAccount
	listener net.Listener
	ipNet    *net.IPNet
	store    *ProxyStore
	stopChan chan struct{}
}

// MultiProxyManager manages dynamic port listeners for all proxy accounts
type MultiProxyManager struct {
	ipNet     *net.IPNet
	ifaceName string
	store     *ProxyStore
	listeners map[int]*PortListener // key = Port
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// NewMultiProxyManager creates a manager for multi-port proxy accounts
func NewMultiProxyManager(ifaceName string, ipNet *net.IPNet, store *ProxyStore) *MultiProxyManager {
	return &MultiProxyManager{
		ipNet:     ipNet,
		ifaceName: ifaceName,
		store:     store,
		listeners: make(map[int]*PortListener),
		stopChan:  make(chan struct{}),
	}
}

// Start launches listeners for all enabled accounts and background tasks
func (m *MultiProxyManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Setup Linux ANY-IP kernel route (ndppd will answer NDP for the prefix)
	_ = SetupAnyIP(m.ifaceName, m.ipNet)

	// 2. Start listeners for all accounts
	for _, acc := range m.store.GetAll() {
		if acc.Enabled {
			if err := m.startListenerLocked(acc); err != nil {
				log.Printf("⚠️ Không thể mở port %d cho proxy '%s': %v", acc.Port, acc.Name, err)
			}
		}
	}

	// 3. Periodic persistence & quota checker routine
	go m.backgroundLoop()

	return nil
}

// Stop shuts down all listeners
func (m *MultiProxyManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.stopChan)
	for port, pl := range m.listeners {
		_ = pl.listener.Close()
		delete(m.listeners, port)
	}
	_ = m.store.Save()
}

// SyncAccount starts, updates, or stops a listener when an account is modified
func (m *MultiProxyManager) SyncAccount(acc *ProxyAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing listener on this port if any
	if pl, exists := m.listeners[acc.Port]; exists {
		_ = pl.listener.Close()
		delete(m.listeners, acc.Port)
	}

	if acc.Enabled {
		return m.startListenerLocked(acc)
	}
	return nil
}

// RemoveAccount stops the listener for a deleted account
func (m *MultiProxyManager) RemoveAccount(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pl, exists := m.listeners[port]; exists {
		_ = pl.listener.Close()
		delete(m.listeners, port)
	}
}

func (m *MultiProxyManager) startListenerLocked(acc *ProxyAccount) error {
	addr := fmt.Sprintf("0.0.0.0:%d", acc.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("không thể lắng nghe cổng %s: %w", addr, err)
	}

	pl := &PortListener{
		account:  acc,
		listener: l,
		ipNet:    m.ipNet,
		store:    m.store,
		stopChan: make(chan struct{}),
	}
	m.listeners[acc.Port] = pl

	go pl.serve()
	log.Printf("✅ Đã kích hoạt proxy [%s] lắng nghe cổng %d (HTTP + SOCKS5)", acc.Name, acc.Port)
	return nil
}

func (m *MultiProxyManager) backgroundLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			// Auto save bandwidth usage & accounts to disk
			_ = m.store.Save()
		}
	}
}

// serve handles connections for a single port listener
func (pl *PortListener) serve() {
	for {
		rawConn, err := pl.listener.Accept()
		if err != nil {
			return // Listener closed
		}

		go pl.handleConn(rawConn)
	}
}

// handleConn determines if incoming connection is SOCKS5 (0x05) or HTTP/HTTPS
func (pl *PortListener) handleConn(rawConn net.Conn) {
	conn := &CountingConn{
		Conn:    rawConn,
		account: pl.account,
	}
	defer conn.Close()

	// Check if proxy account is valid
	if ok, reason := pl.account.CanAccess(); !ok {
		log.Printf("[Từ chối Port %d - %s] %s", pl.account.Port, pl.account.Name, reason)
		return
	}

	pl.account.LastActive = time.Now()

	// Peek first byte to detect protocol
	bufReader := bufio.NewReader(conn)
	firstByte, err := bufReader.Peek(1)
	if err != nil {
		return
	}

	if firstByte[0] == 0x05 {
		// SOCKS5 protocol
		pl.handleSOCKS5(conn, bufReader)
	} else {
		// HTTP / HTTPS CONNECT protocol
		pl.handleHTTP(conn, bufReader)
	}
}

// PickIPv6 picks the outbound IPv6 according to account's rotation policy
func (pl *PortListener) PickIPv6() net.IP {
	acc := pl.account

	switch acc.RotationType {
	case RotationStatic:
		if acc.StaticIP != "" {
			if ip := net.ParseIP(acc.StaticIP); ip != nil {
				return ip
			}
		}
		newIP := GenerateSingleRandomIPv6(pl.ipNet)
		acc.StaticIP = newIP.String()
		_ = pl.store.Save()
		return newIP

	case RotationSticky:
		acc.mu.Lock()
		defer acc.mu.Unlock()
		if acc.currentStickyIP != "" && time.Now().Before(acc.stickyExpireAt) {
			if ip := net.ParseIP(acc.currentStickyIP); ip != nil {
				return ip
			}
		}
		newIP := GenerateSingleRandomIPv6(pl.ipNet)
		acc.currentStickyIP = newIP.String()
		sec := acc.StickySec
		if sec <= 0 {
			sec = 60
		}
		acc.stickyExpireAt = time.Now().Add(time.Duration(sec) * time.Second)
		return newIP

	case RotationPerRequest:
		fallthrough
	default:
		return GenerateSingleRandomIPv6(pl.ipNet)
	}
}

// dialTarget connects outbound binding to the chosen IPv6
func (pl *PortListener) dialTarget(target string) (net.Conn, net.IP, error) {
	outIP := pl.PickIPv6()
	var dialer net.Dialer
	dialer.Timeout = 10 * time.Second

	if outIP != nil {
		dialer.LocalAddr = &net.TCPAddr{
			IP: outIP,
		}
		// 1. Try tcp6 first
		conn, err := dialer.Dial("tcp6", target)
		if err == nil {
			return conn, outIP, nil
		}
		// 2. Try generic tcp
		conn, err = dialer.Dial("tcp", target)
		if err == nil {
			return conn, outIP, nil
		}
	}

	// 3. Fallback to direct dialer (if destination is IPv4-only)
	var fallbackDialer net.Dialer
	fallbackDialer.Timeout = 10 * time.Second
	conn, err := fallbackDialer.Dial("tcp", target)
	return conn, nil, err
}

// ----------------------------------------------------------------------
// HTTP / HTTPS CONNECT HANDLER
// ----------------------------------------------------------------------

func (pl *PortListener) verifyHTTPAuth(r *http.Request) bool {
	if pl.account.Username == "" && pl.account.Password == "" {
		return true // Không yêu cầu auth
	}

	authHeader := r.Header.Get("Proxy-Authorization")
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
		return false
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[0] == pl.account.Username && pair[1] == pl.account.Password
}

func (pl *PortListener) handleHTTP(conn *CountingConn, reader *bufio.Reader) {
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	// Check HTTP Proxy Authentication
	if !pl.verifyHTTPAuth(req) {
		resp := "HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"IPv6 Rotating Proxy\"\r\nContent-Length: 0\r\n\r\n"
		_, _ = conn.Write([]byte(resp))
		return
	}

	if req.Method == http.MethodConnect {
		pl.handleHTTPSConnect(conn, req)
	} else {
		pl.handlePlainHTTP(conn, req)
	}
}

func (pl *PortListener) handleHTTPSConnect(clientConn *CountingConn, req *http.Request) {
	destConn, outIP, err := pl.dialTarget(req.Host)
	if err != nil {
		resp := fmt.Sprintf("HTTP/1.1 503 Service Unavailable\r\n\r\n%s", err.Error())
		_, _ = clientConn.Write([]byte(resp))
		return
	}
	defer destConn.Close()

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	if outIP != nil {
		log.Printf("[Port %d][%s] CONNECT %s -> Outbound IPv6: %s", pl.account.Port, pl.account.Name, req.Host, outIP.String())
	}

	destCounting := &CountingConn{
		Conn:    destConn,
		account: pl.account,
	}

	go io.Copy(destCounting, clientConn)
	io.Copy(clientConn, destCounting)
}

func (pl *PortListener) handlePlainHTTP(clientConn *CountingConn, req *http.Request) {
	destHost := req.URL.Host
	if !strings.Contains(destHost, ":") {
		if req.URL.Scheme == "https" {
			destHost += ":443"
		} else {
			destHost += ":80"
		}
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, outIP, err := pl.dialTarget(addr)
			if outIP != nil {
				log.Printf("[Port %d][%s] HTTP %s -> Outbound IPv6: %s", pl.account.Port, pl.account.Name, req.URL.Host, outIP.String())
			}
			return &CountingConn{Conn: conn, account: pl.account}, err
		},
	}

	outReq, err := http.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), req.Body)
	if err != nil {
		resp := "HTTP/1.1 400 Bad Request\r\n\r\n"
		_, _ = clientConn.Write([]byte(resp))
		return
	}
	outReq.Header = req.Header.Clone()
	outReq.Header.Del("Proxy-Authorization") // Bỏ auth header trước khi chuyển ra web đích

	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		respErr := fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\n\r\n%s", err.Error())
		_, _ = clientConn.Write([]byte(respErr))
		return
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_ = resp.Write(&buf)
	_, _ = clientConn.Write(buf.Bytes())
}

// ----------------------------------------------------------------------
// SOCKS5 HANDLER (RFC 1928 + RFC 1929 Username/Password Authentication)
// ----------------------------------------------------------------------

func (pl *PortListener) handleSOCKS5(clientConn *CountingConn, reader *bufio.Reader) {
	buf := make([]byte, 256)

	// 1. Version identifier / method selection message
	// +----+----------+----------+
	// |VER | NMETHODS | METHODS  |
	// +----+----------+----------+
	if _, err := io.ReadFull(reader, buf[:2]); err != nil {
		return
	}
	if buf[0] != 0x05 {
		return
	}
	nmethods := int(buf[1])
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(reader, methods); err != nil {
		return
	}

	hasNoAuth := false
	hasUserPass := false
	for _, m := range methods {
		if m == 0x00 {
			hasNoAuth = true
		}
		if m == 0x02 {
			hasUserPass = true
		}
	}

	// Determine authentication method
	needAuth := pl.account.Username != "" || pl.account.Password != ""

	if needAuth {
		if !hasUserPass {
			// No acceptable methods
			_, _ = clientConn.Write([]byte{0x05, 0xFF})
			return
		}
		// Choose 0x02 (Username/Password)
		_, _ = clientConn.Write([]byte{0x05, 0x02})

		// Perform RFC 1929 Auth negotiation
		// +----+------+----------+------+----------+
		// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
		// +----+------+----------+------+----------+
		if _, err := io.ReadFull(reader, buf[:2]); err != nil {
			return
		}
		if buf[0] != 0x01 { // Subnegotiation version 1
			return
		}
		ulen := int(buf[1])
		uname := make([]byte, ulen)
		if _, err := io.ReadFull(reader, uname); err != nil {
			return
		}

		if _, err := io.ReadFull(reader, buf[:1]); err != nil {
			return
		}
		plen := int(buf[0])
		passwd := make([]byte, plen)
		if _, err := io.ReadFull(reader, passwd); err != nil {
			return
		}

		if string(uname) != pl.account.Username || string(passwd) != pl.account.Password {
			// Auth failed: 0x01 (version), 0x01 (failure status)
			_, _ = clientConn.Write([]byte{0x01, 0x01})
			return
		}
		// Auth success: 0x01 (version), 0x00 (success status)
		_, _ = clientConn.Write([]byte{0x01, 0x00})
	} else {
		if !hasNoAuth {
			_, _ = clientConn.Write([]byte{0x05, 0xFF})
			return
		}
		// 0x00 (No authentication required)
		_, _ = clientConn.Write([]byte{0x05, 0x00})
	}

	// 2. SOCKS5 Request
	// +----+-----+-------+------+----------+----------+
	// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +----+-----+-------+------+----------+----------+
	if _, err := io.ReadFull(reader, buf[:4]); err != nil {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x01 { // Only CMD = 0x01 (CONNECT) supported
		_, _ = clientConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // Command not supported
		return
	}

	var targetHost string
	switch buf[3] {
	case 0x01: // IPv4
		ipv4 := make([]byte, 4)
		if _, err := io.ReadFull(reader, ipv4); err != nil {
			return
		}
		targetHost = net.IP(ipv4).String()
	case 0x03: // Domain name
		if _, err := io.ReadFull(reader, buf[:1]); err != nil {
			return
		}
		domainLen := int(buf[0])
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(reader, domain); err != nil {
			return
		}
		targetHost = string(domain)
	case 0x04: // IPv6
		ipv6 := make([]byte, 16)
		if _, err := io.ReadFull(reader, ipv6); err != nil {
			return
		}
		targetHost = net.IP(ipv6).String()
	default:
		_, _ = clientConn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var portBuf [2]byte
	if _, err := io.ReadFull(reader, portBuf[:]); err != nil {
		return
	}
	targetPort := (int(portBuf[0]) << 8) | int(portBuf[1])
	targetAddr := fmt.Sprintf("%s:%d", targetHost, targetPort)

	// Dial target via outbound rotated IPv6
	destConn, outIP, err := pl.dialTarget(targetAddr)
	if err != nil {
		_, _ = clientConn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // Connection refused
		return
	}
	defer destConn.Close()

	// SOCKS5 reply success
	_, _ = clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	if outIP != nil {
		log.Printf("[Port %d][%s] SOCKS5 %s -> Outbound IPv6: %s", pl.account.Port, pl.account.Name, targetAddr, outIP.String())
	}

	destCounting := &CountingConn{
		Conn:    destConn,
		account: pl.account,
	}

	go io.Copy(destCounting, clientConn)
	io.Copy(clientConn, destCounting)
}
