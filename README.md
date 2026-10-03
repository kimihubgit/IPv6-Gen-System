# 🌐 IPv6 Rotator & Generator System (Windows & Linux - Go)

Hệ thống chuyên dụng viết bằng **Golang** hỗ trợ **đa nền tảng (Windows & Linux)** giúp bạn:
1. **Sinh ngẫu nhiên hoặc tuần tự** hàng loạt địa chỉ IPv6 từ dải subnet `/64` hoặc `/48`.
2. **Gán đồng thời (đa luồng)** vào card mạng hệ thống:
   - **Windows**: Gán qua `netsh` với tùy chọn `skipassource=true` (an toàn, không ảnh hưởng duyệt web thông thường).
   - **Linux**: Gán siêu tốc qua kernel `ip -6 addr add`, tự động cấu hình `sysctl max_addresses=0` để gán hàng ngàn IP không bị chặn.
3. **Quản lý & Gỡ bỏ sạch sẽ**: Lưu lịch sử các đợt IP đã gán vào `assigned_ips.json` và hỗ trợ gỡ bỏ chỉ với 1 click.
4. **Tích hợp sẵn Local Rotating Proxy Server**:
   - Hỗ trợ cả **HTTP/HTTPS** (`http://127.0.0.1:10808`) và **SOCKS5** (`socks5://127.0.0.1:10809`).
   - Mỗi kết nối / request gửi qua proxy sẽ tự động xoay (round-robin / random) qua một địa chỉ IPv6 trong danh sách đã tạo.
   - Dễ dàng gắn vào các tool nuôi acc, crawler, bot, Antidetect Browser (AdsPower, GoLogin, Multilogin, Dolphin Anty...).
5. **Kiểm tra kết nối Internet thực tế**: Bắn request test outbound gắn `LocalAddr` vào từng IPv6 để kiểm tra router/ISP đã route thông IPv6 chưa.

---

## 📦 Đã Đóng Gói Sẵn 2 Bản Cho Bạn Sử Dụng

Thư mục đã được đóng gói sẵn để bạn dùng ngay:

```text
IPv6-Gen-System/
├── 🪟 release/windows/                 # BẢN DÀNH CHO WINDOWS (1-CLICK)
│   ├── ipv6-gen.exe                    # File thực thi Windows (x86_64)
│   ├── Chay_Tool_Windows.bat           # Nhấp đúp là chạy (Tự xin quyền Administrator)
│   └── HUONG_DAN_WINDOWS.md            # Hướng dẫn chi tiết sử dụng trên Windows
│
├── 🐧 release/linux/                   # BẢN DÀNH CHO LINUX / VPS (1-CLICK)
│   ├── ipv6-gen-linux                  # File thực thi Linux (Static binary, mọi distro)
│   ├── run.sh                          # Script 1-click (Tự cấp quyền + sudo + sysctl)
│   ├── install-service.sh              # 1-Click cài dịch vụ chạy ngầm 24/7 (systemd)
│   └── HUONG_DAN_LINUX.md              # Hướng dẫn chi tiết sử dụng trên VPS/Linux
│
├── build.bat                           # Script biên dịch tự động lại cả 2 bản trên Windows
├── build.sh                            # Script biên dịch tự động lại cả 2 bản trên Linux
└── ... mã nguồn Go đa nền tảng
```

---

## 🪟 1. Hướng Dẫn Sử Dụng Bản Windows

### Cách chạy nhanh nhất:
1. Vào thư mục `release/windows/` (hoặc ngay tại thư mục gốc).
2. Nhấp đúp chuột vào file **`Chay_Tool_Windows.bat`**.
3. Cửa sổ UAC của Windows sẽ bật lên -> Chọn **Yes** để cấp quyền Administrator.
4. Menu điều khiển sẽ hiện lên trực quan:
   - Bấm `[1]` để Sinh & Gán danh sách IPv6 vào Card mạng (`Wi-Fi` hoặc `Ethernet`).
   - Bấm `[3]` để Bật Rotating Proxy Server (`HTTP: 10808`, `SOCKS5: 10809`).
   - Bấm `[4]` để Kiểm tra kết nối Internet thực tế của các IPv6.
   - Bấm `[2]` để Gỡ bỏ sạch sẽ các IPv6 đã gán.

👉 *Xem hướng dẫn chi tiết tại:* [release/windows/HUONG_DAN_WINDOWS.md](release/windows/HUONG_DAN_WINDOWS.md)

---

## 🐧 2. Hướng Dẫn Sử Dụng Bản Linux / VPS

File thực thi `ipv6-gen-linux` được biên dịch tĩnh (**Statically Linked**), tương thích 100% với **Ubuntu, Debian, CentOS, AlmaLinux, Rocky Linux, Alpine...**

### Cách chạy nhanh nhất:
1. Tải thư mục `release/linux/` lên VPS (hoặc copy file `ipv6-gen-linux` và `run.sh`).
2. Chạy lệnh:
   ```bash
   chmod +x run.sh
   ./run.sh
   ```
   *(Script sẽ tự động gọi sudo và cấu hình kernel sysctl để gán không giới hạn IPv6).*

### Chạy Proxy ngầm 24/7 trên VPS (Không sợ tắt khi ngắt SSH):
Chạy script cài đặt dịch vụ nền:
```bash
sudo ./install-service.sh
```
- Dịch vụ `ipv6-proxy.service` sẽ được cài đặt và tự động khởi động cùng hệ thống.
- Quản lý dịch vụ:
  - `systemctl status ipv6-proxy` (Xem trạng thái)
  - `journalctl -u ipv6-proxy -f` (Xem log xoay IP trực tiếp)
  - `sudo systemctl restart ipv6-proxy` (Khởi động lại)
  - `sudo systemctl stop ipv6-proxy` (Dừng)

👉 *Xem hướng dẫn chi tiết tại:* [release/linux/HUONG_DAN_LINUX.md](release/linux/HUONG_DAN_LINUX.md)

---

## 🔌 Tích Hợp Proxy Vào Các Công Cụ

Sau khi bật chức năng **[3] Local Rotating Proxy Server** (hoặc chạy service):
- **HTTP / HTTPS Proxy**: `http://127.0.0.1:10808`
- **SOCKS5 Proxy**: `socks5://127.0.0.1:10809`

### 1. Dùng trong Python (`requests`):
```python
import requests

proxies = {
    'http': 'http://127.0.0.1:10808',
    'https': 'http://127.0.0.1:10808',
}

for i in range(5):
    res = requests.get('https://api64.ipify.org?format=json', proxies=proxies)
    print(f"Lần {i+1}: {res.json()['ip']}")
```

### 2. Dùng trong cURL:
```bash
curl -x http://127.0.0.1:10808 https://api64.ipify.org
```

### 3. Dùng trong Antidetect Browser (AdsPower, GoLogin, Dolphin...):
- Chọn loại proxy: `HTTP` hoặc `SOCKS5`
- Host: `127.0.0.1`
- Port: `10808` (HTTP) hoặc `10809` (SOCKS5)

---

## 🔨 Tự Biên Dịch Lại (Rebuild)

Nếu bạn thay đổi mã nguồn và muốn build lại cả 2 bản:
- Trên Windows: Chạy file `build.bat`
- Trên Linux: Chạy file `./build.sh`

Lệnh build thủ công:
```bash
# Bản Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o release/windows/ipv6-gen.exe .

# Bản Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o release/linux/ipv6-gen-linux .
```
