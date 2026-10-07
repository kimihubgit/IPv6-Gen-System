# 🌐 IPv6 Rotator & Generator System (Golang)

Hệ thống chuyên dụng tối ưu cao viết bằng **Golang**, chia làm 2 nền tảng:
- 🐧 **`linux/` (Khuyên Dùng Cho VPS)**: Chế độ **ANY-IP Kernel + `ndppd`** — **Không cần gán IP vào card mạng (0 IP rác)**, tự động xoay vô hạn 18 tỷ tỷ IPv6 trong dải `/64` theo từng request HTTP/SOCKS5. Chạy ngầm 24/7 dưới dạng `systemd service`.
- 🪟 **`windows/` (Dành Cho Máy Cá Nhân Windows)**: Chế độ **Dynamic Rolling Pool** (xoay cuốn chiếu ~200 IP trong RAM qua `netsh`, không nghẽn mạng Windows) hoặc gán IPv6 tĩnh.

---

## 📁 Cấu Trúc Dự Án

```text
IPv6-Gen-System/
│
├── 🐧 linux/                         # DÀNH CHO VPS / SERVER LINUX (UBUNTU / DEBIAN / CENTOS)
│   ├── main.go                      # Chạy chính với chế độ ANY-IP & Proxy
│   ├── proxy.go                     # Rotating Proxy Server (HTTP: 10808, SOCKS5: 10809)
│   ├── network.go & network_common.go
│   ├── pool.go                      # Rolling Pool dự phòng
│   ├── generator.go & tester.go
│   ├── install-service.sh           # Script 1-click cài Proxy chạy ngầm 24/7 (systemd)
│   ├── run.sh                       # Chạy tương tác qua Menu
│   ├── build.sh                     # Script biên dịch trên Linux
│   ├── ipv6-gen-linux               # File binary Linux biên dịch sẵn (Statically Linked)
│   └── HUONG_DAN_LINUX.md           # 📖 TÀI LIỆU CHI TIẾT A-Z TRIỂN KHAI VPS (ANY-IP + ndppd)
│
├── 🪟 windows/                       # DÀNH CHO MÁY TÍNH WINDOWS
│   ├── main.go                      # Giao diện điều khiển Windows
│   ├── network.go                   # Tương tác netsh Windows
│   ├── admin.go                     # Tự động xin quyền Administrator (UAC)
│   ├── pool.go                      # Dynamic Rolling Pool 200 IPs
│   ├── proxy.go                     # Proxy nội bộ máy Windows
│   ├── Chay_Tool_Windows.bat        # 1-Click mở tool trên Windows
│   ├── build.bat                    # Biên dịch lại ipv6-gen.exe
│   └── HUONG_DAN_WINDOWS.md         # Hướng dẫn chi tiết cho Windows
│
├── build_all.bat                    # Script tự động biên dịch cả 2 bản cùng lúc
└── README.md
```

---

## 🐧 1. Tóm Tắt Triển Khai Trên Linux VPS (ANY-IP + ndppd)

👉 **Xem hướng dẫn chi tiết từ A-Z tại:** [linux/HUONG_DAN_LINUX.md](linux/HUONG_DAN_LINUX.md)

### Tóm tắt 4 bước nhanh:
1. **Cài đặt `ndppd`** (để tự động trả lời NDP cho toàn bộ dải `/64`):
   ```bash
   apt-get update && apt-get install -y ndppd
   ```
   Cấu hình `/etc/ndppd.conf`:
   ```conf
   proxy eth0 {
      router yes
      timeout 500
      ttl 30000
      rule 2403:6a40:0:15::/64 {
         static
      }
   }
   ```
   Khởi động: `systemctl restart ndppd && systemctl enable ndppd`

2. **Cấu hình Kernel**:
   ```bash
   sysctl -w net.ipv6.ip_nonlocal_bind=1
   sysctl -w net.ipv6.conf.all.forwarding=1
   ip -6 route add local 2403:6a40:0:15::/64 dev lo
   ```

3. **Cài dịch vụ Proxy ngầm 24/7**:
   ```bash
   cd linux
   chmod +x run.sh install-service.sh ipv6-gen-linux
   ./install-service.sh
   ```

4. **Mở Firewall**:
   ```bash
   ufw allow 10808/tcp && ufw allow 10809/tcp
   ```

---

## 🪟 2. Sử Dụng Proxy Trên Máy Windows

Sau khi VPS đã chạy Proxy:
* **Host / IP:** `<IP_VPS>` (Ví dụ: `42.96.15.130`)
* **HTTP Proxy Port:** `10808`
* **SOCKS5 Proxy Port:** `10809`

### Test trên PowerShell:
```powershell
curl.exe -s -x http://<IP_VPS>:10808 https://api64.ipify.org
curl.exe -s -x socks5h://<IP_VPS>:10809 https://api64.ipify.org
```
*Mỗi lần gọi sẽ ra một IPv6 hoàn toàn mới xuất phát từ VPS!*

### Dùng trong Antidetect Browser (AdsPower, GoLogin, Hidemyacc...):
* Chọn loại: `HTTP` hoặc `SOCKS5`
* Host: `<IP_VPS>`
* Port: `10808` hoặc `10809`
* User/Pass: Để trống.
