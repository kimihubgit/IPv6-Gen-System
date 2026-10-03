//go:build linux

package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
)

// GetSystemInterfaces returns Linux network interfaces excluding loopbacks
func GetSystemInterfaces() ([]NetInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("không thể đọc danh sách card mạng Linux: %w", err)
	}

	var result []NetInterface
	for _, iface := range ifaces {
		// Ignore loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		state := "Disconnected"
		if iface.Flags&net.FlagUp != 0 {
			state = "Connected"
		}

		ips := GetInterfaceIPv6Addresses(iface.Name)
		result = append(result, NetInterface{
			Index: iface.Index,
			Name:  iface.Name,
			State: state,
			IPs:   ips,
		})
	}

	return result, nil
}

// GetInterfaceIPv6Addresses gets current IPv6 addresses assigned to a Linux interface
func GetInterfaceIPv6Addresses(ifaceName string) []net.IP {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil
	}

	var ips []net.IP
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && ip.To4() == nil {
			ips = append(ips, ip)
		}
	}
	return ips
}

// AddAddressToLinux adds an IPv6 address to an interface using ip command
func AddAddressToLinux(ifaceName string, ip net.IP, prefixLen int) error {
	addrWithPrefix := fmt.Sprintf("%s/%d", ip.String(), prefixLen)
	cmd := exec.Command("ip", "-6", "addr", "add", addrWithPrefix, "dev", ifaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(output))
		if strings.Contains(outStr, "File exists") {
			return nil // Already added
		}
		return fmt.Errorf("%w: %s", err, outStr)
	}
	return nil
}

// DeleteAddressFromLinux deletes an IPv6 address from an interface using ip command
func DeleteAddressFromLinux(ifaceName string, ipStr string) error {
	if !strings.Contains(ipStr, "/") {
		ipStr = ipStr + "/64"
	}
	cmd := exec.Command("ip", "-6", "addr", "del", ipStr, "dev", ifaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(output))
		if strings.Contains(outStr, "Cannot assign requested address") || strings.Contains(outStr, "No such process") {
			return nil // Already deleted
		}
		return fmt.Errorf("%w: %s", err, outStr)
	}
	return nil
}

// ensureLinuxMaxAddresses tunes kernel to allow unlimited IPv6 addresses on the interface
func ensureLinuxMaxAddresses(ifaceName string) {
	_ = exec.Command("sysctl", "-w", "net.ipv6.conf.all.max_addresses=0").Run()
	_ = exec.Command("sysctl", "-w", fmt.Sprintf("net.ipv6.conf.%s.max_addresses=0", ifaceName)).Run()
}

// BatchAddIPv6 adds a slice of IPv6 addresses concurrently on Linux
func BatchAddIPv6(ifaceName string, ips []net.IP, prefixLen int, storeType string, skipAsSource bool, onProgress func(done, total int)) (int, []error) {
	ensureLinuxMaxAddresses(ifaceName)

	total := len(ips)
	concurrency := 16
	if total < concurrency {
		concurrency = total
	}

	jobs := make(chan net.IP, total)
	for _, ip := range ips {
		jobs <- ip
	}
	close(jobs)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	successCount := 0
	processed := 0

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				err := AddAddressToLinux(ifaceName, ip, prefixLen)
				mu.Lock()
				processed++
				if err != nil {
					errs = append(errs, fmt.Errorf("lỗi thêm %s: %v", ip, err))
				} else {
					successCount++
				}
				if onProgress != nil {
					onProgress(processed, total)
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return successCount, errs
}

// BatchDeleteIPv6 removes a list of IPv6 strings concurrently on Linux
func BatchDeleteIPv6(ifaceName string, ips []string, storeType string, onProgress func(done, total int)) (int, []error) {
	total := len(ips)
	concurrency := 16
	if total < concurrency {
		concurrency = total
	}

	jobs := make(chan string, total)
	for _, ip := range ips {
		jobs <- ip
	}
	close(jobs)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	successCount := 0
	processed := 0

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				err := DeleteAddressFromLinux(ifaceName, ip)
				mu.Lock()
				processed++
				if err != nil {
					errs = append(errs, fmt.Errorf("lỗi xóa %s: %v", ip, err))
				} else {
					successCount++
				}
				if onProgress != nil {
					onProgress(processed, total)
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return successCount, errs
}
