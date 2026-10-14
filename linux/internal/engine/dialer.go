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
		if parsed.To4() != nil {
			return nil, fmt.Errorf("địa chỉ %s là IPv4, hệ thống chỉ hỗ trợ thuần IPv6 (Strict IPv6 Only)", cleanHost)
		}
		return []net.IP{parsed}, nil
	}

	now := time.Now()
	dnsCacheMu.RLock()
	entry, ok := dnsCache[cleanHost]
	dnsCacheMu.RUnlock()
	if ok && now.Before(entry.expiresAt) {
		return entry.ips, nil
	}

	// 2-second timeout to prevent DNS lookup hanging under high load
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Query specifically IPv6 (AAAA records) only!
	ips, err := net.DefaultResolver.LookupIP(lookupCtx, "ip6", cleanHost)
	if err != nil || len(ips) == 0 {
		// If lookup failed but we have a slightly expired cache, use it as emergency fallback
		if ok && len(entry.ips) > 0 {
			return entry.ips, nil
		}
		if err != nil {
			return nil, fmt.Errorf("không thể phân giải IPv6 cho %s (web không hỗ trợ IPv6 hoặc lỗi DNS): %w", cleanHost, err)
		}
		return nil, fmt.Errorf("trang web %s không hỗ trợ địa chỉ IPv6 (Strict IPv6 Only)", cleanHost)
	}

	dnsCacheMu.Lock()
	// Gentle cache pruning if too many domains
	if len(dnsCache) > 2000 {
		for k, v := range dnsCache {
			if now.After(v.expiresAt) {
				delete(dnsCache, k)
			}
		}
	}
	dnsCache[cleanHost] = dnsCacheEntry{
		ips:       ips,
		expiresAt: now.Add(5 * time.Minute),
	}
	dnsCacheMu.Unlock()
	return ips, nil
}

// DialOutbound connects strictly to destination via the chosen rotated IPv6.
// Strict IPv6 Only: Absolutely NO IPv4 fallback to guarantee 100% IPv6 anonymity and isolation.
func DialOutbound(ctx context.Context, targetAddr string, outIP net.IP) (net.Conn, net.IP, error) {
	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		host = targetAddr
		port = "80"
	}

	if outIP == nil {
		return nil, nil, fmt.Errorf("không có địa chỉ IPv6 gửi đi (outIP is nil)")
	}

	// 1. Resolve host using zero-latency in-memory DNS cache (Strict IPv6 AAAA)
	ips, err := resolveHost(ctx, host)
	if err != nil {
		return nil, nil, err
	}

	var ipv6List []net.IP
	for _, ip := range ips {
		if ip.To4() == nil {
			ipv6List = append(ipv6List, ip)
		}
	}

	if len(ipv6List) == 0 {
		return nil, nil, fmt.Errorf("trang web %s không hỗ trợ địa chỉ IPv6 (Strict IPv6 Only)", host)
	}

	// 2. Connect strictly via chosen IPv6
	dialer6 := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 30 * time.Second,
		LocalAddr: &net.TCPAddr{
			IP: outIP,
		},
	}

	var lastErr error
	maxTries := len(ipv6List)
	if maxTries > 3 {
		maxTries = 3
	}

	for i := 0; i < maxTries; i++ {
		dstIP := ipv6List[i]
		dstAddr := net.JoinHostPort(dstIP.String(), port)
		conn, dialErr := dialer6.DialContext(ctx, "tcp6", dstAddr)
		if dialErr == nil {
			TuneTCPConn(conn)
			return conn, outIP, nil
		}
		lastErr = dialErr
	}

	return nil, nil, fmt.Errorf("không thể kết nối tới IPv6 của %s qua %s: %w", targetAddr, outIP, lastErr)
}
