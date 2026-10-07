package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

// IPGenerationMode specifies whether generated addresses are random or sequential
type IPGenerationMode int

const (
	ModeRandom IPGenerationMode = iota
	ModeSequential
)

// ParsePrefix parses an IPv6 prefix string (e.g. "2402:800:1234:5678::/64" or "2402:800:1234:5678::")
// and returns the net.IPNet representing the network prefix.
func ParsePrefix(prefixStr string) (*net.IPNet, error) {
	prefixStr = strings.TrimSpace(prefixStr)
	if prefixStr == "64" || prefixStr == "/64" || prefixStr == "48" || prefixStr == "/48" {
		return nil, fmt.Errorf("bạn chỉ mới nhập độ dài prefix '%s' chứ chưa có địa chỉ IP (ví dụ đúng: 2405:4803:c686:2a90::/64). Bạn chỉ cần nhấn Enter để dùng dải mặc định", prefixStr)
	}
	if !strings.Contains(prefixStr, "/") {
		prefixStr += "/64"
	}

	ip, ipNet, err := net.ParseCIDR(prefixStr)
	if err != nil {
		return nil, fmt.Errorf("định dạng IPv6 prefix không hợp lệ: %w", err)
	}

	if ip.To4() != nil {
		return nil, fmt.Errorf("đây là địa chỉ IPv4, vui lòng nhập IPv6 (vd: 2402:800:xxxx:xxxx::/64)")
	}

	ones, bits := ipNet.Mask.Size()
	if bits != 128 {
		return nil, fmt.Errorf("không phải địa chỉ IPv6 128-bit hợp lệ")
	}

	if ones > 120 {
		return nil, fmt.Errorf("prefix /%d quá nhỏ để sinh số lượng địa chỉ IP", ones)
	}

	return ipNet, nil
}

// GenerateIPv6List generates a list of unique IPv6 addresses under the given subnet.
func GenerateIPv6List(ipNet *net.IPNet, count int, mode IPGenerationMode) ([]net.IP, error) {
	if count <= 0 {
		return nil, fmt.Errorf("số lượng IP phải lớn hơn 0")
	}

	ones, _ := ipNet.Mask.Size()
	hostBits := 128 - ones
	if hostBits <= 0 {
		return nil, fmt.Errorf("không còn host bits nào trong subnet")
	}

	seen := make(map[string]bool)
	var result []net.IP

	baseIP := make(net.IP, 16)
	copy(baseIP, ipNet.IP.To16())

	for len(result) < count {
		newIP := make(net.IP, 16)
		copy(newIP, baseIP)

		if mode == ModeRandom {
			// Generate random bytes for the host portion
			randomBytes := make([]byte, 16)
			_, err := rand.Read(randomBytes)
			if err != nil {
				return nil, fmt.Errorf("lỗi sinh số ngẫu nhiên: %w", err)
			}

			// Apply host mask to randomBytes and blend with baseIP
			for i := 0; i < 16; i++ {
				maskByte := ipNet.Mask[i]
				newIP[i] = (baseIP[i] & maskByte) | (randomBytes[i] & ^maskByte)
			}
		} else {
			// Sequential mode
			idx := uint64(len(result) + 1)
			var hostBuf [8]byte
			binary.BigEndian.PutUint64(hostBuf[:], idx)

			// Place into last 8 bytes (assuming standard /64 or host portion)
			copy(newIP[8:], hostBuf[:])
		}

		ipStr := newIP.String()
		if !seen[ipStr] {
			seen[ipStr] = true
			result = append(result, newIP)
		}

		// Safeguard against infinite loop if subnet is exhausted
		if mode == ModeSequential && len(result) == count {
			break
		}
	}

	return result, nil
}

// GenerateSingleRandomIPv6 generates a single random IPv6 address on-the-fly under the subnet.
func GenerateSingleRandomIPv6(ipNet *net.IPNet) net.IP {
	baseIP := ipNet.IP.To16()
	newIP := make(net.IP, 16)
	copy(newIP, baseIP)

	randomBytes := make([]byte, 16)
	_, _ = rand.Read(randomBytes)

	for i := 0; i < 16; i++ {
		maskByte := ipNet.Mask[i]
		newIP[i] = (baseIP[i] & maskByte) | (randomBytes[i] & ^maskByte)
	}
	return newIP
}

// SaveIPListToFile saves IP addresses to a text file (one IP per line)
func SaveIPListToFile(filename string, ips []net.IP) error {
	var builder strings.Builder
	for _, ip := range ips {
		builder.WriteString(ip.String())
		builder.WriteString("\n")
	}
	return os.WriteFile(filename, []byte(builder.String()), 0644)
}
