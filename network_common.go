package main

import (
	"fmt"
	"net"
)

// NetInterface holds basic info about a network adapter
type NetInterface struct {
	Index int
	Name  string
	State string
	IPs   []net.IP
}

// DetectGlobalPrefix finds any public/global IPv6 or ULA address on the interface and returns its /64 prefix
func DetectGlobalPrefix(iface NetInterface) string {
	for _, ip := range iface.IPs {
		// Ignore link-local (fe80::) and loopback (::1)
		if ip.IsLinkLocalUnicast() || ip.IsLoopback() || ip.IsMulticast() {
			continue
		}

		// Take the first 64 bits (8 bytes)
		ip16 := ip.To16()
		if ip16 == nil {
			continue
		}

		prefixIP := make(net.IP, 16)
		copy(prefixIP[:8], ip16[:8])
		return fmt.Sprintf("%s/64", prefixIP.String())
	}
	return ""
}
