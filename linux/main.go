package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func printBanner(isAdmin bool) {
	fmt.Println("==================================================================")
	fmt.Println("         🐧 LINUX IPV6 ROTATOR & GENERATOR TOOL (GO) 🐧          ")
	fmt.Println("   Sinh & Gán hàng loạt IPv6 vào Card Mạng - Tích hợp Rotating Proxy")
	fmt.Println("==================================================================")
	if isAdmin {
		fmt.Println(" [✓] Quyền thực thi: Root / Sudo (Đủ quyền cấu hình card mạng Linux)")
	} else {
		fmt.Println(" [!] CẢNH BÁO: Đang chạy ở quyền User thông thường!")
		fmt.Println("     -> Bạn cần chạy với 'sudo ./ipv6-gen-linux' để gán IP vào Linux.")
	}
	fmt.Println("------------------------------------------------------------------")
}

func readLine(scanner *bufio.Scanner, prompt string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	if scanner.Scan() {
		val := strings.TrimSpace(scanner.Text())
		if val == "" {
			return defaultVal
		}
		return val
	}
	return defaultVal
}

func selectInterface(scanner *bufio.Scanner) (*NetInterface, error) {
	ifaces, err := GetSystemInterfaces()
	if err != nil {
		return nil, err
	}
	if len(ifaces) == 0 {
		return nil, fmt.Errorf("không tìm thấy card mạng nào")
	}

	fmt.Println("\n📋 Danh sách Card Mạng khả dụng:")
	for i, iface := range ifaces {
		stateBadge := "🔴 " + iface.State
		if strings.ToLower(iface.State) == "connected" {
			stateBadge = "🟢 " + iface.State
		}
		prefix := DetectGlobalPrefix(iface)
		prefixInfo := ""
		if prefix != "" {
			prefixInfo = fmt.Sprintf(" | Đã có prefix: %s", prefix)
		}
		fmt.Printf("  [%d] %-25s (%s)%s\n", i+1, iface.Name, stateBadge, prefixInfo)
	}

	// Auto-select connected interface default
	defaultIdx := 1
	for i, iface := range ifaces {
		if strings.ToLower(iface.State) == "connected" {
			defaultIdx = i + 1
			break
		}
	}

	input := readLine(scanner, fmt.Sprintf("👉 Chọn số thứ tự card mạng (1-%d)", len(ifaces)), strconv.Itoa(defaultIdx))
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(ifaces) {
		return nil, fmt.Errorf("lựa chọn không hợp lệ")
	}

	selected := ifaces[choice-1]
	return &selected, nil
}

func handleAddIPv6(scanner *bufio.Scanner, isAdmin bool) {
	if !isAdmin {
		fmt.Println("\n❌ Bạn cần quyền Root / Sudo để gán IP vào card mạng Linux!")
		ans := readLine(scanner, "Bạn có muốn nâng quyền sudo không? (y/n)", "y")
		if strings.ToLower(ans) == "y" {
			err := RelaunchAsAdmin()
			if err != nil {
				fmt.Printf("Lỗi kích hoạt sudo: %v\n", err)
			} else {
				fmt.Println("Đã chạy lại với quyền sudo. Đang thoát tiến trình cũ...")
				os.Exit(0)
			}
		}
		return
	}

	iface, err := selectInterface(scanner)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	detectedPrefix := DetectGlobalPrefix(*iface)
	defaultPrefixPrompt := detectedPrefix
	if defaultPrefixPrompt == "" {
		defaultPrefixPrompt = "2402:800:6000:1234::/64"
	}

	fmt.Printf("\nCard mạng được chọn: [%s]\n", iface.Name)
	prefixInput := readLine(scanner, "👉 Nhập IPv6 Prefix (Nhấn Enter để dùng dải có sẵn)", defaultPrefixPrompt)
	cleanPrefix := strings.TrimSpace(prefixInput)
	if (cleanPrefix == "64" || cleanPrefix == "/64" || cleanPrefix == "48" || cleanPrefix == "/48") && detectedPrefix != "" {
		fmt.Printf("💡 Bạn vừa nhập '%s', hệ thống tự động chọn dải IPv6 có sẵn của máy: %s\n", cleanPrefix, detectedPrefix)
		prefixInput = detectedPrefix
	}
	ipNet, err := ParsePrefix(prefixInput)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	countStr := readLine(scanner, "👉 Nhập số lượng IPv6 cần tạo và gán", "20")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		fmt.Println("❌ Số lượng không hợp lệ!")
		return
	}

	modeChoice := readLine(scanner, "👉 Chế độ tạo IP: [1] Ngẫu nhiên (Random) | [2] Tuần tự (Sequential)", "1")
	mode := ModeRandom
	if modeChoice == "2" {
		mode = ModeSequential
	}

	storeChoice := readLine(scanner, "👉 Chế độ lưu: [1] Tạm thời (mất sau khi khởi động lại máy) | [2] Vĩnh viễn", "1")
	storeType := "active"
	if storeChoice == "2" {
		storeType = "persistent"
	}

	skipSourceChoice := readLine(scanner, "👉 Bật SkipAsSource=true (Khuyên dùng: để không làm ảnh hưởng mạng mặc định)? (y/n)", "y")
	skipAsSource := strings.ToLower(skipSourceChoice) == "y"

	fmt.Printf("\n⏳ Đang sinh %d địa chỉ IPv6...", count)
	ips, err := GenerateIPv6List(ipNet, count, mode)
	if err != nil {
		fmt.Printf("\n❌ Lỗi sinh IP: %v\n", err)
		return
	}
	fmt.Printf(" Đã xong!\n")

	// Save to txt file
	exportFile := fmt.Sprintf("ipv6_%s_%d.txt", time.Now().Format("20060102_150405"), count)
	if err := SaveIPListToFile(exportFile, ips); err == nil {
		fmt.Printf("📄 Đã lưu danh sách IP vào file: %s\n", exportFile)
	}

	// Begin adding to Windows
	fmt.Printf("\n🚀 Đang gán %d IPv6 vào card mạng '%s' (tiến trình đa luồng)...\n", count, iface.Name)
	successCount, errs := BatchAddIPv6(iface.Name, ips, 64, storeType, skipAsSource, func(done, total int) {
		pct := (done * 100) / total
		fmt.Printf("\r   Tiến độ: [%-50s] %d%% (%d/%d)", strings.Repeat("█", pct/2)+strings.Repeat(" ", 50-pct/2), pct, done, total)
	})
	fmt.Println()

	if len(errs) > 0 {
		fmt.Printf("⚠️  Hoàn tất với %d thành công, %d lỗi.\n", successCount, len(errs))
		fmt.Printf("   Ví dụ lỗi đầu tiên: %v\n", errs[0])
	} else {
		fmt.Printf("✅ THÀNH CÔNG: Đã gán toàn bộ %d IPv6 vào card mạng '%s'!\n", successCount, iface.Name)
	}

	// Track in assigned_ips.json
	ipStrs := make([]string, len(ips))
	for i, ip := range ips {
		ipStrs[i] = ip.String()
	}
	_ = SaveAssignedRecord(AssignedIPRecord{
		Interface: iface.Name,
		Prefix:    ipNet.String(),
		Store:     storeType,
		CreatedAt: time.Now(),
		IPs:       ipStrs,
	})

	// Offer to run test or proxy
	fmt.Println("\n------------------------------------------------------------------")
	postChoice := readLine(scanner, "👉 Bạn có muốn: [1] Bật Local Proxy ngay | [2] Kiểm tra kết nối IPv6 | [3] Quay lại Menu chính", "3")
	switch postChoice {
	case "1":
		startProxyWithIPs(scanner, ips)
	case "2":
		testConnectivityWithIPs(ips)
	}
}

func handleRemoveIPv6(scanner *bufio.Scanner, isAdmin bool) {
	if !isAdmin {
		fmt.Println("\n❌ Bạn cần quyền Root / Sudo để gỡ bỏ IP khỏi card mạng Linux!")
		return
	}

	records, err := LoadAssignedRecords()
	if err != nil || len(records) == 0 {
		fmt.Println("\nℹ️  Chưa tìm thấy lịch sử gán IP nào trong file `assigned_ips.json`.")
		fmt.Println("👉 Bạn có thể xóa theo tên card mạng & prefix thủ công.")
		manualClean := readLine(scanner, "Bạn có muốn quét xóa thủ công không? (y/n)", "n")
		if strings.ToLower(manualClean) == "y" {
			manualRemoveIPv6(scanner)
		}
		return
	}

	fmt.Println("\n📋 Lịch sử các đợt đã gán IPv6:")
	for i, r := range records {
		fmt.Printf("  [%d] Card: %-15s | Dải: %-25s | Số lượng: %-4d | Thời gian: %s\n",
			i+1, r.Interface, r.Prefix, len(r.IPs), r.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("  [%d] XÓA TẤT CẢ các đợt trên\n", len(records)+1)

	choiceStr := readLine(scanner, fmt.Sprintf("👉 Chọn đợt cần gỡ bỏ (1-%d)", len(records)+1), "1")
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(records)+1 {
		fmt.Println("❌ Lựa chọn không hợp lệ!")
		return
	}

	if choice == len(records)+1 {
		// Remove all
		for _, r := range records {
			fmt.Printf("\n⏳ Đang xóa %d IP trên card '%s'...\n", len(r.IPs), r.Interface)
			BatchDeleteIPv6(r.Interface, r.IPs, r.Store, func(done, total int) {
				fmt.Printf("\r   Đã xóa: %d/%d", done, total)
			})
			fmt.Println()
		}
		_ = ClearAllAssignedRecords()
		fmt.Println("✅ Đã hoàn tất xóa tất cả và làm sạch lịch sử!")
	} else {
		r := records[choice-1]
		fmt.Printf("\n⏳ Đang xóa %d IP trên card '%s'...\n", len(r.IPs), r.Interface)
		BatchDeleteIPv6(r.Interface, r.IPs, r.Store, func(done, total int) {
			fmt.Printf("\r   Đã xóa: %d/%d", done, total)
		})
		fmt.Println("\n✅ Đã hoàn tất gỡ bỏ đợt được chọn!")
	}
}

func manualRemoveIPv6(scanner *bufio.Scanner) {
	iface, err := selectInterface(scanner)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	prefixStr := readLine(scanner, "👉 Nhập dải IPv6 Prefix cần quét và gỡ (vd: 2402:800:1234::/64)", "")
	if prefixStr == "" {
		return
	}
	ipNet, err := ParsePrefix(prefixStr)
	if err != nil {
		fmt.Println("❌ Prefix không hợp lệ:", err)
		return
	}

	currentIPs := GetInterfaceIPv6Addresses(iface.Name)
	var toDelete []string
	for _, ip := range currentIPs {
		if ipNet.Contains(ip) {
			toDelete = append(toDelete, ip.String())
		}
	}

	if len(toDelete) == 0 {
		fmt.Printf("ℹ️  Không tìm thấy IP nào thuộc dải %s trên card %s.\n", prefixStr, iface.Name)
		return
	}

	fmt.Printf("Tìm thấy %d IP thuộc dải %s trên card %s.\n", len(toDelete), prefixStr, iface.Name)
	confirm := readLine(scanner, "Bạn có chắc chắn muốn xóa tất cả không? (y/n)", "y")
	if strings.ToLower(confirm) != "y" {
		return
	}

	BatchDeleteIPv6(iface.Name, toDelete, "active", func(done, total int) {
		fmt.Printf("\r   Đã xóa: %d/%d", done, total)
	})
	BatchDeleteIPv6(iface.Name, toDelete, "persistent", nil)
	fmt.Println("\n✅ Hoàn tất xóa các IP khớp dải prefix!")
}

func startProxyWithIPs(scanner *bufio.Scanner, ips []net.IP) {
	host := readLine(scanner, "👉 Địa chỉ lắng nghe (0.0.0.0 để máy Windows bên ngoài kết nối được)", "0.0.0.0")
	portStr := readLine(scanner, "👉 Cổng HTTP/HTTPS Proxy", "10808")
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		port = 10808
	}

	proxy := NewRotatingProxy(ips, host, port)
	err = proxy.Start()
	if err != nil {
		fmt.Println("Lỗi khởi chạy proxy:", err)
	}
}

func handleDynamicRollingProxy(scanner *bufio.Scanner, isAdmin bool) {
	if !isAdmin {
		fmt.Println("\n❌ Bạn cần quyền Root / Sudo để tự động gán và xoay IP trên card mạng Linux!")
		return
	}

	fmt.Println("\n==================================================================")
	fmt.Println("    🔄 CHẾ ĐỘ DYNAMIC ROLLING PROXY (XOAY CUỐN CHIẾU)")
	fmt.Println("   Duy trì cố định ~200 IP trên card, tự động xoay liên tục")
	fmt.Println("   👉 Không làm nghẽn bảng định tuyến Windows, không lag máy!")
	fmt.Println("==================================================================")

	iface, err := selectInterface(scanner)
	if err != nil {
		fmt.Println("❌ Lỗi chọn card mạng:", err)
		return
	}

	detectedPrefix := DetectGlobalPrefix(*iface)
	prefixPrompt := "👉 Nhập IPv6 Prefix (vd: 2402:800:xxxx:xxxx::/64)"
	prefixStr := readLine(scanner, prefixPrompt, detectedPrefix)
	if prefixStr == "" {
		fmt.Println("❌ Prefix không được để trống!")
		return
	}

	ipNet, err := ParsePrefix(prefixStr)
	if err != nil {
		fmt.Printf("❌ Prefix không hợp lệ: %v\n", err)
		return
	}

	poolSizeStr := readLine(scanner, "👉 Số IP duy trì trong Pool trên card (Khuyên dùng: 200 cho ~200 luồng)", "200")
	poolSize, err := strconv.Atoi(poolSizeStr)
	if err != nil || poolSize <= 0 {
		poolSize = 200
	}

	batchStr := readLine(scanner, "👉 Số IP xoay mới mỗi đợt", "30")
	batch, err := strconv.Atoi(batchStr)
	if err != nil || batch <= 0 {
		batch = 30
	}

	intervalStr := readLine(scanner, "👉 Chu kỳ tự động xoay (giây)", "30")
	intervalSec, err := strconv.Atoi(intervalStr)
	if err != nil || intervalSec <= 0 {
		intervalSec = 30
	}

	host := readLine(scanner, "👉 Địa chỉ lắng nghe (0.0.0.0 để máy Windows bên ngoài kết nối được)", "0.0.0.0")
	portStr := readLine(scanner, "👉 Cổng HTTP/HTTPS Proxy", "10808")
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		port = 10808
	}

	pool := NewRollingPool(iface.Name, ipNet, poolSize, batch, time.Duration(intervalSec)*time.Second)
	if err := pool.Start(); err != nil {
		fmt.Printf("❌ Lỗi khởi tạo Dynamic Pool: %v\n", err)
		return
	}

	proxy := NewDynamicRotatingProxy(pool, host, port)
	if err := proxy.Start(); err != nil {
		fmt.Printf("❌ Lỗi khởi chạy proxy: %v\n", err)
		pool.Stop()
	}
}

func handleProxyMenu(scanner *bufio.Scanner, isAdmin bool) {
	fmt.Println("\n📋 Chọn chế độ Proxy Server:")
	fmt.Println("  [1] 🔄 Dynamic Rolling Proxy (Tự động gán & xoay cuốn chiếu 200 IPs - Khuyên dùng cho đa luồng, không lag)")
	fmt.Println("  [2] 📌 Proxy tĩnh từ danh sách IP đã gán trước đó hoặc file .txt")

	modeChoice := readLine(scanner, "👉 Chọn chế độ (1-2)", "1")
	if modeChoice == "1" {
		handleDynamicRollingProxy(scanner, isAdmin)
		return
	}

	records, err := LoadAssignedRecords()
	var availableIPs []net.IP

	if err == nil && len(records) > 0 {
		fmt.Println("\n📋 Chọn nguồn danh sách IPv6 để làm Proxy tĩnh:")
		for i, r := range records {
			fmt.Printf("  [%d] Dải %s (%d IPs) trên card %s\n", i+1, r.Prefix, len(r.IPs), r.Interface)
		}
		fmt.Printf("  [%d] Nhập từ file .txt (danh sách IP)\n", len(records)+1)

		choiceStr := readLine(scanner, "👉 Chọn nguồn", "1")
		choice, _ := strconv.Atoi(choiceStr)
		if choice >= 1 && choice <= len(records) {
			for _, s := range records[choice-1].IPs {
				if ip := net.ParseIP(s); ip != nil {
					availableIPs = append(availableIPs, ip)
				}
			}
		}
	}

	if len(availableIPs) == 0 {
		file := readLine(scanner, "👉 Nhập đường dẫn file chứa danh sách IPv6 (.txt)", "ipv6_list.txt")
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ Không đọc được file: %v\n", err)
			return
		}
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if ip := net.ParseIP(l); ip != nil {
				availableIPs = append(availableIPs, ip)
			}
		}
	}

	if len(availableIPs) == 0 {
		fmt.Println("❌ Danh sách IP trống! Hãy gán IP hoặc nhập file hợp lệ.")
		return
	}

	startProxyWithIPs(scanner, availableIPs)
}

func testConnectivityWithIPs(ips []net.IP) {
	if len(ips) == 0 {
		fmt.Println("❌ Danh sách IP trống!")
		return
	}

	sampleSize := 5
	if len(ips) < sampleSize {
		sampleSize = len(ips)
	}

	fmt.Printf("\n🧪 Đang kiểm tra kết nối Internet thực tế với mẫu %d IPv6...\n", sampleSize)
	results := BatchTestIPv6(ips, sampleSize, func(res TestResult) {
		if res.Success {
			fmt.Printf("  [✓] %s -> Ra ngoài thành công! Public IP: %s (Độ trễ: %v)\n", res.IP, res.PublicIP, res.Latency.Round(time.Millisecond))
		} else {
			fmt.Printf("  [✗] %s -> Thất bại: %v\n", res.IP, res.Error)
		}
	})

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	fmt.Println("------------------------------------------------------------------")
	if successCount == len(results) {
		fmt.Printf("🎉 TẤT CẢ %d/%d IP kiểm tra đều kết nối Internet IPv6 thành công!\n", successCount, len(results))
		fmt.Println("   Dải IPv6 của bạn được nhà mạng/router định tuyến hoàn hảo.")
	} else if successCount > 0 {
		fmt.Printf("⚠️  Có %d/%d IP kết nối thành công.\n", successCount, len(results))
	} else {
		fmt.Printf("❌ Toàn bộ %d IP đều không kết nối ra Internet được.\n", len(results))
		fmt.Println("💡 Nguyên nhân thường gặp:")
		fmt.Println("   1. Dải /64 bạn nhập không thuộc dải mà nhà mạng (ISP) cấp phát cho router của bạn.")
		fmt.Println("   2. Router của bạn chưa bật chế độ IPv6 Routing hoặc chưa cấp phép dải này.")
		fmt.Println("   3. Nếu bạn dùng VPS/Tunnel, hãy đảm bảo routing /64 đã trỏ về máy local.")
	}
}

func handleGenerateOnly(scanner *bufio.Scanner) {
	prefixInput := readLine(scanner, "👉 Nhập IPv6 Prefix (Nhấn Enter để dùng dải mẫu)", "2402:800:6000:1234::/64")
	cleanPrefix := strings.TrimSpace(prefixInput)
	if cleanPrefix == "64" || cleanPrefix == "/64" || cleanPrefix == "48" || cleanPrefix == "/48" {
		prefixInput = "2402:800:6000:1234::/64"
	}
	ipNet, err := ParsePrefix(prefixInput)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	countStr := readLine(scanner, "👉 Số lượng IP cần sinh", "100")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		fmt.Println("❌ Số lượng không hợp lệ!")
		return
	}

	modeChoice := readLine(scanner, "👉 Chế độ: [1] Ngẫu nhiên (Random) | [2] Tuần tự (Sequential)", "1")
	mode := ModeRandom
	if modeChoice == "2" {
		mode = ModeSequential
	}

	outFile := readLine(scanner, "👉 Tên file xuất", "ipv6_list.txt")

	ips, err := GenerateIPv6List(ipNet, count, mode)
	if err != nil {
		fmt.Println("❌ Lỗi:", err)
		return
	}

	if err := SaveIPListToFile(outFile, ips); err != nil {
		fmt.Println("❌ Lỗi lưu file:", err)
		return
	}

	fmt.Printf("\n✅ Đã sinh thành công %d địa chỉ IPv6 và lưu vào file '%s'!\n", len(ips), outFile)
	fmt.Println("Mẫu 3 địa chỉ đầu tiên:")
	for i := 0; i < len(ips) && i < 3; i++ {
		fmt.Printf("  - %s\n", ips[i].String())
	}
}

func main() {
	defaultIface := "eth0"

	// Parse CLI flags for automated/headless usage
	actionFlag := flag.String("action", "", "Chức năng thực thi [add|delete|proxy|dynamic-proxy|generate|test]")
	ifaceFlag := flag.String("iface", defaultIface, "Tên card mạng Linux (vd: eth0, ens3, enp1s0)")
	prefixFlag := flag.String("prefix", "", "IPv6 Prefix (vd: 2402:800:1234:5678::/64)")
	countFlag := flag.Int("count", 50, "Số lượng IPv6 cần sinh")
	proxyHostFlag := flag.String("proxy-host", "0.0.0.0", "Địa chỉ IP lắng nghe (0.0.0.0 để máy ngoài VPS kết nối được, 127.0.0.1 nếu chỉ dùng nội bộ)")
	proxyPortFlag := flag.Int("proxy-port", 10808, "Cổng Local Proxy")
	poolSizeFlag := flag.Int("pool-size", 200, "Số lượng IP duy trì trong Dynamic Pool (cho đa luồng)")
	batchFlag := flag.Int("batch", 30, "Số lượng IP xoay mới mỗi đợt")
	intervalFlag := flag.Int("interval", 30, "Chu kỳ xoay (giây)")
	outputFlag := flag.String("out", "ipv6_output.txt", "File xuất kết quả")
	flag.Parse()

	isAdmin := IsAdmin()

	// If CLI flag specified, run non-interactive
	if *actionFlag != "" {
		switch strings.ToLower(*actionFlag) {
		case "generate":
			ipNet, err := ParsePrefix(*prefixFlag)
			if err != nil {
				fmt.Printf("Lỗi prefix: %v\n", err)
				os.Exit(1)
			}
			ips, _ := GenerateIPv6List(ipNet, *countFlag, ModeRandom)
			_ = SaveIPListToFile(*outputFlag, ips)
			fmt.Printf("Đã xuất %d IPs vào %s\n", len(ips), *outputFlag)
			return

		case "add":
			if !isAdmin {
				fmt.Println("Cần quyền Root / Sudo để gán IP!")
				os.Exit(1)
			}
			ipNet, err := ParsePrefix(*prefixFlag)
			if err != nil {
				fmt.Printf("Lỗi prefix: %v\n", err)
				os.Exit(1)
			}
			ips, _ := GenerateIPv6List(ipNet, *countFlag, ModeRandom)
			success, _ := BatchAddIPv6(*ifaceFlag, ips, 64, "active", true, nil)
			fmt.Printf("Đã gán %d/%d IPs vào card %s\n", success, len(ips), *ifaceFlag)
			return

		case "proxy":
			data, err := os.ReadFile(*outputFlag)
			if err != nil {
				fmt.Printf("Lỗi đọc file %s: %v\n", *outputFlag, err)
				os.Exit(1)
			}
			var ips []net.IP
			for _, line := range strings.Split(string(data), "\n") {
				if ip := net.ParseIP(strings.TrimSpace(line)); ip != nil {
					ips = append(ips, ip)
				}
			}
			p := NewRotatingProxy(ips, *proxyHostFlag, *proxyPortFlag)
			_ = p.Start()
			return

		case "dynamic-proxy", "rolling-proxy":
			if !isAdmin {
				fmt.Println("Cần quyền Root / Sudo để chạy dynamic rolling proxy!")
				os.Exit(1)
			}
			ipNet, err := ParsePrefix(*prefixFlag)
			if err != nil {
				fmt.Printf("Lỗi prefix: %v\n", err)
				os.Exit(1)
			}
			pool := NewRollingPool(*ifaceFlag, ipNet, *poolSizeFlag, *batchFlag, time.Duration(*intervalFlag)*time.Second)
			if err := pool.Start(); err != nil {
				fmt.Printf("Lỗi khởi tạo pool: %v\n", err)
				os.Exit(1)
			}
			p := NewDynamicRotatingProxy(pool, *proxyHostFlag, *proxyPortFlag)
			_ = p.Start()
			return
		}
	}

	// Interactive Menu
	scanner := bufio.NewScanner(os.Stdin)

	for {
		printBanner(isAdmin)
		fmt.Println("  [1] ⚡ Sinh & Gán danh sách IPv6 vào Card mạng (Tĩnh)")
		fmt.Println("  [2] 🧹 Gỡ bỏ IPv6 đã gán (Xem lịch sử & Clean up)")
		fmt.Println("  [3] 🔄 Khởi chạy Proxy Server (Có chế độ Dynamic Rolling Pool 200 IPs - Không lag máy)")
		fmt.Println("  [4] 🧪 Kiểm tra kết nối Internet thực tế của IPv6")
		fmt.Println("  [5] 📄 Chỉ sinh danh sách IPv6 ra file .txt (Không can thiệp card mạng)")
		if !isAdmin {
			fmt.Println("  [9] 🔑 Khởi động lại chương trình với quyền Root / Sudo")
		}
		fmt.Println("  [0] 🚪 Thoát")
		fmt.Println("------------------------------------------------------------------")

		choice := readLine(scanner, "👉 Chọn chức năng (0-5)", "1")

		switch choice {
		case "1":
			handleAddIPv6(scanner, isAdmin)
		case "2":
			handleRemoveIPv6(scanner, isAdmin)
		case "3":
			handleProxyMenu(scanner, isAdmin)
		case "4":
			// Test current interface or assigned records
			records, _ := LoadAssignedRecords()
			var sampleIPs []net.IP
			if len(records) > 0 {
				for _, s := range records[len(records)-1].IPs {
					if ip := net.ParseIP(s); ip != nil {
						sampleIPs = append(sampleIPs, ip)
					}
				}
			}
			if len(sampleIPs) == 0 {
				fmt.Println("\nChưa có danh sách IP đã gán. Hãy gán IP ở mục [1] trước hoặc kiểm tra qua file.")
			} else {
				testConnectivityWithIPs(sampleIPs)
			}
		case "5":
			handleGenerateOnly(scanner)
		case "9":
			if !isAdmin {
				_ = RelaunchAsAdmin()
				os.Exit(0)
			}
		case "0":
			fmt.Println("Tạm biệt!")
			return
		default:
			fmt.Println("Lựa chọn không hợp lệ, vui lòng chọn lại!")
		}

		fmt.Println("\nNhấn Enter để tiếp tục...")
		scanner.Scan()
	}
}
