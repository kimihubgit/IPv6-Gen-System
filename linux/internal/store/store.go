package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ProxyStore manages proxy persistence in JSON
type ProxyStore struct {
	filePath string
	mu       sync.RWMutex
	accounts map[string]*ProxyAccount // key = ID
	portMap  map[int]*ProxyAccount    // key = Port
}

// NewProxyStore creates or loads a proxy repository
func NewProxyStore(dataFile string) (*ProxyStore, error) {
	if dataFile == "" {
		dataFile = "data/proxies.json"
	}
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	s := &ProxyStore{
		filePath: dataFile,
		accounts: make(map[string]*ProxyAccount),
		portMap:  make(map[int]*ProxyAccount),
	}

	if err := s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *ProxyStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return nil
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

// Save writes accounts to disk safely
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

// GetByID returns an account by ID
func (s *ProxyStore) GetByID(id string) *ProxyAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accounts[id]
}

// GetByPort returns an account by port
func (s *ProxyStore) GetByPort(port int) *ProxyAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.portMap[port]
}

// Add creates or updates an account
func (s *ProxyStore) Add(acc *ProxyAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.portMap[acc.Port]; ok && existing.ID != acc.ID {
		return fmt.Errorf("cổng %d đã được dùng bởi proxy '%s'", acc.Port, existing.Name)
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
		return fmt.Errorf("không tìm thấy proxy: %s", id)
	}

	delete(s.portMap, acc.Port)
	delete(s.accounts, id)
	return nil
}

// NextAvailablePort finds the next free port
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
