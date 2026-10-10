package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ipv6-gen-linux/internal/engine"
	"ipv6-gen-linux/internal/store"
	"ipv6-gen-linux/internal/web"
)

func detectPublicIPv4() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err == nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil && len(body) > 0 {
			return strings.TrimSpace(string(body))
		}
	}
	return "127.0.0.1"
}

func main() {
	var (
		actionFlag   string
		ifaceFlag    string
		prefixFlag   string
		webPortFlag  int
		dataFileFlag string
		publicIPFlag string
	)

	flag.StringVar(&actionFlag, "action", "server", "Hành động thực thi (server)")
	flag.StringVar(&ifaceFlag, "iface", "eth0", "Tên card mạng vật lý")
	flag.StringVar(&prefixFlag, "prefix", "2001:df4:14c0:7::/64", "Dải subnet IPv6 được cấp (/64)")
	flag.IntVar(&webPortFlag, "web-port", 9090, "Cổng lắng nghe Web Dashboard")
	flag.StringVar(&dataFileFlag, "data", "data/proxies.json", "Đường dẫn file lưu trữ proxy")
	flag.StringVar(&publicIPFlag, "public-ip", "", "Địa chỉ IPv4 công khai của máy chủ")
	flag.Parse()

	log.Println("==================================================================")
	log.Println("     🌐 IPV6 PROXY HUB ENTERPRISE SERVER (LINUX ANY-IP)           ")
	log.Println("==================================================================")

	// Validate Subnet Prefix
	_, ipNet, err := net.ParseCIDR(prefixFlag)
	if err != nil {
		log.Fatalf("❌ Subnet prefix không hợp lệ: %s (%v)", prefixFlag, err)
	}

	if publicIPFlag == "" {
		publicIPFlag = detectPublicIPv4()
	}

	// 1. Initialize Storage Layer
	proxyStore, err := store.NewProxyStore(dataFileFlag)
	if err != nil {
		log.Fatalf("❌ Lỗi khởi tạo cơ sở dữ liệu proxies: %v", err)
	}
	log.Printf("📦 Đã nạp %d proxy từ: %s", len(proxyStore.GetAll()), dataFileFlag)

	// 2. Initialize Engine Layer (Multi-port listeners + Any-IP + Sessions)
	manager := engine.NewMultiProxyManager(ifaceFlag, ipNet, proxyStore)
	if err := manager.Start(); err != nil {
		log.Fatalf("❌ Lỗi khởi động MultiProxyManager: %v", err)
	}

	// 3. Initialize Web Layer (Dashboard UI + REST APIs)
	ws := web.NewWebServer(webPortFlag, prefixFlag, publicIPFlag, manager, proxyStore)
	go func() {
		if err := ws.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Lỗi Web Dashboard: %v", err)
		}
	}()

	log.Println("==================================================================")
	log.Printf("   - Subnet Prefix:       %s", prefixFlag)
	log.Printf("   - VPS Public IPv4:     %s", publicIPFlag)
	log.Printf("   - Web Dashboard:       http://%s:%d", publicIPFlag, webPortFlag)
	log.Println("==================================================================")

	// 4. Graceful Shutdown on SIGINT / SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("🛑 Nhận tín hiệu %s. Đang tiến hành dừng hệ thống an toàn...", sig)

	// Shutdown Web Server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ws.Stop(ctx)

	// Stop Engine and release Any-IP routes
	manager.Stop()

	// Flush persistence
	_ = proxyStore.Save()

	log.Println("✅ Hệ thống đã tắt an toàn. Tạm biệt!")
}
