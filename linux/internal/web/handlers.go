package web

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"

	"ipv6-gen-linux/internal/store"
)

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

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

	rotCounts := map[string]int{
		"request": 0,
		"sticky":  0,
		"static":  0,
	}

	now := time.Now()
	for _, p := range proxies {
		totalBytes += p.BytesUsed
		rotType := string(p.RotationType)
		if rotType == "" {
			rotType = "request"
		}
		rotCounts[rotType]++

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

	suggestPort := 10001
	usedPorts := make(map[int]bool)
	for _, p := range proxies {
		usedPorts[p.Port] = true
	}
	for usedPorts[suggestPort] {
		suggestPort++
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ramMB := fmt.Sprintf("%.1f MB", float64(m.Alloc)/(1024*1024))
	goroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()

	sortedProxies := make([]*store.ProxyAccount, len(proxies))
	copy(sortedProxies, proxies)
	sort.Slice(sortedProxies, func(i, j int) bool {
		return sortedProxies[i].BytesUsed > sortedProxies[j].BytesUsed
	})

	topProxies := make([]map[string]interface{}, 0, 5)
	for i := 0; i < len(sortedProxies) && i < 5; i++ {
		p := sortedProxies[i]
		topProxies = append(topProxies, map[string]interface{}{
			"id":            p.ID,
			"name":          p.Name,
			"username":      p.Username,
			"port":          p.Port,
			"bytes_used":    p.BytesUsed,
			"max_bytes":     p.MaxBytes,
			"rotation_type": p.RotationType,
			"sticky_sec":    p.StickySec,
			"enabled":       p.Enabled,
		})
	}

	res := map[string]interface{}{
		"active_proxies":  activeCount,
		"total_proxies":   len(proxies),
		"total_bytes":     totalBytes,
		"uptime":          uptimeStr,
		"prefix":          ws.prefix,
		"public_ip":       ws.publicIPv4,
		"suggest_port":    suggestPort,
		"ram_usage":       ramMB,
		"goroutines":      goroutines,
		"cpu_cores":       numCPU,
		"rotation_counts": rotCounts,
		"top_proxies":     topProxies,
		"ndp_status":      "Tối Ưu & An Toàn (Active)",
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
			Proto        string  `json:"proto"`
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

		rotType := store.RotationPolicy(req.RotationType)
		if rotType != store.RotationPerRequest && rotType != store.RotationSticky && rotType != store.RotationStatic {
			rotType = store.RotationSticky
		}

		stickySec := req.StickySec
		if stickySec < 1 {
			stickySec = 1
		}

		proto := strings.ToLower(strings.TrimSpace(req.Proto))
		if proto != "http" && proto != "socks5" && proto != "both" {
			proto = "both"
		}

		acc := &store.ProxyAccount{
			ID:           fmt.Sprintf("px-%d", time.Now().UnixNano()),
			Name:         req.Name,
			Port:         req.Port,
			Username:     req.Username,
			Password:     req.Password,
			Proto:        proto,
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

func (ws *WebServer) handleProxiesBulk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		StartPort    int     `json:"start_port"`
		Count        int     `json:"count"`
		NamePrefix   string  `json:"name_prefix"`
		Username     string  `json:"username"`
		Password     string  `json:"password"`
		Proto        string  `json:"proto"`
		MaxGB        float64 `json:"max_gb"`
		ExpireDays   int     `json:"expire_days"`
		RotationType string  `json:"rotation_type"`
		StickySec    int     `json:"sticky_sec"`
		AutoUserPass bool    `json:"auto_user_pass"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu JSON không hợp lệ"})
		return
	}

	if req.Count <= 0 || req.Count > 500 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Số lượng cổng phải từ 1 đến 500"})
		return
	}

	if req.StartPort < 1024 || req.StartPort+req.Count-1 > 65535 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dải cổng không hợp lệ (1024 - 65535)"})
		return
	}

	if req.NamePrefix == "" {
		req.NamePrefix = "Luồng"
	}

	rotType := store.RotationPolicy(req.RotationType)
	if rotType != store.RotationPerRequest && rotType != store.RotationSticky && rotType != store.RotationStatic {
		rotType = store.RotationSticky
	}

	stickySec := req.StickySec
	if stickySec < 1 {
		stickySec = 1
	}

	proto := strings.ToLower(strings.TrimSpace(req.Proto))
	if proto != "http" && proto != "socks5" && proto != "both" {
		proto = "both"
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

	for p := req.StartPort; p < req.StartPort+req.Count; p++ {
		if existing := ws.store.GetByPort(p); existing != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Cổng %d đã được dùng bởi proxy '%s'!", p, existing.Name),
			})
			return
		}
	}

	createdList := make([]*store.ProxyAccount, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		port := req.StartPort + i
		uname := req.Username
		pass := req.Password

		if req.AutoUserPass {
			uname = fmt.Sprintf("user_%d", port)
			pass = generateRandomString(8)
		}

		acc := &store.ProxyAccount{
			ID:           fmt.Sprintf("px-%d-%d", time.Now().UnixNano(), port),
			Name:         fmt.Sprintf("%s #%d (Port %d)", req.NamePrefix, i+1, port),
			Port:         port,
			Username:     uname,
			Password:     pass,
			Proto:        proto,
			MaxBytes:     maxBytes,
			BytesUsed:    0,
			ExpiresAt:    expiresAt,
			RotationType: rotType,
			StickySec:    stickySec,
			Enabled:      true,
			CreatedAt:    time.Now(),
		}

		if err := ws.store.Add(acc); err != nil {
			log.Printf("Lỗi thêm proxy port %d: %v", port, err)
			continue
		}

		if err := ws.manager.SyncAccount(acc); err != nil {
			log.Printf("Lỗi kích hoạt port %d: %v", port, err)
			_ = ws.store.Delete(acc.ID)
			continue
		}

		createdList = append(createdList, acc)
	}

	_ = ws.store.Save()
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"status":  "ok",
		"message": fmt.Sprintf("Đã tạo thành công %d cổng proxy!", len(createdList)),
		"count":   len(createdList),
		"proxies": createdList,
	})
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
			Proto        string   `json:"proto"`
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
		if req.Proto != "" {
			p := strings.ToLower(strings.TrimSpace(req.Proto))
			if p == "http" || p == "socks5" || p == "both" {
				acc.Proto = p
			}
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
			acc.RotationType = store.RotationPolicy(req.RotationType)
		}
		if req.StickySec > 0 {
			acc.StickySec = req.StickySec
		}
		if req.Enabled != nil {
			acc.Enabled = *req.Enabled
		}

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
			ws.manager.RemoveAccount(oldPort)
			acc.Port = req.Port
			_ = ws.manager.SyncAccount(acc)
		} else {
			_ = ws.manager.SyncAccount(acc)
		}

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
	format := r.URL.Query().Get("format")
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
