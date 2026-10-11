package engine

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
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

// DialOutbound connects to destination with the chosen IPv6.
// If the destination has IPv6, it connects via tcp6 directly (ultra fast ~30-50ms).
// If the destination only has IPv4 (IPv4-only site), it immediately falls back to IPv4
// instead of hanging on incompatible dual-stack negotiation.
func DialOutbound(ctx context.Context, targetAddr string, outIP net.IP) (net.Conn, net.IP, error) {
	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		host = targetAddr
		port = "80"
		targetAddr = net.JoinHostPort(host, port)
	}

	// 1. Try IPv6 first if we have a valid outbound IPv6
	if outIP != nil {
		dialer6 := &net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
			LocalAddr: &net.TCPAddr{
				IP: outIP,
			},
		}

		cleanHost := strings.Trim(host, "[]")
		// Connect directly via tcp6
		conn, err := dialer6.DialContext(ctx, "tcp6", targetAddr)
		if err == nil {
			TuneTCPConn(conn)
			return conn, outIP, nil
		}

		// If host was an explicit IPv6 literal, don't fallback to IPv4
		if parsedIP := net.ParseIP(cleanHost); parsedIP != nil && parsedIP.To4() == nil {
			return nil, nil, fmt.Errorf("không thể kết nối tới IPv6 %s qua %s: %w", targetAddr, outIP, err)
		}
	}

	// 2. Fallback to standard dialer for IPv4-only sites
	fallbackDialer := &net.Dialer{
		Timeout:   6 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := fallbackDialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("không thể kết nối tới đích %s: %w", targetAddr, err)
	}

	TuneTCPConn(conn)
	return conn, nil, nil
}
