# 🌐 IPv6 Rotator & Generator System (Go)

Hệ thống chuyên dụng viết bằng **Golang** được **tách biệt hoàn toàn thành 2 thư mục độc lập**:
- 🪟 Thư mục **`windows/`**: Dành riêng 100% cho máy tính Windows (Source code Windows, script `.bat`, file `.exe`).
- 🐧 Thư mục **`linux/`**: Dành riêng 100% cho VPS / Server Linux (Source code Linux, script `.sh`, dịch vụ systemd, binary Linux).

---

## 📁 Cấu Trúc Thư Mục Tách Biệt Độc Lập

```text
IPv6-Gen-System/
│
├── 🪟 windows/                      # MÃ NGUỒN & CÔNG CỤ DÀNH RIÊNG CHO WINDOWS
│   ├── main.go                     # File chạy chính tối ưu cho Windows
│   ├── network.go                  # Xử lý card mạng bằng netsh Windows
│   ├── admin.go                    # Xử lý quyền UAC / Administrator Windows
│   ├── pool.go                     # Chế độ Dynamic Rolling Pool (Xoay cuốn chiếu 200 IPs, không lag máy)
│   ├── proxy.go                    # Local Rotating Proxy (HTTP 10808 / SOCKS5 10809)
│   ├── generator.go                # Sinh ngẫu nhiên IPv6 theo dải prefix
│   ├── tester.go                   # Kiểm tra kết nối outbound
│   ├── config.go & network_common.go
│   ├── go.mod                      # Module Go độc lập của Windows
│   ├── Chay_Tool_Windows.bat       # Nhấp đúp chuột là chạy ngay (tự xin UAC)
│   ├── build.bat                   # Script biên dịch lại ipv6-gen.exe
│   ├── ipv6-gen.exe                # File thực thi Windows đã build sẵn
│   └── HUONG_DAN_WINDOWS.md        # Hướng dẫn chi tiết cho Windows
│
├── 🐧 linux/                        # MÃ NGUỒN & CÔNG CỤ DÀNH RIÊNG CHO LINUX / VPS
│   ├── main.go                     # File chạy chính tối ưu cho Linux
│   ├── network.go                  # Xử lý card mạng bằng iproute2 kernel Linux
│   ├── admin.go                    # Xử lý quyền root / sudo Linux
│   ├── pool.go                     # Dynamic Rolling Pool cho Linux
│   ├── proxy.go                    # Rotating Proxy Server
│   ├── generator.go & tester.go & config.go
│   ├── go.mod                      # Module Go độc lập của Linux
│   ├── run.sh                      # Script 1-click chạy ngay (tự cấp quyền + sudo)
│   ├── build.sh                    # Script biên dịch trên Linux
│   ├── install-service.sh          # Cài dịch vụ chạy ngầm 24/7 (systemd)
│   ├── ipv6-gen-linux              # File binary Linux đã build sẵn
│   └── HUONG_DAN_LINUX.md          # Hướng dẫn chi tiết cho VPS Linux
│
├── build_all.bat                   # Biên dịch tự động toàn bộ cả 2 bản
└── README.md
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
