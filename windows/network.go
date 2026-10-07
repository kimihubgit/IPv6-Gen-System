
package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

// GetSystemInterfaces queries netsh to get friendly network interface names and states on Windows
func GetSystemInterfaces() ([]NetInterface, error) {
	cmd := exec.Command("netsh", "interface", "ipv6", "show", "interfaces")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("không thể đọc danh sách card mạng: %w", err)
	}

	var interfaces []NetInterface
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// Skip header lines until dashes
	foundHeader := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "---") {
			foundHeader = true
			break
		}
	}

	if !foundHeader {
		return nil, fmt.Errorf("không đọc được định dạng netsh interfaces")
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		var idx int
		fmt.Sscanf(fields[0], "%d", &idx)
		state := fields[3]
		name := strings.Join(fields[4:], " ")

		// Filter out loopback and teredo pseudo-interfaces
		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, "loopback") || strings.Contains(nameLower, "teredo") {
			continue
		}

		iface := NetInterface{
			Index: idx,
			Name:  name,
			State: state,
		}

		// Query existing IPv6 addresses for this interface
		iface.IPs = GetInterfaceIPv6Addresses(name)
		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// GetInterfaceIPv6Addresses gets current IPv6 addresses assigned to a Windows interface
func GetInterfaceIPv6Addresses(ifaceName string) []net.IP {
	cmd := exec.Command("netsh", "interface", "ipv6", "show", "addresses", ifaceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var ips []net.IP
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Address ") && strings.Contains(line, " Parameters") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipStr := parts[1]
				if idx := strings.Index(ipStr, "%"); idx != -1 {
					ipStr = ipStr[:idx]
				}
				if ip := net.ParseIP(ipStr); ip != nil {
					ips = append(ips, ip)
				}
			}
		}
	}
	return ips
}

// AddAddressToWindows runs netsh to add a single IPv6 address
func AddAddressToWindows(ifaceName string, ip net.IP, prefixLen int, storeType string, skipAsSource bool) error {
	skipStr := "false"
	if skipAsSource {
		skipStr = "true"
	}

	addrWithPrefix := fmt.Sprintf("%s/%d", ip.String(), prefixLen)
	cmd := exec.Command("netsh", "interface", "ipv6", "add", "address",
		fmt.Sprintf("interface=%s", ifaceName),
		fmt.Sprintf("address=%s", addrWithPrefix),
		fmt.Sprintf("store=%s", storeType),
		fmt.Sprintf("skipassource=%s", skipStr),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// DeleteAddressFromWindows runs netsh to remove an IPv6 address
func DeleteAddressFromWindows(ifaceName string, ipStr string, storeType string) error {
	cmd := exec.Command("netsh", "interface", "ipv6", "delete", "address",
		fmt.Sprintf("interface=%s", ifaceName),
		fmt.Sprintf("address=%s", ipStr),
		fmt.Sprintf("store=%s", storeType),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// BatchAddIPv6 adds a slice of IPv6 addresses concurrently using a worker pool on Windows
func BatchAddIPv6(ifaceName string, ips []net.IP, prefixLen int, storeType string, skipAsSource bool, onProgress func(done, total int)) (int, []error) {
	total := len(ips)
	concurrency := 12
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
				err := AddAddressToWindows(ifaceName, ip, prefixLen, storeType, skipAsSource)
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

// BatchDeleteIPv6 removes a list of IPv6 strings concurrently on Windows
func BatchDeleteIPv6(ifaceName string, ips []string, storeType string, onProgress func(done, total int)) (int, []error) {
	total := len(ips)
	concurrency := 12
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
				err := DeleteAddressFromWindows(ifaceName, ip, storeType)
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
