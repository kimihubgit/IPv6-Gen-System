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
	w.Header().Set("Content-Type", "application/json")
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
	if err == nil {
		ws.mu.Lock()
		delete(ws.sessions, cookie.Value)
		ws.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "admin_token",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (ws *WebServer) handleStats(w http.ResponseWriter, r *http.Request) {
	proxies := ws.store.GetAll()
	totalProxies := len(proxies)
	activeProxies := 0
	var totalBytes int64 = 0

	for _, p := range proxies {
		totalBytes += p.BytesUsed
		if p.Enabled && !p.IsExpired() && !p.IsTrafficExceeded() {
			activeProxies++
		}
	}

	uptime := time.Since(ws.startTime).Round(time.Second).String()

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"total_proxies":  totalProxies,
		"active_proxies": activeProxies,
		"total_bytes":    totalBytes,
		"prefix":         ws.prefix,
		"public_ip":      ws.publicIPv4,
		"uptime":         uptime,
		"suggest_port":   ws.store.NextAvailablePort(10001),
	})
}

func (ws *WebServer) handleProxies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		proxies := ws.store.GetAll()
		jsonResponse(w, http.StatusOK, proxies)

	case http.MethodPost:
		var req struct {
			Name         string `json:"name"`
			Port         int    `json:"port"`
			Username     string `json:"username"`
			Password     string `json:"password"`
			MaxGB        float64 `json:"max_gb"`        // 0 = Không giới hạn
			ExpireDays   int    `json:"expire_days"`   // 0 = Vĩnh viễn
			RotationType string `json:"rotation_type"` // request | sticky | static
			StickySec    int    `json:"sticky_sec"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu JSON không hợp lệ"})
			return
		}

		if req.Port <= 0 || req.Port > 65535 {
			req.Port = ws.store.NextAvailablePort(10001)
		}
		if req.Name == "" {
			req.Name = fmt.Sprintf("Proxy-%d", req.Port)
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
		if rotType != RotationSticky && rotType != RotationStatic {
			rotType = RotationPerRequest
		}
		if rotType == RotationSticky && req.StickySec <= 0 {
			req.StickySec = 60
		}

		acc := &ProxyAccount{
			ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
			Name:         req.Name,
			Port:         req.Port,
			Username:     req.Username,
			Password:     req.Password,
			Proto:        "both",
			MaxBytes:     maxBytes,
			BytesUsed:    0,
			ExpiresAt:    expiresAt,
			RotationType: rotType,
			StickySec:    req.StickySec,
			Enabled:      true,
			CreatedAt:    time.Now(),
		}

		if err := ws.store.Add(acc); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := ws.manager.SyncAccount(acc); err != nil {
			_ = ws.store.Delete(acc.ID)
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Không thể mở port: %v", err)})
			return
		}

		_ = ws.store.Save()
		jsonResponse(w, http.StatusCreated, acc)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ws *WebServer) handleProxyItem(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/proxies/")
	parts := strings.Split(id, "/")
	proxyID := parts[0]

	acc := ws.store.GetByID(proxyID)
	if acc == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Không tìm thấy proxy"})
		return
	}

	// Sub-action: /api/proxies/:id/reset
	if len(parts) > 1 && parts[1] == "reset" && r.Method == http.MethodPost {
		acc.BytesUsed = 0
		_ = ws.store.Save()
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Đã reset dung lượng về 0 GB"})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req struct {
			Name         string   `json:"name"`
			Username     string   `json:"username"`
			Password     string   `json:"password"`
			MaxGB        *float64 `json:"max_gb"`
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
  <title>IPv6 Proxy Hub - Cloud Manager</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #0b0f19;
      --card-bg: rgba(18, 24, 38, 0.75);
      --card-border: rgba(255, 255, 255, 0.08);
      --primary: #6366f1;
      --primary-hover: #4f46e5;
      --primary-glow: rgba(99, 102, 241, 0.25);
      --success: #10b981;
      --warning: #f59e0b;
      --danger: #ef4444;
      --text: #f3f4f6;
      --text-muted: #9ca3af;
      --glass: backdrop-filter: blur(16px);
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: 'Inter', sans-serif;
      background: var(--bg);
      color: var(--text);
      min-height: 100vh;
      background-image: 
        radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.12) 0px, transparent 50%),
        radial-gradient(at 100% 100%, rgba(16, 185, 129, 0.08) 0px, transparent 50%);
      background-attachment: fixed;
    }
    .container { max-width: 1300px; margin: 0 auto; padding: 24px; }
    header {
      display: flex; justify-content: space-between; align-items: center;
      padding-bottom: 24px; border-bottom: 1px solid var(--card-border);
      margin-bottom: 24px;
    }
    .logo { display: flex; align-items: center; gap: 12px; }
    .logo-icon {
      width: 44px; height: 44px; border-radius: 12px;
      background: linear-gradient(135deg, #6366f1, #a855f7);
      display: flex; align-items: center; justify-content: center;
      font-size: 22px; box-shadow: 0 0 20px var(--primary-glow);
    }
    .logo-text h1 { font-size: 20px; font-weight: 800; letter-spacing: -0.5px; }
    .logo-text p { font-size: 13px; color: var(--text-muted); }
    .btn {
      padding: 10px 18px; border-radius: 10px; font-weight: 600;
      font-size: 14px; cursor: pointer; transition: all 0.2s;
      display: inline-flex; align-items: center; gap: 8px;
      border: 1px solid transparent; text-decoration: none;
    }
    .btn-primary {
      background: var(--primary); color: white;
      box-shadow: 0 4px 14px var(--primary-glow);
    }
    .btn-primary:hover { background: var(--primary-hover); transform: translateY(-1px); }
    .btn-secondary {
      background: rgba(255, 255, 255, 0.06); color: var(--text);
      border-color: var(--card-border);
    }
    .btn-secondary:hover { background: rgba(255, 255, 255, 0.1); }
    .btn-danger { background: rgba(239, 68, 68, 0.15); color: #f87171; border-color: rgba(239, 68, 68, 0.3); }
    .btn-danger:hover { background: rgba(239, 68, 68, 0.25); }
    .btn-sm { padding: 6px 12px; font-size: 12px; border-radius: 8px; }

    /* Stats Grid */
    .stats-grid {
      display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
      gap: 16px; margin-bottom: 28px;
    }
    .stat-card {
      background: var(--card-bg); border: 1px solid var(--card-border);
      border-radius: 16px; padding: 20px; backdrop-filter: blur(12px);
      display: flex; flex-direction: column; gap: 8px;
    }
    .stat-title { font-size: 13px; color: var(--text-muted); font-weight: 500; }
    .stat-value { font-size: 26px; font-weight: 800; color: #fff; }
    .stat-sub { font-size: 12px; color: var(--text-muted); }

    /* Section & Toolbar */
    .toolbar {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 16px; gap: 12px; flex-wrap: wrap;
    }
    .search-box {
      background: rgba(255, 255, 255, 0.05); border: 1px solid var(--card-border);
      padding: 10px 16px; border-radius: 10px; color: #fff; width: 280px;
      font-size: 14px; outline: none;
    }
    .search-box:focus { border-color: var(--primary); }

    /* Table */
    .table-card {
      background: var(--card-bg); border: 1px solid var(--card-border);
      border-radius: 16px; overflow: hidden; backdrop-filter: blur(12px);
    }
    table { width: 100%; border-collapse: collapse; text-align: left; }
    th {
      background: rgba(255, 255, 255, 0.03); padding: 14px 18px;
      font-size: 12px; font-weight: 600; text-transform: uppercase;
      letter-spacing: 0.5px; color: var(--text-muted);
      border-bottom: 1px solid var(--card-border);
    }
    td {
      padding: 16px 18px; font-size: 14px;
      border-bottom: 1px solid rgba(255, 255, 255, 0.04);
      vertical-align: middle;
    }
    tr:last-child td { border-bottom: none; }
    tr:hover td { background: rgba(255, 255, 255, 0.02); }

    .badge {
      display: inline-flex; align-items: center; gap: 6px;
      padding: 4px 10px; border-radius: 20px; font-size: 12px;
      font-weight: 600;
    }
    .badge-success { background: rgba(16, 185, 129, 0.15); color: #34d399; }
    .badge-warning { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }
    .badge-danger { background: rgba(239, 68, 68, 0.15); color: #f87171; }
    .badge-muted { background: rgba(255, 255, 255, 0.08); color: var(--text-muted); }

    .progress-bar {
      height: 6px; background: rgba(255, 255, 255, 0.1);
      border-radius: 3px; overflow: hidden; margin-top: 6px; width: 120px;
    }
    .progress-fill {
      height: 100%; background: linear-gradient(90deg, #6366f1, #a855f7);
      border-radius: 3px; transition: width 0.3s;
    }

    .auth-box {
      font-family: monospace; font-size: 13px;
      background: rgba(0, 0, 0, 0.3); padding: 4px 8px;
      border-radius: 6px; border: 1px solid rgba(255, 255, 255, 0.05);
      cursor: pointer; user-select: all;
    }

    /* Modal */
    .modal-overlay {
      position: fixed; inset: 0; background: rgba(0, 0, 0, 0.7);
      backdrop-filter: blur(8px); display: none; align-items: center;
      justify-content: center; z-index: 1000; padding: 20px;
    }
    .modal {
      background: #111827; border: 1px solid var(--card-border);
      border-radius: 20px; width: 100%; max-width: 520px;
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.7);
      overflow: hidden;
    }
    .modal-header {
      padding: 20px 24px; border-bottom: 1px solid var(--card-border);
      display: flex; justify-content: space-between; align-items: center;
    }
    .modal-header h3 { font-size: 18px; font-weight: 700; }
    .modal-body { padding: 24px; display: flex; flex-direction: column; gap: 16px; }
    .form-group { display: flex; flex-direction: column; gap: 6px; }
    .form-group label { font-size: 13px; font-weight: 600; color: var(--text-muted); }
    .form-control {
      background: rgba(255, 255, 255, 0.05); border: 1px solid var(--card-border);
      padding: 10px 14px; border-radius: 10px; color: #fff; font-size: 14px;
      outline: none; transition: border 0.2s;
    }
    .form-control:focus { border-color: var(--primary); }
    .modal-footer {
      padding: 16px 24px; border-top: 1px solid var(--card-border);
      display: flex; justify-content: flex-end; gap: 10px;
    }

    /* Toast */
    .toast {
      position: fixed; bottom: 24px; right: 24px; z-index: 2000;
      background: rgba(17, 24, 39, 0.95); border: 1px solid var(--card-border);
      padding: 14px 20px; border-radius: 12px; font-size: 14px; font-weight: 500;
      box-shadow: 0 10px 30px rgba(0,0,0,0.5); backdrop-filter: blur(10px);
      display: none; align-items: center; gap: 10px;
    }
    .toast.show { display: flex; animation: slideIn 0.3s; }
    @keyframes slideIn { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }

    /* Login Screen */
    .login-wrapper {
      min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 20px;
    }
    .login-card {
      background: var(--card-bg); border: 1px solid var(--card-border);
      border-radius: 24px; padding: 36px; width: 100%; max-width: 400px;
      backdrop-filter: blur(20px); box-shadow: 0 20px 40px rgba(0,0,0,0.5);
    }
  </style>
</head>
<body>

  <!-- LOGIN SCREEN -->
  <div id="loginSection" class="login-wrapper" style="display: none;">
    <div class="login-card">
      <div style="text-align: center; margin-bottom: 28px;">
        <div class="logo-icon" style="margin: 0 auto 16px; width: 56px; height: 56px; font-size: 28px;">🌐</div>
        <h2 style="font-size: 22px; font-weight: 800;">Đăng Nhập Quản Trị</h2>
        <p style="font-size: 13px; color: var(--text-muted); margin-top: 4px;">IPv6 Rotating Proxy Hub</p>
      </div>
      <form id="loginForm" onsubmit="doLogin(event)">
        <div class="form-group" style="margin-bottom: 14px;">
          <label>Tài khoản</label>
          <input type="text" id="loginUser" class="form-control" required placeholder="admin" autofocus>
        </div>
        <div class="form-group" style="margin-bottom: 24px;">
          <label>Mật khẩu</label>
          <input type="password" id="loginPass" class="form-control" required placeholder="••••••••">
        </div>
        <button type="submit" class="btn btn-primary" style="width: 100%; justify-content: center; padding: 12px;">Đăng Nhập</button>
      </form>
    </div>
  </div>

  <!-- DASHBOARD -->
  <div id="appSection" class="container" style="display: none;">
    <header>
      <div class="logo">
        <div class="logo-icon">🌐</div>
        <div class="logo-text">
          <h1>IPv6 Proxy Hub</h1>
          <p>Hệ thống xoay IPv6 không giới hạn (ANY-IP + ndppd)</p>
        </div>
      </div>
      <div style="display: flex; gap: 10px;">
        <button class="btn btn-secondary" onclick="openExportModal()">📋 Export Proxy</button>
        <button class="btn btn-secondary" onclick="openSettingsModal()">⚙️ Đổi Pass Admin</button>
        <button class="btn btn-danger btn-sm" onclick="doLogout()">Đăng Xuất</button>
      </div>
    </header>

    <!-- Stats -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-title">Proxy Đang Hoạt Động</div>
        <div class="stat-value" id="statActive">0 / 0</div>
        <div class="stat-sub">Các cổng sẵn sàng phục vụ</div>
      </div>
      <div class="stat-card">
        <div class="stat-title">Tổng Dung Lượng Đã Dùng</div>
        <div class="stat-value" id="statTraffic">0.00 GB</div>
        <div class="stat-sub">Băng thông toàn bộ proxy</div>
      </div>
      <div class="stat-card">
        <div class="stat-title">Dải IPv6 Subnet</div>
        <div class="stat-value" style="font-size: 16px; font-family: monospace;" id="statPrefix">--</div>
        <div class="stat-sub">18,446,744,073,709,551,616 IPs</div>
      </div>
      <div class="stat-card">
        <div class="stat-title">IP VPS / Uptime</div>
        <div class="stat-value" style="font-size: 18px;" id="statIP">--</div>
        <div class="stat-sub" id="statUptime">Uptime: --</div>
      </div>
    </div>

    <!-- Toolbar -->
    <div class="toolbar">
      <div style="display: flex; gap: 10px; align-items: center;">
        <input type="text" id="searchInput" class="search-box" placeholder="🔍 Tìm theo tên, port, user..." oninput="filterTable()">
        <button class="btn btn-secondary btn-sm" onclick="loadData()">🔄 Làm Mới</button>
      </div>
      <button class="btn btn-primary" onclick="openCreateModal()">➕ Tạo Proxy Mới</button>
    </div>

    <!-- Table -->
    <div class="table-card">
      <table>
        <thead>
          <tr>
            <th>Proxy / Khách Hàng</th>
            <th>Cổng (Port)</th>
            <th>Xác Thực (User:Pass)</th>
            <th>Dung Lượng (GB)</th>
            <th>Hạn Sử Dụng</th>
            <th>Chế Độ Xoay</th>
            <th>Trạng Thái</th>
            <th style="text-align: right;">Thao Tác</th>
          </tr>
        </thead>
        <tbody id="proxyTableBody">
          <tr><td colspan="8" style="text-align: center; color: var(--text-muted); padding: 40px;">Đang tải dữ liệu...</td></tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- CREATE / EDIT MODAL -->
  <div id="proxyModal" class="modal-overlay">
    <div class="modal">
      <div class="modal-header">
        <h3 id="modalTitle">Tạo Proxy Mới</h3>
        <button class="btn btn-secondary btn-sm" onclick="closeModal('proxyModal')">✕</button>
      </div>
      <form id="proxyForm" onsubmit="saveProxy(event)">
        <input type="hidden" id="proxyId">
        <div class="modal-body">
          <div class="form-group">
            <label>Tên Khách Hàng / Ghi Chú</label>
            <input type="text" id="proxyName" class="form-control" required placeholder="VD: Khách Anh Tuấn - Dàn Nuôi FB">
          </div>
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px;">
            <div class="form-group">
              <label>Cổng Kết Nối (Port)</label>
              <input type="number" id="proxyPort" class="form-control" required placeholder="10001">
            </div>
            <div class="form-group">
              <label>Chế Độ Xoay IP</label>
              <select id="proxyRotation" class="form-control" onchange="toggleStickyInput()">
                <option value="request">Xoay mỗi Request (Mới 100%)</option>
                <option value="sticky">Giữ IP X giây (Sticky)</option>
                <option value="static">Cố định 1 IP (Tĩnh)</option>
              </select>
            </div>
          </div>
          <div id="stickyGroup" class="form-group" style="display: none;">
            <label>Thời Gian Giữ IP (Giây)</label>
            <input type="number" id="proxyStickySec" class="form-control" placeholder="60" value="60">
          </div>
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px;">
            <div class="form-group">
              <label>Tài Khoản (Username)</label>
              <input type="text" id="proxyUser" class="form-control" placeholder="Để trống nếu không cần">
            </div>
            <div class="form-group">
              <label>Mật Khẩu (Password)</label>
              <input type="text" id="proxyPass" class="form-control" placeholder="Để trống nếu không cần">
            </div>
          </div>
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px;">
            <div class="form-group">
              <label>Giới Hạn Dung Lượng (GB)</label>
              <input type="number" step="0.1" id="proxyMaxGB" class="form-control" placeholder="0 = Không giới hạn" value="0">
            </div>
            <div class="form-group">
              <label>Thời Hạn (Số Ngày)</label>
              <input type="number" id="proxyExpireDays" class="form-control" placeholder="0 = Vĩnh viễn" value="30">
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick="closeModal('proxyModal')">Hủy</button>
          <button type="submit" class="btn btn-primary">Lưu Proxy</button>
        </div>
      </form>
    </div>
  </div>

  <!-- EXPORT MODAL -->
  <div id="exportModal" class="modal-overlay">
    <div class="modal" style="max-width: 600px;">
      <div class="modal-header">
        <h3>Danh Sách Proxy (IP:Port:User:Pass)</h3>
        <button class="btn btn-secondary btn-sm" onclick="closeModal('exportModal')">✕</button>
      </div>
      <div class="modal-body">
        <p style="font-size: 13px; color: var(--text-muted);">Định dạng chuẩn để import thẳng vào AdsPower, GoLogin, Hidemyacc:</p>
        <textarea id="exportText" class="form-control" rows="10" readonly style="font-family: monospace; font-size: 13px;"></textarea>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick="closeModal('exportModal')">Đóng</button>
        <button class="btn btn-primary" onclick="copyExportText()">📋 Sao Chép Toàn Bộ</button>
      </div>
    </div>
  </div>

  <!-- SETTINGS MODAL -->
  <div id="settingsModal" class="modal-overlay">
    <div class="modal">
      <div class="modal-header">
        <h3>Cài Đặt Quản Trị</h3>
        <button class="btn btn-secondary btn-sm" onclick="closeModal('settingsModal')">✕</button>
      </div>
      <form onsubmit="saveSettings(event)">
        <div class="modal-body">
          <div class="form-group">
            <label>Mật khẩu hiện tại</label>
            <input type="password" id="curPass" class="form-control" required placeholder="admin123">
          </div>
          <div class="form-group">
            <label>Tài khoản mới (Username)</label>
            <input type="text" id="newAdminUser" class="form-control" placeholder="admin">
          </div>
          <div class="form-group">
            <label>Mật khẩu mới (Password)</label>
            <input type="password" id="newAdminPass" class="form-control" placeholder="••••••••">
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick="closeModal('settingsModal')">Hủy</button>
          <button type="submit" class="btn btn-primary">Cập Nhật</button>
        </div>
      </form>
    </div>
  </div>

  <!-- TOAST -->
  <div id="toast" class="toast">
    <span id="toastIcon">✅</span>
    <span id="toastMsg">Thao tác thành công</span>
  </div>

  <script>
    let globalProxies = [];
    let publicIP = "";

    function showToast(msg, icon = "✅") {
      const t = document.getElementById("toast");
      document.getElementById("toastMsg").innerText = msg;
      document.getElementById("toastIcon").innerText = icon;
      t.classList.add("show");
      setTimeout(() => t.classList.remove("show"), 3000);
    }

    function openModal(id) { document.getElementById(id).style.display = "flex"; }
    function closeModal(id) { document.getElementById(id).style.display = "none"; }

    function toggleStickyInput() {
      const r = document.getElementById("proxyRotation").value;
      document.getElementById("stickyGroup").style.display = (r === "sticky") ? "flex" : "none";
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
        showToast(data.error || "Sai tài khoản hoặc mật khẩu", "❌");
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
          document.getElementById("statActive").innerText = stats.active_proxies + " / " + stats.total_proxies;
          document.getElementById("statTraffic").innerText = (stats.total_bytes / (1024*1024*1024)).toFixed(2) + " GB";
          document.getElementById("statPrefix").innerText = stats.prefix;
          document.getElementById("statIP").innerText = stats.public_ip;
          document.getElementById("statUptime").innerText = "Uptime: " + stats.uptime;
          document.getElementById("proxyPort").placeholder = stats.suggest_port;
          if (!document.getElementById("proxyPort").value) {
            document.getElementById("proxyPort").value = stats.suggest_port;
          }
        }

        if (proxiesRes.ok) {
          globalProxies = await proxiesRes.json();
          renderTable(globalProxies);
        }
      } catch (e) {
        console.error(e);
      }
    }

    function renderTable(list) {
      const tbody = document.getElementById("proxyTableBody");
      if (!list || list.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" style="text-align: center; color: var(--text-muted); padding: 40px;">Chưa có proxy nào. Hãy bấm "Tạo Proxy Mới" ở góc trên bên phải!</td></tr>';
        return;
      }

      tbody.innerHTML = list.map(p => {
        const usedGB = (p.bytes_used / (1024*1024*1024)).toFixed(2);
        const maxGB = p.max_bytes > 0 ? (p.max_bytes / (1024*1024*1024)).toFixed(2) : "∞";
        const percent = p.max_bytes > 0 ? Math.min(100, (p.bytes_used / p.max_bytes) * 100).toFixed(0) : 0;

        let statusBadge = '<span class="badge badge-success">🟢 Đang Chạy</span>';
        if (!p.enabled) {
          statusBadge = '<span class="badge badge-muted">⚪ Đã Khóa</span>';
        } else if (p.expires_at && new Date(p.expires_at) < new Date()) {
          statusBadge = '<span class="badge badge-danger">🔴 Hết Hạn</span>';
        } else if (p.max_bytes > 0 && p.bytes_used >= p.max_bytes) {
          statusBadge = '<span class="badge badge-warning">🟠 Hết Dung Lượng</span>';
        }

        let rotBadge = '<span class="badge badge-muted">Mỗi Request</span>';
        if (p.rotation_type === 'sticky') {
          rotBadge = '<span class="badge badge-muted">Giữ ' + (p.sticky_sec || 60) + 's</span>';
        } else if (p.rotation_type === 'static') {
          rotBadge = '<span class="badge badge-muted">Cố Định 1 IP</span>';
        }

        let expireStr = "Vĩnh viễn";
        if (p.expires_at) {
          const exp = new Date(p.expires_at);
          const diffDays = Math.ceil((exp - new Date()) / (1000*60*60*24));
          expireStr = diffDays > 0 ? (diffDays + " ngày nữa") : "Đã hết hạn";
        }

        const authStr = (p.username && p.password) ? (p.username + ":" + p.password) : "Không Auth";
        const proxyString = publicIP + ":" + p.port + (p.username ? (":" + p.username + ":" + p.password) : "");

        return '<tr>' +
          '<td><strong>' + escapeHtml(p.name) + '</strong></td>' +
          '<td><span style="font-weight: 700; color: #a5b4fc;">' + p.port + '</span> <span style="font-size: 11px; color: var(--text-muted);">(HTTP/SOCKS5)</span></td>' +
          '<td><span class="auth-box" onclick="copyToClipboard(\'' + proxyString + '\')" title="Bấm để copy chuỗi proxy">' + escapeHtml(authStr) + '</span></td>' +
          '<td>' +
            '<div>' + usedGB + ' / ' + maxGB + ' GB</div>' +
            (p.max_bytes > 0 ? '<div class="progress-bar"><div class="progress-fill" style="width: ' + percent + '%;"></div></div>' : '') +
          '</td>' +
          '<td>' + expireStr + '</td>' +
          '<td>' + rotBadge + '</td>' +
          '<td>' + statusBadge + '</td>' +
          '<td style="text-align: right;">' +
            '<button class="btn btn-secondary btn-sm" style="margin-right: 4px;" onclick="copyToClipboard(\'' + proxyString + '\')" title="Copy chuỗi proxy">📋</button>' +
            '<button class="btn btn-secondary btn-sm" style="margin-right: 4px;" onclick="resetTraffic(\'' + p.id + '\')" title="Reset dung lượng về 0">🔄 0 GB</button>' +
            '<button class="btn btn-secondary btn-sm" style="margin-right: 4px;" onclick="toggleEnable(\'' + p.id + '\', ' + !p.enabled + ')">' + (p.enabled ? "Khóa" : "Mở") + '</button>' +
            '<button class="btn btn-danger btn-sm" onclick="deleteProxy(\'' + p.id + '\')" title="Xóa">🗑️</button>' +
          '</td>' +
        '</tr>';
      }).join('');
    }

    function filterTable() {
      const q = document.getElementById("searchInput").value.toLowerCase();
      const filtered = globalProxies.filter(p => 
        p.name.toLowerCase().includes(q) ||
        p.port.toString().includes(q) ||
        (p.username && p.username.toLowerCase().includes(q))
      );
      renderTable(filtered);
    }

    function openCreateModal() {
      document.getElementById("modalTitle").innerText = "Tạo Proxy Mới";
      document.getElementById("proxyId").value = "";
      document.getElementById("proxyName").value = "";
      document.getElementById("proxyUser").value = "user" + Math.floor(1000 + Math.random() * 9000);
      document.getElementById("proxyPass").value = Math.random().toString(36).slice(-8);
      document.getElementById("proxyMaxGB").value = "0";
      document.getElementById("proxyExpireDays").value = "30";
      document.getElementById("proxyRotation").value = "request";
      toggleStickyInput();
      openModal("proxyModal");
    }

    async function saveProxy(e) {
      e.preventDefault();
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

      const res = await fetch("/api/proxies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (res.ok) {
        showToast("Đã tạo proxy thành công!");
        closeModal("proxyModal");
        loadData();
      } else {
        showToast(data.error || "Lỗi tạo proxy", "❌");
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
        showToast("Đã reset dung lượng về 0 GB");
        loadData();
      }
    }

    async function deleteProxy(id) {
      if (!confirm("Bạn có chắc chắn muốn xóa proxy này?")) return;
      const res = await fetch("/api/proxies/" + id, { method: "DELETE" });
      if (res.ok) {
        showToast("Đã xóa proxy");
        loadData();
      }
    }

    async function openExportModal() {
      const res = await fetch("/api/export");
      const text = await res.text();
      document.getElementById("exportText").value = text;
      openModal("exportModal");
    }

    function copyExportText() {
      const t = document.getElementById("exportText");
      t.select();
      navigator.clipboard.writeText(t.value);
      showToast("Đã sao chép toàn bộ danh sách proxy!");
    }

    function copyToClipboard(text) {
      navigator.clipboard.writeText(text);
      showToast("Đã copy: " + text);
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
        showToast("Đã đổi thông tin đăng nhập thành công!");
        closeModal("settingsModal");
      } else {
        showToast(data.error || "Lỗi đổi thông tin", "❌");
      }
    }

    function escapeHtml(str) {
      if (!str) return '';
      return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
    }

    // Init
    checkAuth();
  </script>
</body>
</html>
`
