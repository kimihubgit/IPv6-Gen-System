package engine

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ipv6-gen-linux/internal/store"
)

// SessionIPEntry tracks the rotated IPv6 assigned to a connection/session
type SessionIPEntry struct {
	IP        net.IP
	ExpiresAt time.Time
}

// SubnetPool maintains a warm revolving pool of pre-generated IPv6 addresses.
// It avoids hammering upstream gateway router ARP/NDP tables during heavy concurrency bursts,
// while guaranteeing that concurrent threads get different IPv6 addresses and load simultaneously.
type SubnetPool struct {
	ips     []net.IP
	counter atomic.Uint64
	ipNet   *net.IPNet
	mu      sync.RWMutex
}

func NewSubnetPool(ipNet *net.IPNet, size int) *SubnetPool {
	if size <= 0 {
		size = 256
	}
	pool := &SubnetPool{
		ips:   make([]net.IP, size),
		ipNet: ipNet,
	}
	for i := 0; i < size; i++ {
		pool.ips[i] = PickRandomIPv6(ipNet)
	}
	return pool
}

func (p *SubnetPool) Next() net.IP {
	if p == nil || len(p.ips) == 0 {
		return PickRandomIPv6(p.ipNet)
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	idx := p.counter.Add(1) % uint64(len(p.ips))
	return p.ips[idx]
}

// RefreshBatch gently replaces a small subset of IPs in the pool to keep them rotating over time
func (p *SubnetPool) RefreshBatch(count int) {
	if p == nil || p.ipNet == nil || len(p.ips) == 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := 0; i < count; i++ {
		var b [2]byte
		_, _ = rand.Read(b[:])
		idx := int(binary.BigEndian.Uint16(b[:])) % len(p.ips)
		p.ips[idx] = PickRandomIPv6(p.ipNet)
	}
}

// SessionManager manages sticky sessions and IPv6 rotation for a port listener
type SessionManager struct {
	sessions map[string]*SessionIPEntry
	pool     *SubnetPool
	ipNet    *net.IPNet
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewSessionManager creates a session manager with warm pool and background cleanup
func NewSessionManager(ipNet *net.IPNet) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*SessionIPEntry),
		pool:     NewSubnetPool(ipNet, 256),
		ipNet:    ipNet,
		stopChan: make(chan struct{}),
	}
	go sm.cleanupLoop()
	return sm
}

// Close stops the cleanup loop
func (sm *SessionManager) Close() {
	select {
	case <-sm.stopChan:
	default:
		close(sm.stopChan)
	}
}

func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case now := <-ticker.C:
			// 1. Clean expired sticky sessions
			sm.mu.Lock()
			for k, v := range sm.sessions {
				if now.After(v.ExpiresAt) {
					delete(sm.sessions, k)
				}
			}
			sm.mu.Unlock()

			// 2. Gently rotate a batch of pool IPs so addresses evolve over time
			if sm.pool != nil {
				sm.pool.RefreshBatch(16)
			}
		}
	}
}

// ExtractSessionKey extracts base username and session tag from incoming credentials
// Format: "username-session-123" -> base = "username", tag = "123"
func ExtractSessionKey(rawUser string) (baseUser string, sessionKey string) {
	if idx := strings.Index(rawUser, "-session-"); idx != -1 {
		baseUser = rawUser[:idx]
		sessionKey = rawUser[idx+len("-session-"):]
		return baseUser, sessionKey
	}
	return rawUser, ""
}

// PickIPv6 selects an outbound IPv6 according to the account's rotation policy and session
func (sm *SessionManager) PickIPv6(sessionKey string, acc *store.ProxyAccount, ipNet *net.IPNet) net.IP {
	switch acc.RotationType {
	case store.RotationStatic:
		if acc.StaticIP != "" {
			if ip := net.ParseIP(acc.StaticIP); ip != nil {
				return ip
			}
		}
		if sm.pool != nil {
			return sm.pool.Next()
		}
		return PickRandomIPv6(ipNet)

	case store.RotationSticky:
		key := sessionKey
		if key == "" {
			key = "default_port_session"
		}

		sec := acc.StickySec
		if sec < 1 {
			sec = 1
		}

		sm.mu.RLock()
		entry, exists := sm.sessions[key]
		sm.mu.RUnlock()

		now := time.Now()
		if exists && now.Before(entry.ExpiresAt) {
			return entry.IP
		}

		// Expired or new session -> allocate fresh IPv6 from pool
		var newIP net.IP
		if sm.pool != nil {
			newIP = sm.pool.Next()
		} else {
			newIP = PickRandomIPv6(ipNet)
		}

		sm.mu.Lock()
		sm.sessions[key] = &SessionIPEntry{
			IP:        newIP,
			ExpiresAt: now.Add(time.Duration(sec) * time.Second),
		}
		sm.mu.Unlock()
		return newIP

	default: // RotationPerRequest
		if sm.pool != nil {
			return sm.pool.Next()
		}
		return PickRandomIPv6(ipNet)
	}
}
