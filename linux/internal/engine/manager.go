package engine

import (
	"log"
	"net"
	"sync"
	"time"

	"ipv6-gen-linux/internal/network"
	"ipv6-gen-linux/internal/store"
)

// MultiProxyManager manages dynamic port listeners for all proxy accounts
type MultiProxyManager struct {
	ipNet     *net.IPNet
	ifaceName string
	store     *store.ProxyStore
	listeners map[int]*PortListener // key = Port
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// NewMultiProxyManager creates a manager for multi-port proxy accounts
func NewMultiProxyManager(ifaceName string, ipNet *net.IPNet, store *store.ProxyStore) *MultiProxyManager {
	return &MultiProxyManager{
		ipNet:     ipNet,
		ifaceName: ifaceName,
		store:     store,
		listeners: make(map[int]*PortListener),
		stopChan:  make(chan struct{}),
	}
}

// Start launches listeners for all enabled accounts and sets up kernel Any-IP routing
func (m *MultiProxyManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Setup Any-IP Kernel routing
	_ = network.SetupAnyIP(m.ifaceName, m.ipNet)

	// 2. Start listeners
	for _, acc := range m.store.GetAll() {
		if err := m.startListenerLocked(acc); err != nil {
			log.Printf("⚠️ Không thể mở port %d cho proxy '%s': %v", acc.Port, acc.Name, err)
		}
	}

	// 3. Background persistence and quota checker
	go m.backgroundLoop()

	return nil
}

// Stop shuts down all listeners and tears down local route
func (m *MultiProxyManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.stopChan:
		return
	default:
		close(m.stopChan)
		for port, pl := range m.listeners {
			pl.Close()
			delete(m.listeners, port)
		}
		network.TeardownAnyIP(m.ipNet)
		log.Println("🛑 Đã dừng toàn bộ MultiProxyManager và giải phóng cổng.")
	}
}

func (m *MultiProxyManager) startListenerLocked(acc *store.ProxyAccount) error {
	if !acc.Enabled {
		return nil
	}

	if _, exists := m.listeners[acc.Port]; exists {
		return nil
	}

	pl, err := NewPortListener(acc, m.ipNet)
	if err != nil {
		return err
	}

	m.listeners[acc.Port] = pl
	log.Printf("✅ Đã kích hoạt proxy [%s] lắng nghe cổng %d (HTTP + SOCKS5)", acc.Name, acc.Port)
	return nil
}

// SyncAccount starts or stops a listener when an account is updated
func (m *MultiProxyManager) SyncAccount(acc *store.ProxyAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pl, exists := m.listeners[acc.Port]; exists {
		if !acc.Enabled {
			pl.Close()
			delete(m.listeners, acc.Port)
			log.Printf("⏸️ Đã tạm dừng cổng %d", acc.Port)
			return nil
		}
		// Update account reference
		pl.account = acc
		return nil
	}

	if acc.Enabled {
		return m.startListenerLocked(acc)
	}

	return nil
}

// RemoveAccount stops and unbinds a port
func (m *MultiProxyManager) RemoveAccount(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pl, exists := m.listeners[port]; exists {
		pl.Close()
		delete(m.listeners, port)
		log.Printf("🗑️ Đã giải phóng cổng %d", port)
	}
}

// ActivePorts returns the list of currently active listening ports
func (m *MultiProxyManager) ActivePorts() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ports []int
	for p := range m.listeners {
		ports = append(ports, p)
	}
	return ports
}

func (m *MultiProxyManager) backgroundLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			// Auto save store
			_ = m.store.Save()

			// Check expired accounts
			now := time.Now()
			for _, acc := range m.store.GetAll() {
				if acc.Enabled {
					if acc.ExpiresAt != nil && acc.ExpiresAt.Before(now) {
						m.RemoveAccount(acc.Port)
						log.Printf("⌛ Proxy '%s' (Port %d) đã hết hạn", acc.Name, acc.Port)
					}
					if acc.MaxBytes > 0 && acc.BytesUsed >= acc.MaxBytes {
						m.RemoveAccount(acc.Port)
						log.Printf("🚫 Proxy '%s' (Port %d) đã đạt hạn mức dung lượng", acc.Name, acc.Port)
					}
				}
			}
		}
	}
}
