package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WebServer manages the Web Admin Dashboard and REST APIs
type WebServer struct {
	port       int
	prefix     string
	publicIPv4 string
	manager    *MultiProxyManager
	store      *ProxyStore
	adminUser  string
	adminPass  string
	sessions   map[string]time.Time
	mu         sync.RWMutex
	startTime  time.Time
}

// NewWebServer creates a new Web Dashboard server
func NewWebServer(port int, prefix string, publicIPv4 string, manager *MultiProxyManager, store *ProxyStore) *WebServer {
	if port <= 0 {
		port = 9090
	}
	return &WebServer{
		port:       port,
		prefix:     prefix,
		publicIPv4: publicIPv4,
		manager:    manager,
		store:      store,
		adminUser:  "admin",
		adminPass:  "admin123", // Mật khẩu mặc định, có thể đổi trong giao diện
		sessions:   make(map[string]time.Time),
		startTime:  time.Now(),
	}
}

// Start launches the Web Dashboard HTTP listener
func (ws *WebServer) Start() error {
	mux := http.NewServeMux()

	// Static & App routes
	mux.HandleFunc("/", ws.handleIndex)

	// API routes
	mux.HandleFunc("/api/login", ws.handleLogin)
	mux.HandleFunc("/api/logout", ws.handleLogout)
	mux.HandleFunc("/api/stats", ws.authMiddleware(ws.handleStats))
	mux.HandleFunc("/api/proxies", ws.authMiddleware(ws.handleProxies))
	mux.HandleFunc("/api/proxies/", ws.authMiddleware(ws.handleProxyItem))
	mux.HandleFunc("/api/export", ws.authMiddleware(ws.handleExport))
	mux.HandleFunc("/api/settings", ws.authMiddleware(ws.handleSettings))

	addr := fmt.Sprintf("0.0.0.0:%d", ws.port)
	log.Printf("🚀 Web Dashboard đã sẵn sàng tại: http://%s:%d (Đăng nhập: %s / %s)",
		ws.publicIPv4, ws.port, ws.adminUser, ws.adminPass)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return server.ListenAndServe()
}

// Session helpers
func (ws *WebServer) generateSession() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	ws.mu.Lock()
	ws.sessions[token] = time.Now().Add(24 * time.Hour)
	ws.mu.Unlock()
	return token
}

func (ws *WebServer) isValidSession(r *http.Request) bool {
	cookie, err := r.Cookie("admin_token")
	if err != nil || cookie.Value == "" {
		return false
	}

	ws.mu.RLock()
	exp, exists := ws.sessions[cookie.Value]
	ws.mu.RUnlock()

	if !exists || time.Now().After(exp) {
		return false
	}
	return true
}

func (ws *WebServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ws.isValidSession(r) {
			http.Error(w, `{"error":"Chưa đăng nhập hoặc phiên làm việc đã hết hạn"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// ----------------------------------------------------------------------
// API HANDLERS
// ----------------------------------------------------------------------

func (ws *WebServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu không hợp lệ"})
		return
	}

	ws.mu.RLock()
	match := req.Username == ws.adminUser && req.Password == ws.adminPass
	ws.mu.RUnlock()

	if !match {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Tài khoản hoặc mật khẩu không chính xác!"})
		return
	}

	token := ws.generateSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Đăng nhập thành công"})
}

func (ws *WebServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_token")
	if err == nil && cookie.Value != "" {
		ws.mu.Lock()
		delete(ws.sessions, cookie.Value)
		ws.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Đã đăng xuất"})
}

func (ws *WebServer) handleStats(w http.ResponseWriter, r *http.Request) {
	proxies := ws.store.GetAll()
	activeCount := 0
	var totalBytes int64 = 0

	now := time.Now()
	for _, p := range proxies {
		totalBytes += p.BytesUsed
		if p.Enabled {
			if p.ExpiresAt != nil && p.ExpiresAt.Before(now) {
				continue
			}
			if p.MaxBytes > 0 && p.BytesUsed >= p.MaxBytes {
				continue
			}
			activeCount++
		}
	}

	uptimeDuration := time.Since(ws.startTime).Round(time.Second)
	days := int(uptimeDuration.Hours()) / 24
	hours := int(uptimeDuration.Hours()) % 24
	mins := int(uptimeDuration.Minutes()) % 60
	secs := int(uptimeDuration.Seconds()) % 60
	var uptimeStr string
	if days > 0 {
		uptimeStr = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	} else if hours > 0 {
		uptimeStr = fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	} else {
		uptimeStr = fmt.Sprintf("%dm %ds", mins, secs)
	}

	// Suggest next free port
	suggestPort := 10001
	usedPorts := make(map[int]bool)
	for _, p := range proxies {
		usedPorts[p.Port] = true
	}
	for usedPorts[suggestPort] {
		suggestPort++
	}

	res := map[string]interface{}{
		"active_proxies": activeCount,
		"total_proxies":  len(proxies),
		"total_bytes":    totalBytes,
		"uptime":         uptimeStr,
		"prefix":         ws.prefix,
		"public_ip":      ws.publicIPv4,
		"suggest_port":   suggestPort,
	}

	jsonResponse(w, http.StatusOK, res)
}

func (ws *WebServer) handleProxies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		proxies := ws.store.GetAll()
		jsonResponse(w, http.StatusOK, proxies)

	case http.MethodPost:
		var req struct {
			Name         string  `json:"name"`
			Port         int     `json:"port"`
			Username     string  `json:"username"`
			Password     string  `json:"password"`
			MaxGB        float64 `json:"max_gb"`
			ExpireDays   int     `json:"expire_days"`
			RotationType string  `json:"rotation_type"`
			StickySec    int     `json:"sticky_sec"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu JSON không hợp lệ"})
			return
		}

		if req.Port < 1024 || req.Port > 65535 {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Cổng phải nằm trong khoảng 1024 - 65535"})
			return
		}

		if req.Name == "" {
			req.Name = fmt.Sprintf("Proxy Port %d", req.Port)
		}

		var maxBytes int64 = 0
		if req.MaxGB > 0 {
			maxBytes = int64(req.MaxGB * 1024 * 1024 * 1024)
		}

		var expiresAt *time.Time
		if req.ExpireDays > 0 {
			exp := time.Now().AddDate(0, 0, req.ExpireDays)
			expiresAt = &exp
		}

		rotType := RotationPolicy(req.RotationType)
		if rotType != RotationPerRequest && rotType != RotationSticky && rotType != RotationStatic {
			rotType = RotationPerRequest
		}

		stickySec := req.StickySec
		if stickySec <= 0 {
			stickySec = 10
		}

		acc := &ProxyAccount{
			ID:           fmt.Sprintf("px-%d", time.Now().UnixNano()),
			Name:         req.Name,
			Port:         req.Port,
			Username:     req.Username,
			Password:     req.Password,
			Proto:        "both",
			MaxBytes:     maxBytes,
			BytesUsed:    0,
			ExpiresAt:    expiresAt,
			RotationType: rotType,
			StickySec:    stickySec,
			Enabled:      true,
			CreatedAt:    time.Now(),
		}

		if err := ws.store.Add(acc); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := ws.manager.SyncAccount(acc); err != nil {
			_ = ws.store.Delete(acc.ID)
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Không thể mở cổng %d: %v", req.Port, err)})
			return
		}

		_ = ws.store.Save()
		jsonResponse(w, http.StatusCreated, acc)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ws *WebServer) handleProxyItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/proxies/")
	parts := strings.Split(path, "/")
	id := parts[0]

	acc := ws.store.GetByID(id)
	if acc == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Không tìm thấy proxy"})
		return
	}

	if len(parts) > 1 && parts[1] == "reset" && r.Method == http.MethodPost {
		acc.BytesUsed = 0
		_ = ws.store.Save()
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Đã reset dung lượng về 0 GB"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		jsonResponse(w, http.StatusOK, acc)

	case http.MethodPut:
		var req struct {
			Name         string   `json:"name"`
			Port         int      `json:"port"`
			Username     string   `json:"username"`
			Password     string   `json:"password"`
			MaxGB        *float64 `json:"max_gb"`
			ExpireDays   *int     `json:"expire_days"`
			AddDays      int      `json:"add_days"`
			SetExpireNil bool     `json:"set_expire_nil"`
			RotationType string   `json:"rotation_type"`
			StickySec    int      `json:"sticky_sec"`
			Enabled      *bool    `json:"enabled"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu JSON không hợp lệ"})
			return
		}

		if req.Name != "" {
			acc.Name = req.Name
		}
		if req.Username != "" {
			acc.Username = req.Username
		}
		if req.Password != "" {
			acc.Password = req.Password
		}
		if req.MaxGB != nil {
			if *req.MaxGB <= 0 {
				acc.MaxBytes = 0
			} else {
				acc.MaxBytes = int64(*req.MaxGB * 1024 * 1024 * 1024)
			}
		}

		if req.SetExpireNil {
			acc.ExpiresAt = nil
		} else if req.ExpireDays != nil {
			if *req.ExpireDays <= 0 {
				acc.ExpiresAt = nil
			} else {
				exp := time.Now().AddDate(0, 0, *req.ExpireDays)
				acc.ExpiresAt = &exp
			}
		} else if req.AddDays > 0 {
			base := time.Now()
			if acc.ExpiresAt != nil && acc.ExpiresAt.After(base) {
				base = *acc.ExpiresAt
			}
			exp := base.AddDate(0, 0, req.AddDays)
			acc.ExpiresAt = &exp
		}

		if req.RotationType != "" {
			acc.RotationType = RotationPolicy(req.RotationType)
		}
		if req.StickySec > 0 {
			acc.StickySec = req.StickySec
		}
		if req.Enabled != nil {
			acc.Enabled = *req.Enabled
		}

		// Handle port update if changed
		if req.Port > 0 && req.Port != acc.Port {
			if req.Port < 1024 || req.Port > 65535 {
				jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Cổng phải nằm trong khoảng 1024 - 65535"})
				return
			}
			if existing := ws.store.GetByPort(req.Port); existing != nil && existing.ID != acc.ID {
				jsonResponse(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Cổng %d đã được dùng bởi proxy khác", req.Port)})
				return
			}
			oldPort := acc.Port
			acc.Port = req.Port
			ws.manager.RemoveAccount(oldPort)
		}

		_ = ws.manager.SyncAccount(acc)
		_ = ws.store.Save()
		jsonResponse(w, http.StatusOK, acc)

	case http.MethodDelete:
		ws.manager.RemoveAccount(acc.Port)
		_ = ws.store.Delete(acc.ID)
		_ = ws.store.Save()
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Đã xóa proxy"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ws *WebServer) handleExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format") // "ip_port_user_pass" | "json"
	proxies := ws.store.GetAll()

	if format == "json" {
		jsonResponse(w, http.StatusOK, proxies)
		return
	}

	var sb strings.Builder
	for _, p := range proxies {
		if p.Enabled {
			if p.Username != "" && p.Password != "" {
				sb.WriteString(fmt.Sprintf("%s:%d:%s:%s\n", ws.publicIPv4, p.Port, p.Username, p.Password))
			} else {
				sb.WriteString(fmt.Sprintf("%s:%d\n", ws.publicIPv4, p.Port))
			}
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(sb.String()))
}

func (ws *WebServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewUsername     string `json:"new_username"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu không hợp lệ"})
		return
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if req.CurrentPassword != ws.adminPass {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Mật khẩu hiện tại không đúng"})
		return
	}

	if req.NewUsername != "" {
		ws.adminUser = req.NewUsername
	}
	if req.NewPassword != "" {
		ws.adminPass = req.NewPassword
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Đã đổi thông tin đăng nhập thành công"})
}

// ----------------------------------------------------------------------
// EMBEDDED DASHBOARD SINGLE-PAGE APPLICATION (HTML / CSS / JS)
// ----------------------------------------------------------------------

func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="vi">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>IPv6 Proxy Hub - Cloud Enterprise Manager</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-dark: #080c14;
      --bg-surface: rgba(15, 23, 42, 0.72);
      --bg-surface-elevated: rgba(22, 34, 60, 0.85);
      --border-subtle: rgba(255, 255, 255, 0.08);
      --border-focus: rgba(99, 102, 241, 0.6);
      
      --primary: #6366f1;
      --primary-hover: #4f46e5;
      --primary-glow: rgba(99, 102, 241, 0.35);
      
      --cyan: #06b6d4;
      --cyan-glow: rgba(6, 182, 212, 0.3);
      
      --emerald: #10b981;
      --emerald-glow: rgba(16, 185, 129, 0.3);
      
      --amber: #f59e0b;
      --amber-glow: rgba(245, 158, 11, 0.3);
      
      --rose: #f43f5e;
      --rose-glow: rgba(244, 63, 94, 0.3);
      
      --text: #f8fafc;
      --text-muted: #94a3b8;
      --text-dim: #64748b;
      
      --radius-sm: 8px;
      --radius-md: 12px;
      --radius-lg: 18px;
      --radius-xl: 24px;
    }

    * { box-sizing: border-box; margin: 0; padding: 0; }
    
    body {
      font-family: 'Plus Jakarta Sans', sans-serif;
      background: var(--bg-dark);
      color: var(--text);
      min-height: 100vh;
      line-height: 1.5;
      background-image: 
        radial-gradient(circle at 10% 10%, rgba(99, 102, 241, 0.15) 0px, transparent 45%),
        radial-gradient(circle at 90% 20%, rgba(6, 182, 212, 0.12) 0px, transparent 40%),
        radial-gradient(circle at 50% 90%, rgba(139, 92, 246, 0.1) 0px, transparent 50%);
      background-attachment: fixed;
      overflow-x: hidden;
    }

    /* Custom Scrollbar */
    ::-webkit-scrollbar { width: 8px; height: 8px; }
    ::-webkit-scrollbar-track { background: rgba(0, 0, 0, 0.2); }
    ::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.15); border-radius: 4px; }
    ::-webkit-scrollbar-thumb:hover { background: rgba(255, 255, 255, 0.25); }

    .app-container {
      max-width: 1440px;
      margin: 0 auto;
      padding: 24px 32px 48px;
    }

    /* Glass Effect Utility */
    .glass-card {
      background: var(--bg-surface);
      backdrop-filter: blur(20px) saturate(180%);
      -webkit-backdrop-filter: blur(20px) saturate(180%);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg);
      box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.5);
    }

    /* Header */
    header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 24px;
      margin-bottom: 28px;
      border-radius: var(--radius-xl);
    }

    .brand-section {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .brand-logo {
      width: 48px;
      height: 48px;
      border-radius: 14px;
      background: linear-gradient(135deg, #6366f1, #06b6d4);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 0 24px var(--primary-glow);
      position: relative;
    }

    .brand-logo svg {
      width: 26px;
      height: 26px;
      fill: none;
      stroke: #fff;
      stroke-width: 2;
    }

    .brand-title h1 {
      font-size: 20px;
      font-weight: 800;
      letter-spacing: -0.5px;
      background: linear-gradient(120deg, #ffffff, #cbd5e1);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }

    .brand-title p {
      font-size: 13px;
      color: var(--text-muted);
      display: flex;
      align-items: center;
      gap: 6px;
    }

    .live-beacon {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 3px 10px;
      border-radius: 20px;
      background: rgba(16, 185, 129, 0.12);
      border: 1px solid rgba(16, 185, 129, 0.3);
      color: #34d399;
      font-size: 11px;
      font-weight: 600;
      letter-spacing: 0.5px;
      text-transform: uppercase;
    }

    .pulse-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: #10b981;
      box-shadow: 0 0 10px #10b981;
      animation: pulseAnim 2s infinite;
    }

    @keyframes pulseAnim {
      0%, 100% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.4; transform: scale(0.85); }
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .ip-badge {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 14px;
      border-radius: var(--radius-md);
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid var(--border-subtle);
      font-family: 'JetBrains Mono', monospace;
      font-size: 13px;
      color: #e2e8f0;
      cursor: pointer;
      transition: all 0.2s;
    }

    .ip-badge:hover {
      background: rgba(255, 255, 255, 0.08);
      border-color: rgba(255, 255, 255, 0.15);
    }

    /* Buttons */
    .btn {
      padding: 9px 16px;
      border-radius: var(--radius-md);
      font-weight: 600;
      font-size: 13.5px;
      cursor: pointer;
      transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
      display: inline-flex;
      align-items: center;
      gap: 8px;
      border: 1px solid transparent;
      outline: none;
      text-decoration: none;
      user-select: none;
    }

    .btn:active { transform: scale(0.97); }

    .btn svg {
      width: 16px;
      height: 16px;
      stroke-width: 2;
    }

    .btn-primary {
      background: linear-gradient(135deg, #6366f1, #4f46e5);
      color: #fff;
      box-shadow: 0 4px 18px var(--primary-glow);
      border-color: rgba(255, 255, 255, 0.15);
    }

    .btn-primary:hover {
      background: linear-gradient(135deg, #4f46e5, #4338ca);
      box-shadow: 0 6px 24px var(--primary-glow);
      transform: translateY(-1px);
    }

    .btn-secondary {
      background: rgba(255, 255, 255, 0.05);
      color: #e2e8f0;
      border-color: var(--border-subtle);
    }

    .btn-secondary:hover {
      background: rgba(255, 255, 255, 0.09);
      border-color: rgba(255, 255, 255, 0.15);
      transform: translateY(-1px);
    }

    .btn-danger {
      background: rgba(244, 63, 94, 0.12);
      color: #fda4af;
      border-color: rgba(244, 63, 94, 0.25);
    }

    .btn-danger:hover {
      background: rgba(244, 63, 94, 0.22);
      border-color: rgba(244, 63, 94, 0.4);
    }

    .btn-sm {
      padding: 6px 12px;
      font-size: 12.5px;
      border-radius: var(--radius-sm);
    }

    .btn-icon-only {
      padding: 8px;
      border-radius: var(--radius-sm);
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }

    /* KPI Stats Grid */
    .kpi-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 20px;
      margin-bottom: 28px;
    }

    @media (max-width: 1100px) {
      .kpi-grid { grid-template-columns: repeat(2, 1fr); }
    }
    @media (max-width: 640px) {
      .kpi-grid { grid-template-columns: 1fr; }
    }

    .kpi-card {
      padding: 22px 24px;
      display: flex;
      flex-direction: column;
      position: relative;
      overflow: hidden;
      transition: all 0.3s;
    }

    .kpi-card:hover {
      transform: translateY(-2px);
      border-color: rgba(255, 255, 255, 0.15);
      box-shadow: 0 14px 34px -10px rgba(0, 0, 0, 0.6);
    }

    .kpi-card::before {
      content: '';
      position: absolute;
      top: 0; left: 0; right: 0;
      height: 2px;
      background: var(--kpi-accent, var(--primary));
      opacity: 0.8;
    }

    .kpi-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
    }

    .kpi-title {
      font-size: 13px;
      font-weight: 600;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .kpi-icon-wrap {
      width: 40px;
      height: 40px;
      border-radius: 12px;
      background: var(--kpi-bg, rgba(99, 102, 241, 0.12));
      border: 1px solid var(--kpi-border, rgba(99, 102, 241, 0.25));
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--kpi-color, var(--primary));
    }

    .kpi-icon-wrap svg {
      width: 20px;
      height: 20px;
      stroke-width: 2;
    }

    .kpi-value {
      font-size: 28px;
      font-weight: 800;
      letter-spacing: -0.5px;
      color: #fff;
      margin-bottom: 4px;
      line-height: 1.2;
    }

    .kpi-desc {
      font-size: 12.5px;
      color: var(--text-dim);
    }

    /* Filter & Controls Toolbar */
    .controls-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 16px;
      margin-bottom: 18px;
      flex-wrap: wrap;
    }

    .tabs-group {
      display: flex;
      background: rgba(0, 0, 0, 0.3);
      padding: 4px;
      border-radius: var(--radius-md);
      border: 1px solid var(--border-subtle);
    }

    .tab-btn {
      padding: 7px 16px;
      border-radius: 8px;
      font-size: 13px;
      font-weight: 600;
      background: transparent;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      transition: all 0.2s;
    }

    .tab-btn.active {
      background: rgba(255, 255, 255, 0.1);
      color: #fff;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    }

    .search-input-wrap {
      position: relative;
      min-width: 300px;
    }

    .search-input-wrap svg {
      position: absolute;
      left: 14px;
      top: 50%;
      transform: translateY(-50%);
      width: 16px;
      height: 16px;
      color: var(--text-dim);
      pointer-events: none;
    }

    .search-input {
      width: 100%;
      padding: 10px 16px 10px 40px;
      border-radius: var(--radius-md);
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid var(--border-subtle);
      color: #fff;
      font-size: 13.5px;
      outline: none;
      transition: all 0.2s;
    }

    .search-input:focus {
      background: rgba(255, 255, 255, 0.07);
      border-color: var(--primary);
      box-shadow: 0 0 0 3px var(--primary-glow);
    }

    /* Table Design */
    .table-container {
      overflow: hidden;
      border-radius: var(--radius-lg);
    }

    .table-responsive {
      overflow-x: auto;
    }

    table {
      width: 100%;
      border-collapse: separate;
      border-spacing: 0;
      text-align: left;
    }

    th {
      background: rgba(255, 255, 255, 0.025);
      padding: 16px 20px;
      font-size: 11.5px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.7px;
      color: var(--text-dim);
      border-bottom: 1px solid var(--border-subtle);
      white-space: nowrap;
    }

    td {
      padding: 16px 20px;
      font-size: 13.5px;
      border-bottom: 1px solid rgba(255, 255, 255, 0.04);
      vertical-align: middle;
      color: #e2e8f0;
      transition: background 0.15s;
    }

    tr:last-child td { border-bottom: none; }
    tbody tr:hover td { background: rgba(255, 255, 255, 0.025); }

    .client-title {
      font-weight: 700;
      color: #fff;
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .client-avatar {
      width: 32px;
      height: 32px;
      border-radius: 9px;
      background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(6, 182, 212, 0.2));
      border: 1px solid rgba(255, 255, 255, 0.1);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      color: #cbd5e1;
    }

    .port-pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-family: 'JetBrains Mono', monospace;
      font-weight: 600;
      color: #38bdf8;
      background: rgba(56, 189, 248, 0.1);
      border: 1px solid rgba(56, 189, 248, 0.2);
      padding: 4px 10px;
      border-radius: 8px;
      font-size: 13px;
    }

    .auth-snippet {
      font-family: 'JetBrains Mono', monospace;
      font-size: 12.5px;
      background: rgba(0, 0, 0, 0.35);
      border: 1px solid rgba(255, 255, 255, 0.08);
      padding: 4px 10px;
      border-radius: 7px;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      color: #cbd5e1;
      transition: all 0.2s;
    }

    .auth-snippet:hover {
      background: rgba(255, 255, 255, 0.08);
      border-color: rgba(255, 255, 255, 0.18);
      color: #fff;
    }

    /* Progress Traffic */
    .traffic-box {
      min-width: 130px;
    }

    .traffic-label {
      font-size: 12.5px;
      font-weight: 600;
      color: #cbd5e1;
      display: flex;
      justify-content: space-between;
      margin-bottom: 5px;
    }

    .progress-track {
      height: 6px;
      background: rgba(255, 255, 255, 0.08);
      border-radius: 4px;
      overflow: hidden;
      position: relative;
    }

    .progress-fill {
      height: 100%;
      border-radius: 4px;
      background: linear-gradient(90deg, #6366f1, #06b6d4);
      transition: width 0.4s ease;
    }

    .progress-fill.full {
      background: linear-gradient(90deg, #f59e0b, #ef4444);
    }

    /* Badges */
    .badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 10px;
      border-radius: 20px;
      font-size: 12px;
      font-weight: 600;
      letter-spacing: 0.3px;
    }

    .badge-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    .badge-success {
      background: rgba(16, 185, 129, 0.12);
      color: #34d399;
      border: 1px solid rgba(16, 185, 129, 0.25);
    }
    .badge-success .badge-dot { background: #10b981; box-shadow: 0 0 6px #10b981; }

    .badge-danger {
      background: rgba(244, 63, 94, 0.12);
      color: #fda4af;
      border: 1px solid rgba(244, 63, 94, 0.25);
    }
    .badge-danger .badge-dot { background: #f43f5e; box-shadow: 0 0 6px #f43f5e; }

    .badge-warning {
      background: rgba(245, 158, 11, 0.12);
      color: #fcd34d;
      border: 1px solid rgba(245, 158, 11, 0.25);
    }
    .badge-warning .badge-dot { background: #f59e0b; box-shadow: 0 0 6px #f59e0b; }

    .badge-neutral {
      background: rgba(255, 255, 255, 0.06);
      color: var(--text-muted);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
    .badge-neutral .badge-dot { background: var(--text-dim); }

    .badge-indigo {
      background: rgba(99, 102, 241, 0.12);
      color: #a5b4fc;
      border: 1px solid rgba(99, 102, 241, 0.25);
    }

    /* Actions Table Group */
    .actions-group {
      display: flex;
      align-items: center;
      justify-content: flex-end;
      gap: 6px;
    }

    /* Modals */
    .modal-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(3, 7, 18, 0.75);
      backdrop-filter: blur(12px);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1000;
      padding: 20px;
    }

    .modal-box {
      background: #0f172a;
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-xl);
      width: 100%;
      max-width: 540px;
      box-shadow: 0 25px 60px -15px rgba(0, 0, 0, 0.8);
      overflow: hidden;
      animation: modalPop 0.25s cubic-bezier(0.16, 1, 0.3, 1);
    }

    @keyframes modalPop {
      from { transform: scale(0.95); opacity: 0; }
      to { transform: scale(1); opacity: 1; }
    }

    .modal-header {
      padding: 20px 24px;
      border-bottom: 1px solid var(--border-subtle);
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: rgba(255, 255, 255, 0.015);
    }

    .modal-header h3 {
      font-size: 17px;
      font-weight: 700;
      color: #fff;
    }

    .modal-body {
      padding: 24px;
      display: flex;
      flex-direction: column;
      gap: 16px;
      max-height: 80vh;
      overflow-y: auto;
    }

    .modal-footer {
      padding: 16px 24px;
      border-top: 1px solid var(--border-subtle);
      display: flex;
      justify-content: flex-end;
      gap: 10px;
      background: rgba(0, 0, 0, 0.2);
    }

    .form-group {
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .form-label {
      font-size: 13px;
      font-weight: 600;
      color: #cbd5e1;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .form-control {
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-md);
      padding: 10px 14px;
      color: #fff;
      font-size: 13.5px;
      outline: none;
      transition: all 0.2s;
      width: 100%;
    }

    .form-control:focus {
      background: rgba(255, 255, 255, 0.07);
      border-color: var(--primary);
      box-shadow: 0 0 0 3px var(--primary-glow);
    }

    .input-with-action {
      display: flex;
      gap: 8px;
    }

    .preview-proxy-card {
      background: rgba(0, 0, 0, 0.35);
      border: 1px dashed rgba(255, 255, 255, 0.15);
      border-radius: var(--radius-md);
      padding: 12px 14px;
      font-family: 'JetBrains Mono', monospace;
      font-size: 12.5px;
      color: #38bdf8;
      display: flex;
      justify-content: space-between;
      align-items: center;
      word-break: break-all;
    }

    /* Toast Notification */
    .toast-container {
      position: fixed;
      bottom: 24px;
      right: 24px;
      z-index: 2000;
      display: flex;
      flex-direction: column;
      gap: 10px;
      pointer-events: none;
    }

    .toast-item {
      pointer-events: auto;
      background: rgba(15, 23, 42, 0.95);
      border: 1px solid rgba(255, 255, 255, 0.12);
      backdrop-filter: blur(16px);
      padding: 12px 18px;
      border-radius: var(--radius-md);
      box-shadow: 0 12px 30px rgba(0, 0, 0, 0.6);
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 13.5px;
      font-weight: 500;
      color: #f8fafc;
      animation: toastIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    }

    @keyframes toastIn {
      from { transform: translateY(20px); opacity: 0; }
      to { transform: translateY(0); opacity: 1; }
    }

    /* Login Screen */
    .login-wrapper {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 20px;
    }

    .login-box {
      width: 100%;
      max-width: 420px;
      padding: 40px 36px;
      border-radius: var(--radius-xl);
      text-align: center;
    }

    .login-icon {
      width: 60px;
      height: 60px;
      margin: 0 auto 20px;
      border-radius: 18px;
      background: linear-gradient(135deg, #6366f1, #06b6d4);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 0 30px var(--primary-glow);
    }

    .login-icon svg {
      width: 32px;
      height: 32px;
      stroke: #fff;
    }
  </style>
</head>
<body>

  <!-- LOGIN SECTION -->
  <div id="loginSection" class="login-wrapper" style="display: none;">
    <div class="glass-card login-box">
      <div class="login-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
      </div>
      <h2 style="font-size: 24px; font-weight: 800; margin-bottom: 6px;">IPv6 Proxy Hub</h2>
      <p style="font-size: 13.5px; color: var(--text-muted); margin-bottom: 28px;">Đăng nhập bảng điều khiển hệ thống ANY-IP</p>
      
      <form id="loginForm" onsubmit="doLogin(event)" style="text-align: left;">
        <div class="form-group" style="margin-bottom: 16px;">
          <label class="form-label">Tài Khoản Quản Trị</label>
          <input type="text" id="loginUser" class="form-control" required placeholder="admin" autofocus>
        </div>
        <div class="form-group" style="margin-bottom: 24px;">
          <label class="form-label">Mật Khẩu</label>
          <input type="password" id="loginPass" class="form-control" required placeholder="••••••••">
        </div>
        <button type="submit" class="btn btn-primary" style="width: 100%; justify-content: center; padding: 12px; font-size: 14.5px;">
          Xác Nhận Đăng Nhập
        </button>
      </form>
    </div>
  </div>

  <!-- MAIN DASHBOARD -->
  <div id="appSection" class="app-container" style="display: none;">
    
    <!-- Top Bar -->
    <header class="glass-card">
      <div class="brand-section">
        <div class="brand-logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
        </div>
        <div class="brand-title">
          <div style="display: flex; align-items: center; gap: 10px;">
            <h1>IPV6 PROXY HUB</h1>
            <span class="live-beacon"><span class="pulse-dot"></span> LIVE ANY-IP</span>
          </div>
          <p>Kernel Rotator • 18,446,744,073,709,551,616 Địa Chỉ IPv6</p>
        </div>
      </div>

      <div class="header-actions">
        <div class="ip-badge" onclick="copyToClipboard(publicIP)" title="Bấm để copy IP VPS">
          <svg style="width: 15px; height: 15px; stroke: #38bdf8;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/><path d="m14 9 3 3-3 3"/></svg>
          <span id="topPublicIP">--</span>
        </div>
        <button class="btn btn-secondary btn-sm" onclick="openExportModal()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
          Xuất Danh Sách
        </button>
        <button class="btn btn-secondary btn-sm" onclick="openSettingsModal()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
          Cài Đặt
        </button>
        <button class="btn btn-danger btn-sm" onclick="doLogout()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
          Thoát
        </button>
      </div>
    </header>

    <!-- KPI Metrics Cards -->
    <div class="kpi-grid">
      <!-- Card 1 -->
      <div class="glass-card kpi-card" style="--kpi-accent: #10b981; --kpi-bg: rgba(16, 185, 129, 0.12); --kpi-border: rgba(16, 185, 129, 0.25); --kpi-color: #34d399;">
        <div class="kpi-header">
          <span class="kpi-title">Proxy Đang Hoạt Động</span>
          <div class="kpi-icon-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect width="20" height="8" x="2" y="2" rx="2" ry="2"/><rect width="20" height="8" x="2" y="14" rx="2" ry="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>
          </div>
        </div>
        <div class="kpi-value" id="kpiActive">0 / 0</div>
        <div class="kpi-desc">Cổng proxy sẵn sàng phục vụ</div>
      </div>

      <!-- Card 2 -->
      <div class="glass-card kpi-card" style="--kpi-accent: #06b6d4; --kpi-bg: rgba(6, 182, 212, 0.12); --kpi-border: rgba(6, 182, 212, 0.25); --kpi-color: #38bdf8;">
        <div class="kpi-header">
          <span class="kpi-title">Lưu Lượng Tiêu Thụ</span>
          <div class="kpi-icon-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M12 2v20"/><path d="m17 5-5-3-5 3"/><path d="m17 19-5 3-5-3"/></svg>
          </div>
        </div>
        <div class="kpi-value" id="kpiTraffic">0.00 GB</div>
        <div class="kpi-desc">Tổng băng thông tải lên & tải về</div>
      </div>

      <!-- Card 3 -->
      <div class="glass-card kpi-card" style="--kpi-accent: #8b5cf6; --kpi-bg: rgba(139, 92, 246, 0.12); --kpi-border: rgba(139, 92, 246, 0.25); --kpi-color: #a78bfa;">
        <div class="kpi-header">
          <span class="kpi-title">Dải Mạng IPv6 Subnet</span>
          <div class="kpi-icon-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"/><line x1="2" x2="22" y1="12" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
          </div>
        </div>
        <div class="kpi-value" style="font-size: 19px; font-family: 'JetBrains Mono', monospace;" id="kpiPrefix">--</div>
        <div class="kpi-desc">Dải mạng Any-IP (Không gán eth0)</div>
      </div>

      <!-- Card 4 -->
      <div class="glass-card kpi-card" style="--kpi-accent: #f59e0b; --kpi-bg: rgba(245, 158, 11, 0.12); --kpi-border: rgba(245, 158, 11, 0.25); --kpi-color: #fbbf24;">
        <div class="kpi-header">
          <span class="kpi-title">Thời Gian Hoạt Động</span>
          <div class="kpi-icon-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
        </div>
        <div class="kpi-value" id="kpiUptime">--</div>
        <div class="kpi-desc" id="kpiHost">VPS: --</div>
      </div>
    </div>

    <!-- Controls Toolbar -->
    <div class="controls-toolbar">
      <div style="display: flex; gap: 14px; align-items: center; flex-wrap: wrap;">
        <!-- Status Tabs -->
        <div class="tabs-group">
          <button class="tab-btn active" onclick="setFilter('all', this)">Tất Cả</button>
          <button class="tab-btn" onclick="setFilter('active', this)">🟢 Đang Chạy</button>
          <button class="tab-btn" onclick="setFilter('expired', this)">🔴 Hết Hạn</button>
          <button class="tab-btn" onclick="setFilter('disabled', this)">⚪ Đã Khóa</button>
        </div>

        <!-- Search Box -->
        <div class="search-input-wrap">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="11" cy="11" r="8"/><line x1="21" x2="16.65" y1="21" y2="16.65"/></svg>
          <input type="text" id="searchInput" class="search-input" placeholder="Tìm theo tên khách, port, user..." oninput="applyFilters()">
        </div>
      </div>

      <div style="display: flex; gap: 10px;">
        <button class="btn btn-secondary" onclick="loadData()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          Làm Mới
        </button>
        <button class="btn btn-primary" onclick="openCreateModal()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><line x1="12" x2="12" y1="5" y2="19"/><line x1="5" x2="19" y1="12" y2="12"/></svg>
          Tạo Proxy Mới
        </button>
      </div>
    </div>

    <!-- Table Card -->
    <div class="glass-card table-container">
      <div class="table-responsive">
        <table>
          <thead>
            <tr>
              <th>Proxy / Khách Hàng</th>
              <th>Cổng Kết Nối</th>
              <th>Xác Thực (Auth)</th>
              <th>Lưu Lượng Dùng</th>
              <th>Hạn Sử Dụng</th>
              <th>Chế Độ Xoay</th>
              <th>Trạng Thái</th>
              <th style="text-align: right;">Thao Tác</th>
            </tr>
          </thead>
          <tbody id="proxyTableBody">
            <tr>
              <td colspan="8" style="text-align: center; color: var(--text-dim); padding: 50px;">
                Đang nạp dữ liệu từ hệ thống...
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </div>

  <!-- CREATE / EDIT MODAL -->
  <div id="proxyModal" class="modal-backdrop">
    <div class="modal-box">
      <div class="modal-header">
        <h3 id="modalTitle">Tạo Proxy Mới</h3>
        <button class="btn btn-secondary btn-icon-only" onclick="closeModal('proxyModal')">
          <svg style="width: 14px; height: 14px;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
        </button>
      </div>
      <form id="proxyForm" onsubmit="saveProxy(event)">
        <input type="hidden" id="proxyId">
        <div class="modal-body">
          
          <div class="form-group">
            <label class="form-label">Tên Khách Hàng / Nhãn Ghi Chú</label>
            <input type="text" id="proxyName" class="form-control" required placeholder="VD: Khách Dàn Nuôi Facebook 01" oninput="updateLivePreview()">
          </div>

          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 14px;">
            <div class="form-group">
              <label class="form-label">Cổng Riêng (Port)</label>
              <input type="number" id="proxyPort" class="form-control" required placeholder="10001" oninput="updateLivePreview()">
            </div>
            <div class="form-group">
              <label class="form-label">Chế Độ Xoay IPv6</label>
              <select id="proxyRotation" class="form-control" onchange="toggleStickyInput()">
                <option value="request">Xoay Mỗi Request (100% Mới)</option>
                <option value="sticky">Xoay Theo Chu Kỳ Giây (10s, 30s, 60s... [Sticky])</option>
                <option value="static">Cố Định 1 IPv6 (Static IP)</option>
              </select>
            </div>
          </div>

          <div id="stickyGroup" class="form-group" style="display: none;">
            <div class="form-label">
              <span>Chu Kỳ Giữ IP Trước Khi Xoay (Giây)</span>
              <span style="font-size: 11px; color: #38bdf8; font-weight: normal;">VD: 10 = Xoay IP mới mỗi 10 giây</span>
            </div>
            <input type="number" id="proxyStickySec" class="form-control" placeholder="10" value="10">
            <div style="display: flex; gap: 6px; margin-top: 6px; flex-wrap: wrap;">
              <button type="button" class="btn btn-secondary btn-sm" style="font-size: 11px; padding: 3px 8px;" onclick="setStickyPreset(10)">⚡ 10 giây</button>
              <button type="button" class="btn btn-secondary btn-sm" style="font-size: 11px; padding: 3px 8px;" onclick="setStickyPreset(30)">⏱️ 30 giây</button>
              <button type="button" class="btn btn-secondary btn-sm" style="font-size: 11px; padding: 3px 8px;" onclick="setStickyPreset(60)">⏱️ 1 phút (60s)</button>
              <button type="button" class="btn btn-secondary btn-sm" style="font-size: 11px; padding: 3px 8px;" onclick="setStickyPreset(300)">⏱️ 5 phút</button>
            </div>
          </div>

          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 14px;">
            <div class="form-group">
              <div class="form-label">
                <span>Tài Khoản (User)</span>
                <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 6px; font-size: 11px;" onclick="genRandomUser()">🎲 Random</button>
              </div>
              <input type="text" id="proxyUser" class="form-control" placeholder="Để trống nếu không cần" oninput="updateLivePreview()">
            </div>
            <div class="form-group">
              <div class="form-label">
                <span>Mật Khẩu (Pass)</span>
                <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 6px; font-size: 11px;" onclick="genRandomPass()">🎲 Random</button>
              </div>
              <input type="text" id="proxyPass" class="form-control" placeholder="Để trống nếu không cần" oninput="updateLivePreview()">
            </div>
          </div>

          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 14px;">
            <div class="form-group">
              <label class="form-label">Hạn Mức Băng Thông (GB)</label>
              <input type="number" step="0.5" id="proxyMaxGB" class="form-control" placeholder="0 = Không giới hạn" value="0">
            </div>
            <div class="form-group">
              <label class="form-label">Thời Hạn Sử Dụng (Ngày)</label>
              <input type="number" id="proxyExpireDays" class="form-control" placeholder="0 = Vĩnh viễn" value="30">
            </div>
          </div>

          <div class="form-group">
            <label class="form-label" style="font-size: 12px; color: var(--text-dim);">Xem Trước Chuỗi Kết Nối Cho Khách:</label>
            <div class="preview-proxy-card">
              <span id="liveProxyPreview">--</span>
              <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 8px;" onclick="copyLivePreview()">Copy</button>
            </div>
          </div>

        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick="closeModal('proxyModal')">Đóng</button>
          <button type="submit" class="btn btn-primary" id="btnSubmitProxy">Lưu Cấu Hình</button>
        </div>
      </form>
    </div>
  </div>

  <!-- QUICK TEST / SNIPPET MODAL -->
  <div id="snippetModal" class="modal-backdrop">
    <div class="modal-box" style="max-width: 620px;">
      <div class="modal-header">
        <h3 id="snippetModalTitle">Hướng Dẫn Kết Nối Proxy</h3>
        <button class="btn btn-secondary btn-icon-only" onclick="closeModal('snippetModal')">
          <svg style="width: 14px; height: 14px;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
        </button>
      </div>
      <div class="modal-body">
        
        <div class="form-group">
          <label class="form-label">Định Dạng AdsPower / GoLogin / Hidemyacc:</label>
          <div class="input-with-action">
            <input type="text" id="snippetAntidetect" class="form-control" readonly style="font-family: 'JetBrains Mono', monospace;">
            <button class="btn btn-secondary" onclick="copyInputVal('snippetAntidetect')">Copy</button>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Lệnh Test cURL (Windows CMD / PowerShell / Linux):</label>
          <textarea id="snippetCurl" class="form-control" rows="3" readonly style="font-family: 'JetBrains Mono', monospace; font-size: 12.5px;"></textarea>
        </div>

        <div class="form-group">
          <label class="form-label">Code Mẫu Python (Requests):</label>
          <textarea id="snippetPython" class="form-control" rows="4" readonly style="font-family: 'JetBrains Mono', monospace; font-size: 12px;"></textarea>
        </div>

      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick="closeModal('snippetModal')">Đóng</button>
      </div>
    </div>
  </div>

  <!-- EXPORT MODAL -->
  <div id="exportModal" class="modal-backdrop">
    <div class="modal-box" style="max-width: 600px;">
      <div class="modal-header">
        <h3>Xuất Toàn Bộ Danh Sách Proxy</h3>
        <button class="btn btn-secondary btn-icon-only" onclick="closeModal('exportModal')">
          <svg style="width: 14px; height: 14px;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
        </button>
      </div>
      <div class="modal-body">
        <p style="font-size: 13px; color: var(--text-muted);">
          Định dạng chuẩn <code>IP:Port:User:Pass</code>, có thể nhập hàng loạt vào mọi trình duyệt Antidetect:
        </p>
        <textarea id="exportText" class="form-control" rows="11" readonly style="font-family: 'JetBrains Mono', monospace; font-size: 13px;"></textarea>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick="downloadExportFile()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
          Tải File .txt
        </button>
        <button class="btn btn-primary" onclick="copyExportText()">Sao Chép Tất Cả</button>
      </div>
    </div>
  </div>

  <!-- SETTINGS MODAL -->
  <div id="settingsModal" class="modal-backdrop">
    <div class="modal-box">
      <div class="modal-header">
        <h3>Cài Đặt Tài Khoản Quản Trị</h3>
        <button class="btn btn-secondary btn-icon-only" onclick="closeModal('settingsModal')">
          <svg style="width: 14px; height: 14px;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
        </button>
      </div>
      <form onsubmit="saveSettings(event)">
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">Mật Khẩu Hiện Tại</label>
            <input type="password" id="curPass" class="form-control" required placeholder="admin123">
          </div>
          <div class="form-group">
            <label class="form-label">Tài Khoản Mới (Username)</label>
            <input type="text" id="newAdminUser" class="form-control" placeholder="Để trống nếu giữ nguyên">
          </div>
          <div class="form-group">
            <label class="form-label">Mật Khẩu Mới (Password)</label>
            <input type="password" id="newAdminPass" class="form-control" placeholder="Để trống nếu giữ nguyên">
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick="closeModal('settingsModal')">Hủy</button>
          <button type="submit" class="btn btn-primary">Cập Nhật Thông Tin</button>
        </div>
      </form>
    </div>
  </div>

  <!-- TOAST CONTAINER -->
  <div id="toastContainer" class="toast-container"></div>

  <script>
    let globalProxies = [];
    let publicIP = "";
    let currentFilter = "all";

    function showToast(msg, type = "success") {
      const container = document.getElementById("toastContainer");
      const item = document.createElement("div");
      item.className = "toast-item";
      
      let icon = '<svg style="width: 18px; height: 18px; stroke: #34d399;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>';
      if (type === "error") {
        icon = '<svg style="width: 18px; height: 18px; stroke: #f43f5e;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"/><line x1="15" x2="9" y1="9" y2="15"/><line x1="9" x2="15" y1="9" y2="15"/></svg>';
      } else if (type === "info") {
        icon = '<svg style="width: 18px; height: 18px; stroke: #38bdf8;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"/><line x1="12" x2="12" y1="16" y2="12"/><line x1="12" x2="12.01" y1="8" y2="8"/></svg>';
      }

      item.innerHTML = icon + '<span>' + escapeHtml(msg) + '</span>';
      container.appendChild(item);

      setTimeout(() => {
        item.style.opacity = '0';
        item.style.transform = 'translateY(10px)';
        item.style.transition = 'all 0.3s';
        setTimeout(() => item.remove(), 300);
      }, 3000);
    }

    function openModal(id) {
      document.getElementById(id).style.display = "flex";
    }

    function closeModal(id) {
      document.getElementById(id).style.display = "none";
    }

    function setFilter(type, btn) {
      currentFilter = type;
      document.querySelectorAll(".tab-btn").forEach(b => b.classList.remove("active"));
      btn.classList.add("active");
      applyFilters();
    }

    function toggleStickyInput() {
      const r = document.getElementById("proxyRotation").value;
      document.getElementById("stickyGroup").style.display = (r === "sticky") ? "flex" : "none";
    }

    function setStickyPreset(sec) {
      document.getElementById("proxyStickySec").value = sec;
    }

    function genRandomUser() {
      document.getElementById("proxyUser").value = "user" + Math.floor(1000 + Math.random() * 9000);
      updateLivePreview();
    }

    function genRandomPass() {
      const chars = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789";
      let res = "";
      for (let i = 0; i < 8; i++) res += chars.charAt(Math.floor(Math.random() * chars.length));
      document.getElementById("proxyPass").value = res;
      updateLivePreview();
    }

    function updateLivePreview() {
      const port = document.getElementById("proxyPort").value || "PORT";
      const u = document.getElementById("proxyUser").value;
      const p = document.getElementById("proxyPass").value;
      let s = (publicIP || "IP_VPS") + ":" + port;
      if (u) s += ":" + u + ":" + p;
      document.getElementById("liveProxyPreview").innerText = s;
    }

    function copyLivePreview() {
      const text = document.getElementById("liveProxyPreview").innerText;
      copyToClipboard(text);
    }

    async function checkAuth() {
      try {
        const res = await fetch("/api/stats");
        if (res.ok) {
          document.getElementById("loginSection").style.display = "none";
          document.getElementById("appSection").style.display = "block";
          loadData();
        } else {
          document.getElementById("loginSection").style.display = "flex";
          document.getElementById("appSection").style.display = "none";
        }
      } catch (e) {
        document.getElementById("loginSection").style.display = "flex";
      }
    }

    async function doLogin(e) {
      e.preventDefault();
      const user = document.getElementById("loginUser").value;
      const pass = document.getElementById("loginPass").value;
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: user, password: pass })
      });
      const data = await res.json();
      if (res.ok) {
        showToast("Đăng nhập thành công!");
        checkAuth();
      } else {
        showToast(data.error || "Sai tài khoản hoặc mật khẩu", "error");
      }
    }

    async function doLogout() {
      await fetch("/api/logout", { method: "POST" });
      location.reload();
    }

    async function loadData() {
      try {
        const [statsRes, proxiesRes] = await Promise.all([
          fetch("/api/stats"),
          fetch("/api/proxies")
        ]);

        if (statsRes.ok) {
          const stats = await statsRes.json();
          publicIP = stats.public_ip;
          document.getElementById("topPublicIP").innerText = stats.public_ip;
          document.getElementById("kpiActive").innerText = stats.active_proxies + " / " + stats.total_proxies;
          document.getElementById("kpiTraffic").innerText = (stats.total_bytes / (1024*1024*1024)).toFixed(2) + " GB";
          document.getElementById("kpiPrefix").innerText = stats.prefix;
          document.getElementById("kpiUptime").innerText = stats.uptime;
          document.getElementById("kpiHost").innerText = "Máy chủ: " + stats.public_ip;

          if (!document.getElementById("proxyPort").value) {
            document.getElementById("proxyPort").value = stats.suggest_port;
            updateLivePreview();
          }
        }

        if (proxiesRes.ok) {
          globalProxies = await proxiesRes.json();
          applyFilters();
        }
      } catch (e) {
        console.error(e);
      }
    }

    function applyFilters() {
      const q = document.getElementById("searchInput").value.toLowerCase();
      const now = new Date();

      const filtered = globalProxies.filter(p => {
        // Query match
        const matchQ = p.name.toLowerCase().includes(q) ||
                       p.port.toString().includes(q) ||
                       (p.username && p.username.toLowerCase().includes(q));
        if (!matchQ) return false;

        // Status match
        if (currentFilter === "active") {
          return p.enabled && (!p.expires_at || new Date(p.expires_at) > now) && (p.max_bytes === 0 || p.bytes_used < p.max_bytes);
        } else if (currentFilter === "expired") {
          return (p.expires_at && new Date(p.expires_at) <= now) || (p.max_bytes > 0 && p.bytes_used >= p.max_bytes);
        } else if (currentFilter === "disabled") {
          return !p.enabled;
        }
        return true;
      });

      renderTable(filtered);
    }

    function renderTable(list) {
      const tbody = document.getElementById("proxyTableBody");
      if (!list || list.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" style="text-align: center; color: var(--text-dim); padding: 50px;">Không có proxy nào phù hợp với bộ lọc.</td></tr>';
        return;
      }

      tbody.innerHTML = list.map(p => {
        const usedGB = (p.bytes_used / (1024*1024*1024)).toFixed(2);
        const maxGB = p.max_bytes > 0 ? (p.max_bytes / (1024*1024*1024)).toFixed(2) : "∞";
        const percent = p.max_bytes > 0 ? Math.min(100, (p.bytes_used / p.max_bytes) * 100).toFixed(0) : 0;
        const isQuotaFull = p.max_bytes > 0 && p.bytes_used >= p.max_bytes;

        const isExpired = p.expires_at && new Date(p.expires_at) < new Date();

        let statusBadge = '<span class="badge badge-success"><span class="badge-dot"></span> Đang Chạy</span>';
        if (!p.enabled) {
          statusBadge = '<span class="badge badge-neutral"><span class="badge-dot"></span> Đã Khóa</span>';
        } else if (isExpired) {
          statusBadge = '<span class="badge badge-danger"><span class="badge-dot"></span> Hết Hạn</span>';
        } else if (isQuotaFull) {
          statusBadge = '<span class="badge badge-warning"><span class="badge-dot"></span> Hết GB</span>';
        }

        let rotBadge = '<span class="badge badge-indigo">Xoay Mỗi Request</span>';
        if (p.rotation_type === 'sticky') {
          rotBadge = '<span class="badge badge-neutral" style="color: #38bdf8; border-color: rgba(56, 189, 248, 0.3); background: rgba(56, 189, 248, 0.1);">🔄 Xoay mỗi ' + (p.sticky_sec || 10) + 's</span>';
        } else if (p.rotation_type === 'static') {
          rotBadge = '<span class="badge badge-neutral" style="color: #c084fc;">Cố Định 1 IP</span>';
        }

        let expireStr = '<span style="color: var(--text-dim);">Vĩnh viễn</span>';
        if (p.expires_at) {
          const exp = new Date(p.expires_at);
          const diffDays = Math.ceil((exp - new Date()) / (1000*60*60*24));
          if (diffDays > 0) {
            expireStr = '<span style="color: #34d399; font-weight: 600;">Còn ' + diffDays + ' ngày</span>';
          } else {
            expireStr = '<span style="color: #f43f5e; font-weight: 600;">Đã hết hạn</span>';
          }
        }

        const authStr = (p.username && p.password) ? (p.username + ":" + p.password) : "Không Auth";
        const proxyString = publicIP + ":" + p.port + (p.username ? (":" + p.username + ":" + p.password) : "");

        return '<tr>' +
          '<td>' +
            '<div class="client-title">' +
              '<div class="client-avatar">⚡</div>' +
              '<div>' +
                '<div>' + escapeHtml(p.name) + '</div>' +
                '<div style="font-size: 11.5px; color: var(--text-dim); font-weight: normal;">ID: ' + p.id.slice(0, 8) + '</div>' +
              '</div>' +
            '</div>' +
          '</td>' +
          '<td>' +
            '<span class="port-pill" onclick="copyToClipboard(\'' + p.port + '\')" title="Bấm để copy Port">' + p.port + '</span>' +
          '</td>' +
          '<td>' +
            '<span class="auth-snippet" onclick="copyToClipboard(\'' + proxyString + '\')" title="Bấm để copy chuỗi proxy">' +
              '<svg style="width: 13px; height: 13px; stroke: #94a3b8;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
              escapeHtml(authStr) +
            '</span>' +
          '</td>' +
          '<td>' +
            '<div class="traffic-box">' +
              '<div class="traffic-label">' +
                '<span>' + usedGB + ' GB</span>' +
                '<span style="color: var(--text-dim); font-weight: normal;">/ ' + maxGB + ' GB</span>' +
              '</div>' +
              '<div class="progress-track">' +
                '<div class="progress-fill ' + (isQuotaFull ? "full" : "") + '" style="width: ' + (p.max_bytes > 0 ? percent : 100) + '%;"></div>' +
              '</div>' +
            '</div>' +
          '</td>' +
          '<td>' + expireStr + '</td>' +
          '<td>' + rotBadge + '</td>' +
          '<td>' + statusBadge + '</td>' +
          '<td>' +
            '<div class="actions-group">' +
              '<button class="btn btn-secondary btn-icon-only" onclick="copyToClipboard(\'' + proxyString + '\')" title="Copy chuỗi proxy">' +
                '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>' +
              '</button>' +
              '<button class="btn btn-secondary btn-icon-only" onclick="openSnippetModal(\'' + p.id + '\')" title="Xem hướng dẫn kết nối & code test">' +
                '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>' +
              '</button>' +
              '<button class="btn btn-secondary btn-icon-only" onclick="openEditModal(\'' + p.id + '\')" title="Chỉnh sửa proxy">' +
                '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>' +
              '</button>' +
              '<button class="btn btn-secondary btn-icon-only" onclick="resetTraffic(\'' + p.id + '\')" title="Reset dung lượng về 0 GB">' +
                '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>' +
              '</button>' +
              '<button class="btn btn-secondary btn-icon-only" onclick="toggleEnable(\'' + p.id + '\', ' + !p.enabled + ')" title="' + (p.enabled ? "Khóa proxy" : "Mở khóa proxy") + '">' +
                (p.enabled 
                  ? '<svg style="stroke: #f59e0b;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>'
                  : '<svg style="stroke: #10b981;" viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 9.9-1"/></svg>') +
              '</button>' +
              '<button class="btn btn-danger btn-icon-only" onclick="deleteProxy(\'' + p.id + '\')" title="Xóa">' +
                '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>' +
              '</button>' +
            '</div>' +
          '</td>' +
        '</tr>';
      }).join('');
    }

    function openCreateModal() {
      document.getElementById("modalTitle").innerText = "Tạo Proxy Mới";
      document.getElementById("btnSubmitProxy").innerText = "Tạo Proxy Mới";
      document.getElementById("proxyId").value = "";
      document.getElementById("proxyName").value = "";
      document.getElementById("proxyUser").value = "user" + Math.floor(1000 + Math.random() * 9000);
      document.getElementById("proxyPass").value = Math.random().toString(36).slice(-8);
      document.getElementById("proxyMaxGB").value = "0";
      document.getElementById("proxyExpireDays").value = "30";
      document.getElementById("proxyRotation").value = "request";
      document.getElementById("proxyStickySec").value = "10";
      toggleStickyInput();
      updateLivePreview();
      openModal("proxyModal");
    }

    function openEditModal(id) {
      const p = globalProxies.find(x => x.id === id);
      if (!p) return;

      document.getElementById("modalTitle").innerText = "Chỉnh Sửa Proxy: " + p.name;
      document.getElementById("btnSubmitProxy").innerText = "Lưu Thay Đổi";
      document.getElementById("proxyId").value = p.id;
      document.getElementById("proxyName").value = p.name;
      document.getElementById("proxyPort").value = p.port;
      document.getElementById("proxyUser").value = p.username || "";
      document.getElementById("proxyPass").value = p.password || "";
      document.getElementById("proxyMaxGB").value = p.max_bytes > 0 ? (p.max_bytes / (1024*1024*1024)).toFixed(1) : 0;
      
      let expDays = 0;
      if (p.expires_at) {
        const diff = Math.ceil((new Date(p.expires_at) - new Date()) / (1000*60*60*24));
        expDays = diff > 0 ? diff : 0;
      }
      document.getElementById("proxyExpireDays").value = expDays;
      document.getElementById("proxyRotation").value = p.rotation_type || "request";
      document.getElementById("proxyStickySec").value = p.sticky_sec || 60;

      toggleStickyInput();
      updateLivePreview();
      openModal("proxyModal");
    }

    async function saveProxy(e) {
      e.preventDefault();
      const id = document.getElementById("proxyId").value;
      const payload = {
        name: document.getElementById("proxyName").value,
        port: parseInt(document.getElementById("proxyPort").value) || 0,
        username: document.getElementById("proxyUser").value,
        password: document.getElementById("proxyPass").value,
        max_gb: parseFloat(document.getElementById("proxyMaxGB").value) || 0,
        expire_days: parseInt(document.getElementById("proxyExpireDays").value) || 0,
        rotation_type: document.getElementById("proxyRotation").value,
        sticky_sec: parseInt(document.getElementById("proxyStickySec").value) || 60
      };

      const url = id ? ("/api/proxies/" + id) : "/api/proxies";
      const method = id ? "PUT" : "POST";

      const res = await fetch(url, {
        method: method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (res.ok) {
        showToast(id ? "Đã cập nhật proxy thành công!" : "Đã tạo proxy mới thành công!");
        closeModal("proxyModal");
        loadData();
      } else {
        showToast(data.error || "Có lỗi xảy ra", "error");
      }
    }

    async function toggleEnable(id, enabled) {
      const res = await fetch("/api/proxies/" + id, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled: enabled })
      });
      if (res.ok) {
        showToast(enabled ? "Đã mở khóa proxy" : "Đã khóa proxy");
        loadData();
      }
    }

    async function resetTraffic(id) {
      if (!confirm("Bạn có chắc chắn muốn reset dung lượng đã dùng về 0 GB?")) return;
      const res = await fetch("/api/proxies/" + id + "/reset", { method: "POST" });
      if (res.ok) {
        showToast("Đã reset dung lượng về 0 GB!");
        loadData();
      }
    }

    async function deleteProxy(id) {
      if (!confirm("Bạn có chắc chắn muốn xóa vĩnh viễn proxy này?")) return;
      const res = await fetch("/api/proxies/" + id, { method: "DELETE" });
      if (res.ok) {
        showToast("Đã xóa proxy thành công!");
        loadData();
      }
    }

    function openSnippetModal(id) {
      const p = globalProxies.find(x => x.id === id);
      if (!p) return;

      document.getElementById("snippetModalTitle").innerText = "Cấu Hình Kết Nối: " + p.name;
      const authPart = (p.username && p.password) ? (p.username + ":" + p.password) : "";
      
      const antidetectStr = publicIP + ":" + p.port + (authPart ? (":" + authPart) : "");
      document.getElementById("snippetAntidetect").value = antidetectStr;

      const authCurl = authPart ? (" -U " + authPart) : "";
      const curlStr = "curl.exe -s -x http://" + publicIP + ":" + p.port + authCurl + " https://api64.ipify.org";
      document.getElementById("snippetCurl").value = curlStr;

      const pyAuth = authPart ? (p.username + ":" + p.password + "@") : "";
      const pyStr = "import requests\n\nproxy = 'http://" + pyAuth + publicIP + ":" + p.port + "'\nres = requests.get('https://api64.ipify.org?format=json', proxies={'http': proxy, 'https': proxy})\nprint('IPv6 xoay:', res.json()['ip'])";
      document.getElementById("snippetPython").value = pyStr;

      openModal("snippetModal");
    }

    async function openExportModal() {
      const res = await fetch("/api/export");
      const text = await res.text();
      document.getElementById("exportText").value = text || "Chưa có proxy nào đang hoạt động.";
      openModal("exportModal");
    }

    function copyExportText() {
      const t = document.getElementById("exportText");
      t.select();
      navigator.clipboard.writeText(t.value);
      showToast("Đã sao chép toàn bộ danh sách proxy!");
    }

    function downloadExportFile() {
      const text = document.getElementById("exportText").value;
      const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "proxies_" + (new Date().toISOString().slice(0, 10)) + ".txt";
      a.click();
      showToast("Đã tải file text danh sách proxy!");
    }

    function copyInputVal(id) {
      const v = document.getElementById(id).value;
      copyToClipboard(v);
    }

    function copyToClipboard(text) {
      if (!text) return;
      navigator.clipboard.writeText(text);
      showToast("Đã sao chép: " + text, "info");
    }

    function openSettingsModal() {
      document.getElementById("curPass").value = "";
      document.getElementById("newAdminUser").value = "";
      document.getElementById("newAdminPass").value = "";
      openModal("settingsModal");
    }

    async function saveSettings(e) {
      e.preventDefault();
      const cur = document.getElementById("curPass").value;
      const u = document.getElementById("newAdminUser").value;
      const p = document.getElementById("newAdminPass").value;

      const res = await fetch("/api/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ current_password: cur, new_username: u, new_password: p })
      });
      const data = await res.json();
      if (res.ok) {
        showToast("Đã đổi thông tin quản trị thành công!");
        closeModal("settingsModal");
      } else {
        showToast(data.error || "Lỗi cập nhật mật khẩu", "error");
      }
    }

    function escapeHtml(str) {
      if (!str) return '';
      return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
    }

    // Khởi tạo
    checkAuth();
  </script>
</body>
</html>
`
