package engine

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"time"
)

// PickRandomIPv6 picks an IPv6 address uniformly distributed within the given subnet
func PickRandomIPv6(ipNet *net.IPNet) net.IP {
	ip := make(net.IP, len(ipNet.IP))
	copy(ip, ipNet.IP)

	// Fill host bits (the portion not covered by the subnet mask) with secure random bytes
	hostBytes := make([]byte, len(ipNet.Mask))
	_, _ = rand.Read(hostBytes)

	for i := 0; i < len(ip); i++ {
		invertedMask := ^ipNet.Mask[i]
		ip[i] = (ip[i] & ipNet.Mask[i]) | (hostBytes[i] & invertedMask)
	}

	return ip
}

// DialOutboundIPv6 establishes an outbound TCP connection using a specific source IPv6
func DialOutboundIPv6(ctx context.Context, targetAddr string, outIP net.IP) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   12 * time.Second,
		KeepAlive: 30 * time.Second,
		LocalAddr: &net.TCPAddr{
			IP: outIP,
		},
	}

	// Ensure target has port
	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		host = targetAddr
		port = "80"
		targetAddr = net.JoinHostPort(host, port)
	}

	conn, err := dialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới đích %s qua IPv6 %s: %w", targetAddr, outIP, err)
	}

	return conn, nil
}
