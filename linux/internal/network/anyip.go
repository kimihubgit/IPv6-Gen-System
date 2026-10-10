package network

import (
	"fmt"
	"log"
	"net"
	"os/exec"
)

// SetupAnyIP configures Linux kernel route and sysctl for on-the-fly nonlocal binding
func SetupAnyIP(iface string, ipNet *net.IPNet) error {
	sysctlSettings := [][]string{
		{"net.ipv6.ip_nonlocal_bind", "1"},
		{"net.ipv6.conf.all.forwarding", "1"},
		{"fs.file-max", "2097152"},
		{"net.core.somaxconn", "65535"},
		{"net.ipv4.tcp_max_syn_backlog", "65535"},
		{"net.core.netdev_max_backlog", "65535"},
		{"net.ipv6.neigh.default.gc_thresh1", "4096"},
		{"net.ipv6.neigh.default.gc_thresh2", "8192"},
		{"net.ipv6.neigh.default.gc_thresh3", "16384"},
	}

	for _, s := range sysctlSettings {
		_ = exec.Command("sysctl", "-w", fmt.Sprintf("%s=%s", s[0], s[1])).Run()
	}

	if iface != "" {
		_ = exec.Command("ip", "link", "set", iface, "promisc", "on").Run()
		_ = exec.Command("ip", "link", "set", iface, "mtu", "1440").Run()
	}

	prefixStr := ipNet.String()
	// Gán Any-IP vào card ảo Loopback 'lo' (ndppd trên eth0 sẽ trả lời Neighbor Solicitations từ router nhà mạng)
	cmd := exec.Command("ip", "-6", "route", "replace", "local", prefixStr, "dev", "lo")
	if err := cmd.Run(); err != nil {
		log.Printf("⚠️ Cảnh báo gán route local Any-IP: %v", err)
	} else {
		log.Printf("🛡️ Đã kích hoạt Any-IP Route Local cho dải %s -> dev lo (NDP Safe)", prefixStr)
	}

	return nil
}

// TeardownAnyIP cleans up local route on shutdown
func TeardownAnyIP(ipNet *net.IPNet) {
	if ipNet != nil {
		prefixStr := ipNet.String()
		_ = exec.Command("ip", "-6", "route", "del", "local", prefixStr, "dev", "lo").Run()
		log.Printf("🛑 Đã gỡ bỏ Any-IP Route Local: %s", prefixStr)
	}
}
