# 🌐 Windows IPv6 Rotator & Generator (Go)

Chương trình chuyên dụng viết bằng **Golang** giúp bạn:
1. **Sinh ngẫu nhiên hoặc tuần tự** hàng loạt địa chỉ IPv6 từ dải subnet `/64` hoặc `/48`.
2. **Gán đồng thời (đa luồng) vào card mạng Windows** (Wi-Fi, Ethernet...) thông qua `netsh` với tùy chọn an toàn `skipassource=true` (không làm ảnh hưởng tới kết nối duyệt web thông thường).
3. **Quản lý & Gỡ bỏ sạch sẽ**: Lưu lịch sử các đợt IP đã gán vào `assigned_ips.json` và hỗ trợ gỡ bỏ chỉ với 1 click.
4. **Tích hợp sẵn Local Rotating Proxy Server**:
   - Hỗ trợ cả **HTTP/HTTPS** (`http://127.0.0.1:10808`) và **SOCKS5** (`socks5://127.0.0.1:10809`).
   - Mỗi kết nối / request gửi qua proxy sẽ tự động xoay (round-robin / random) qua một địa chỉ IPv6 trong danh sách đã tạo.
   - Dễ dàng gắn vào các tool nuôi acc, crawler, bot, Antidetect Browser (AdsPower, GoLogin, Multilogin, Dolphin Anty...).
5. **Kiểm tra kết nối Internet thực tế**: Bắn request test outbound gắn `LocalAddr` vào từng IPv6 để kiểm tra router/ISP đã route thông IPv6 chưa.

---

## 📁 Cấu trúc thư mục

```text
ipv6 gen/
├── admin_windows.go    # Kiểm tra quyền Admin & gọi UAC tự động
├── config.go           # Quản lý lưu trữ trạng thái assigned_ips.json
├── generator.go        # Thuật toán sinh địa chỉ IPv6 (Random / Sequential)
├── network.go          # Quét card mạng & gán/xóa IP qua netsh đa luồng
├── proxy.go            # HTTP/HTTPS Tunneling & SOCKS5 Rotating Proxy
├── tester.go           # Module test kết nối IPv6 thực tế
├── main.go             # Giao diện dòng lệnh tương tác (Menu TUI) & CLI Flags
├── ipv6-gen.exe        # File thực thi đã biên dịch sẵn
└── README.md
```

---

## 🚀 Cách sử dụng

### Cách 1: Chạy giao diện tương tác (Khuyên dùng)

1. Nhấp chuột phải vào file **`ipv6-gen.exe`** và chọn **Run as administrator** (để có quyền cấu hình card mạng Windows).
   *(Nếu mở thông thường, chương trình sẽ hiển thị tùy chọn [9] để tự kích hoạt lại với quyền Admin).*
2. Màn hình Menu xuất hiện:
   ```text
   ==================================================================
          🌐 WINDOWS IPV6 ROTATOR & GENERATOR TOOL (GO) 🌐           
      Sinh & Gán hàng loạt IPv6 vào Card Mạng - Tích hợp Rotating Proxy
   ==================================================================
    [✓] Quyền thực thi: Administrator (Đủ quyền cấu hình card mạng)
   ------------------------------------------------------------------
     [1] ⚡ Sinh & Gán danh sách IPv6 vào Card mạng Windows
     [2] 🧹 Gỡ bỏ IPv6 đã gán (Xem lịch sử & Clean up)
     [3] 🚀 Khởi chạy Local Rotating Proxy Server (HTTP / SOCKS5)
     [4] 🧪 Kiểm tra kết nối Internet thực tế của IPv6
     [5] 📄 Chỉ sinh danh sách IPv6 ra file .txt (Không can thiệp card mạng)
     [0] 🚪 Thoát
   ```

#### Chi tiết các bước gán IP (Mục 1):
- **Chọn card mạng**: Tự động liệt kê các card mạng đang có (Wi-Fi, Ethernet) kèm trạng thái kết nối.
- **Nhập IPv6 Prefix**: Tự động nhận diện dải `/64` của mạng bạn (nếu có), hoặc bạn nhập dải IPv6 của bạn (ví dụ: `2402:800:6000:1234::/64`).
- **Nhập số lượng**: Số lượng IP muốn sinh (ví dụ: `50`, `100`, `500`).
- **Chế độ lưu**:
  - `active`: Tạm thời (khởi động lại máy sẽ tự sạch).
  - `persistent`: Vĩnh viễn (tồn tại cả sau khi reboot).
- **SkipAsSource**: Đặt là `true` để Windows không tự ý dùng các IP phụ này cho các ứng dụng thông thường, chỉ dùng khi công cụ chỉ định IP đó.
- Sau khi gán xong, chương trình tự động xuất ra 1 file text danh sách IP và lưu trạng thái vào `assigned_ips.json`.

---

### Cách 2: Sử dụng dòng lệnh (CLI Flags - Tự động hóa / Script)

Bạn có thể tích hợp vào file `.bat` hoặc script Python:

```bash
# 1. Chỉ sinh 100 IPv6 từ dải ra file text:
.\ipv6-gen.exe -action generate -prefix "2402:800:6000:1234::/64" -count 100 -out my_ips.txt

# 2. Gán 50 IPv6 vào card Wi-Fi (yêu cầu Admin):
.\ipv6-gen.exe -action add -iface "Wi-Fi" -prefix "2402:800:6000:1234::/64" -count 50

# 3. Khởi chạy Rotating Proxy Server từ danh sách file:
.\ipv6-gen.exe -action proxy -out my_ips.txt -proxy-port 10808
```

---

## 🔌 Tích hợp Proxy vào các công cụ

Sau khi bật chức năng **[3] Local Rotating Proxy Server**:
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

## 💡 Lưu ý quan trọng về Định tuyến IPv6 (Routing)

- Để một địa chỉ IPv6 có thể **truy cập Internet thực tế**, dải `/64` mà bạn gán bắt buộc phải thuộc dải subnet được nhà mạng (Viettel, VNPT, FPT...) cấp cho modem/router của bạn, hoặc được định tuyến qua 1 đường VPN/Tunnel (WireGuard, Hurricane Electric IPv6 Tunnel).
- Nếu bạn nhập một dải IP ngẫu nhiên không thuộc sở hữu của router, card mạng Windows vẫn gán thành công nhưng gói tin ra ngoài sẽ bị router hoặc ISP hủy (Drop). Bạn có thể dùng tính năng **[4] Kiểm tra kết nối** trong menu để xác minh.
