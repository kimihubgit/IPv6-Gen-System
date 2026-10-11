package engine

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// PickRandomIPv6 picks an IPv6 address uniformly distributed within the given subnet.
// Optimized for /64 subnets to execute in nanoseconds without allocations.
func PickRandomIPv6(ipNet *net.IPNet) net.IP {
	baseIP := ipNet.IP.To16()
	if baseIP == nil {
		return ipNet.IP
	}

	newIP := make(net.IP, 16)
	copy(newIP, baseIP)

	ones, _ := ipNet.Mask.Size()
	if ones == 64 {
		// Standard /64: top 8 bytes are network prefix, bottom 8 bytes are random host ID
		var hostBuf [8]byte
		_, _ = rand.Read(hostBuf[:])
		// Avoid all-zeros or router anycast
		if binary.BigEndian.Uint64(hostBuf[:]) == 0 {
			hostBuf[7] = 0x01
		}
		copy(newIP[8:], hostBuf[:])
		return newIP
	}

	// Generic subnet mask
	randomBytes := make([]byte, 16)
	_, _ = rand.Read(randomBytes)
	for i := 0; i < 16; i++ {
		maskByte := ipNet.Mask[i]
		newIP[i] = (baseIP[i] & maskByte) | (randomBytes[i] & ^maskByte)
	}
	return newIP
}

// TuneTCPConn sets TCP_NODELAY and KeepAlive to ensure lowest possible latency
func TuneTCPConn(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}

// In-memory DNS cache to avoid hammering systemd-resolved across concurrent threads
type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
}

var (
	dnsCacheMu sync.RWMutex
	dnsCache   = make(map[string]dnsCacheEntry)
)

func resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	cleanHost := strings.Trim(host, "[]")
	if parsed := net.ParseIP(cleanHost); parsed != nil {
		return []net.IP{parsed}, nil
	}

	now := time.Now()
	dnsCacheMu.RLock()
	entry, ok := dnsCache[cleanHost]
	dnsCacheMu.RUnlock()
	if ok && now.Before(entry.expiresAt) {
		return entry.ips, nil
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", cleanHost)
	if err != nil {
		return nil, err
	}

	dnsCacheMu.Lock()
	dnsCache[cleanHost] = dnsCacheEntry{
		ips:       ips,
		expiresAt: now.Add(2 * time.Minute),
	}
	dnsCacheMu.Unlock()
	return ips, nil
}

// DialOutbound connects to destination with the chosen IPv6.
// If the destination has IPv6, it connects via tcp6 directly (ultra fast ~30-50ms).
// If the destination only has IPv4 (IPv4-only site), it immediately connects via IPv4.
func DialOutbound(ctx context.Context, targetAddr string, outIP net.IP) (net.Conn, net.IP, error) {
	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		host = targetAddr
		port = "80"
	}

	// 1. Resolve host using zero-latency in-memory DNS cache
	ips, err := resolveHost(ctx, host)
	if err != nil {
		// Fallback to direct dial if DNS resolve failed
		fallbackDialer := &net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		conn, dialErr := fallbackDialer.DialContext(ctx, "tcp", targetAddr)
		if dialErr == nil {
			TuneTCPConn(conn)
			return conn, nil, nil
		}
		return nil, nil, fmt.Errorf("không thể phân giải và kết nối tới %s: %w", targetAddr, err)
	}

	var ipv6List []net.IP
	var ipv4List []net.IP
	for _, ip := range ips {
		if ip.To4() == nil {
			ipv6List = append(ipv6List, ip)
		} else {
			ipv4List = append(ipv4List, ip)
		}
	}

	// 2. Try IPv6 first if destination supports IPv6 and we have outIP
	if len(ipv6List) > 0 && outIP != nil {
		dialer6 := &net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			LocalAddr: &net.TCPAddr{
				IP: outIP,
			},
		}

		for _, dstIP := range ipv6List {
			dstAddr := net.JoinHostPort(dstIP.String(), port)
			conn, dialErr := dialer6.DialContext(ctx, "tcp6", dstAddr)
			if dialErr == nil {
				TuneTCPConn(conn)
				return conn, outIP, nil
			}
		}

		// If destination was an explicit IPv6 literal, don't fallback to IPv4
		cleanHost := strings.Trim(host, "[]")
		if parsed := net.ParseIP(cleanHost); parsed != nil && parsed.To4() == nil {
			return nil, nil, fmt.Errorf("không thể kết nối tới IPv6 %s qua %s", targetAddr, outIP)
		}
	}

	// 3. Fallback to IPv4 if destination is IPv4 or IPv6 dial failed
	fallbackDialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if len(ipv4List) > 0 {
		for _, dstIP := range ipv4List {
			dstAddr := net.JoinHostPort(dstIP.String(), port)
			conn, dialErr := fallbackDialer.DialContext(ctx, "tcp", dstAddr)
			if dialErr == nil {
				TuneTCPConn(conn)
				return conn, nil, nil
			}
		}
	}

	// Final generic dial
	conn, err := fallbackDialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("không thể kết nối tới đích %s: %w", targetAddr, err)
	}

	TuneTCPConn(conn)
	return conn, nil, nil
}
