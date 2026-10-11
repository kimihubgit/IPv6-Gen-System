package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"sync"
	"time"

	"ipv6-gen-linux/internal/engine"
	"ipv6-gen-linux/internal/store"
)

// WebServer manages the Web Admin Dashboard and REST APIs
type WebServer struct {
	port       int
	prefix     string
	publicIPv4 string
	manager    *engine.MultiProxyManager
	store      *store.ProxyStore
	adminUser  string
	adminPass  string
	sessions   map[string]time.Time
	mu         sync.RWMutex
	startTime  time.Time
	httpServer *http.Server
}

// NewWebServer creates a new Web Dashboard server
func NewWebServer(port int, prefix string, publicIPv4 string, manager *engine.MultiProxyManager, store *store.ProxyStore) *WebServer {
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
		adminPass:  "admin123",
		sessions:   make(map[string]time.Time),
		startTime:  time.Now(),
	}
}

// Start launches the Web Dashboard HTTP listener
func (ws *WebServer) Start() error {
	_ = mime.AddExtensionType(".jsx", "text/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".js", "text/javascript; charset=utf-8")

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/login", ws.handleLogin)
	mux.HandleFunc("/api/logout", ws.handleLogout)
	mux.HandleFunc("/api/stats", ws.authMiddleware(ws.handleStats))
	mux.HandleFunc("/api/proxies", ws.authMiddleware(ws.handleProxies))
	mux.HandleFunc("/api/proxies/bulk", ws.authMiddleware(ws.handleProxiesBulk))
	mux.HandleFunc("/api/proxies/", ws.authMiddleware(ws.handleProxyItem))
	mux.HandleFunc("/api/export", ws.authMiddleware(ws.handleExport))
	mux.HandleFunc("/api/settings", ws.authMiddleware(ws.handleSettings))

	// Static Theme & UI routes
	webDir := "web"
	if fi, err := os.Stat(webDir); err == nil && fi.IsDir() {
		log.Println("📁 Đang phục vụ theme/giao diện trực tiếp từ thư mục đĩa: ./web")
		mux.Handle("/", http.FileServer(http.Dir(webDir)))
	} else {
		log.Println("⚠️ Thư mục ./web không tồn tại trên đĩa")
	}

	addr := fmt.Sprintf("0.0.0.0:%d", ws.port)
	log.Printf("🚀 Web Dashboard đã sẵn sàng tại: http://%s:%d (Đăng nhập: %s / %s)",
		ws.publicIPv4, ws.port, ws.adminUser, ws.adminPass)

	ws.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return ws.httpServer.ListenAndServe()
}

// Stop terminates the Web Dashboard gracefully
func (ws *WebServer) Stop(ctx context.Context) error {
	if ws.httpServer != nil {
		return ws.httpServer.Shutdown(ctx)
	}
	return nil
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
