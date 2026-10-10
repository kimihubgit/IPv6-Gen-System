package engine

import (
	"net"
	"strings"
	"sync"
	"time"

	"ipv6-gen-linux/internal/store"
)

// SessionIPEntry tracks the rotated IPv6 assigned to a connection/session
type SessionIPEntry struct {
	IP        net.IP
	ExpiresAt time.Time
}

// SessionManager manages sticky sessions for a port listener
type SessionManager struct {
	sessions map[string]*SessionIPEntry
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewSessionManager creates a session manager with background cleanup
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*SessionIPEntry),
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
			sm.mu.Lock()
			for k, v := range sm.sessions {
				if now.After(v.ExpiresAt) {
					delete(sm.sessions, k)
				}
			}
			sm.mu.Unlock()
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
		return PickRandomIPv6(ipNet)

	case store.RotationSticky:
		key := sessionKey
		if key == "" {
			key = "default_port_session"
		}

		sec := acc.StickySec
		if sec <= 0 {
			sec = 10
		}

		sm.mu.RLock()
		entry, exists := sm.sessions[key]
		sm.mu.RUnlock()

		now := time.Now()
		if exists && now.Before(entry.ExpiresAt) {
			return entry.IP
		}

		// Expired or new session -> allocate fresh random IPv6
		newIP := PickRandomIPv6(ipNet)
		sm.mu.Lock()
		sm.sessions[key] = &SessionIPEntry{
			IP:        newIP,
			ExpiresAt: now.Add(time.Duration(sec) * time.Second),
		}
		sm.mu.Unlock()
		return newIP

	default: // RotationPerRequest
		return PickRandomIPv6(ipNet)
	}
}
