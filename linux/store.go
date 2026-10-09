package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// RotationPolicy defines how IPv6 addresses are rotated for this proxy
type RotationPolicy string

const (
	RotationPerRequest RotationPolicy = "request" // Mỗi request 1 IP ngẫu nhiên mới
	RotationSticky     RotationPolicy = "sticky"  // Giữ nguyên 1 IP trong X giây
	RotationStatic     RotationPolicy = "static"  // Cố định 1 IP duy nhất
)

// ProxyAccount represents a commercial proxy package allocated on a dedicated port
type ProxyAccount struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`          // Tên khách hàng / Ghi chú
	Port         int            `json:"port"`          // Cổng kết nối (ví dụ: 10001)
	Username     string         `json:"username"`      // Tài khoản auth
	Password     string         `json:"password"`      // Mật khẩu auth
	Proto        string         `json:"proto"`         // "both" (HTTP+SOCKS5) | "http" | "socks5"
	MaxBytes     int64          `json:"max_bytes"`     // Giới hạn dung lượng (bytes). 0 = Không giới hạn
	BytesUsed    int64          `json:"bytes_used"`    // Số bytes đã sử dụng
	ExpiresAt    *time.Time     `json:"expires_at"`    // Thời gian hết hạn (nil = Vĩnh viễn)
	RotationType RotationPolicy `json:"rotation_type"` // "request" | "sticky" | "static"
	StickySec    int            `json:"sticky_sec"`    // Số giây giữ IP (nếu sticky)
	StaticIP     string         `json:"static_ip"`     // IP cố định (nếu static)
	Enabled      bool           `json:"enabled"`       // Bật / Khóa proxy
	CreatedAt    time.Time      `json:"created_at"`
	LastActive   time.Time      `json:"last_active"`

	// Runtime fields (không lưu vào json)
	currentStickyIP   string    `json:"-"`
	stickyExpireAt    time.Time `json:"-"`
	mu                sync.RWMutex `json:"-"`
}

// IsExpired checks if the account has passed its expiration date
func (p *ProxyAccount) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}

// IsTrafficExceeded checks if the account has consumed all allocated bandwidth
func (p *ProxyAccount) IsTrafficExceeded() bool {
	if p.MaxBytes <= 0 {
		return false
	}
	return atomic.LoadInt64(&p.BytesUsed) >= p.MaxBytes
}

// CanAccess returns true if the proxy is enabled, not expired, and bandwidth not exceeded
func (p *ProxyAccount) CanAccess() (bool, string) {
	if !p.Enabled {
		return false, "Proxy đã bị tạm khóa"
	}
	if p.IsExpired() {
		return false, "Proxy đã hết hạn sử dụng"
	}
	if p.IsTrafficExceeded() {
		return false, "Proxy đã sử dụng hết dung lượng (GB)"
	}
	return true, ""
}

// AddTraffic increments used bytes atomically
func (p *ProxyAccount) AddTraffic(bytes int64) {
	if bytes > 0 {
		atomic.AddInt64(&p.BytesUsed, bytes)
	}
}

// ProxyStore manages persistent proxy accounts in a JSON file
type ProxyStore struct {
	filePath string
	mu       sync.RWMutex
	accounts map[string]*ProxyAccount // key = ID
	portMap  map[int]*ProxyAccount    // key = Port
}

// NewProxyStore creates or loads a proxy store from disk
func NewProxyStore(dataFile string) (*ProxyStore, error) {
	if dataFile == "" {
		dataFile = "data/proxies.json"
	}
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	store := &ProxyStore{
		filePath: dataFile,
		accounts: make(map[string]*ProxyAccount),
		portMap:  make(map[int]*ProxyAccount),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *ProxyStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return nil // File chưa tồn tại, bắt đầu rỗng
	}
	if err != nil {
		return err
	}

	var list []*ProxyAccount
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("lỗi đọc file lưu trữ proxies: %w", err)
	}

	s.accounts = make(map[string]*ProxyAccount)
	s.portMap = make(map[int]*ProxyAccount)

	for _, acc := range list {
		s.accounts[acc.ID] = acc
		s.portMap[acc.Port] = acc
	}
	return nil
}

// Save writes current accounts to disk
func (s *ProxyStore) Save() error {
	s.mu.RLock()
	var list []*ProxyAccount
	for _, acc := range s.accounts {
		list = append(list, acc)
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, s.filePath)
}

// GetAll returns all accounts
func (s *ProxyStore) GetAll() []*ProxyAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ProxyAccount
	for _, acc := range s.accounts {
		result = append(result, acc)
	}
	return result
}

// GetByID returns account by ID
func (s *ProxyStore) GetByID(id string) *ProxyAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accounts[id]
}

// GetByPort returns account listening on specific port
func (s *ProxyStore) GetByPort(port int) *ProxyAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.portMap[port]
}

// Add creates or updates an account
func (s *ProxyStore) Add(acc *ProxyAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check port conflict with other accounts
	if existing, ok := s.portMap[acc.Port]; ok && existing.ID != acc.ID {
		return fmt.Errorf("cổng (Port) %d đã được sử dụng bởi proxy '%s'", acc.Port, existing.Name)
	}

	s.accounts[acc.ID] = acc
	s.portMap[acc.Port] = acc
	return nil
}

// Delete removes an account
func (s *ProxyStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc, ok := s.accounts[id]
	if !ok {
		return fmt.Errorf("không tìm thấy proxy có ID: %s", id)
	}

	delete(s.portMap, acc.Port)
	delete(s.accounts, id)
	return nil
}

// NextAvailablePort finds the next available port starting from basePort
func (s *ProxyStore) NextAvailablePort(basePort int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p := basePort
	for {
		if _, used := s.portMap[p]; !used {
			return p
		}
		p++
	}
}
