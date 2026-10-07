# 🌐 Hướng Dẫn Sử Dụng Tool IPv6 Trên Windows

Bản dành riêng cho hệ điều hành **Windows 10 / Windows 11 / Windows Server**.

---

## ⚡ Cách Chạy Đơn Giản Nhất (1-Click)

Bạn chỉ cần **nhấp đúp chuột vào file `Chay_Tool_Windows.bat`**:
- Script sẽ tự động kiểm tra và mở hộp thoại **UAC ("Yes")** để cấp quyền Administrator.
- Bạn **không cần** phải nhớ nhấn chuột phải chọn *"Run as administrator"*.

---

## 📋 Hướng Dẫn Các Chức Năng Chính

Sau khi mở giao diện Menu, bạn sẽ thấy 5 chức năng chính:

```text
==================================================================
        🌐 WINDOWS IPV6 ROTATOR & GENERATOR TOOL (GO) 🌐           
   Sinh & Gán hàng loạt IPv6 vào Card Mạng - Tích hợp Rotating Proxy
==================================================================
 [✓] Quyền thực thi: Administrator (Đủ quyền cấu hình card mạng)
------------------------------------------------------------------
  [1] ⚡ Sinh & Gán danh sách IPv6 vào Card mạng
  [2] 🧹 Gỡ bỏ IPv6 đã gán (Xem lịch sử & Clean up)
  [3] 🚀 Khởi chạy Local Rotating Proxy Server (HTTP / SOCKS5)
  [4] 🧪 Kiểm tra kết nối Internet thực tế của IPv6
  [5] 📄 Chỉ sinh danh sách IPv6 ra file .txt (Không can thiệp card mạng)
  [0] 🚪 Thoát
```

### 1. [Chức năng 1] Sinh & Gán IPv6 vào Card mạng Windows
- **Chọn Card mạng**: Nhập số thứ tự card mạng đang kết nối Internet (thường là `Wi-Fi` hoặc `Ethernet`). Tool sẽ tự động nhận diện và gợi ý sẵn dải IPv6 đang có của bạn.
- **Nhập IPv6 Prefix**: 
  - Nếu card mạng đã có dải IPv6 từ Router/ISP (ví dụ: `2402:800:6000:1234::/64`), bạn chỉ cần nhấn `Enter` để dùng mặc định.
  - Hoặc nhập dải IPv6 theo nhu cầu của bạn.
- **Số lượng IP cần sinh**: Nhập số lượng mong muốn (ví dụ: `50`, `100`, `500`).
- **Chế độ tạo**:
  - `[1] Random` (Ngẫu nhiên - khuyên dùng để tránh trùng lặp).
  - `[2] Sequential` (Tuần tự: 1, 2, 3...).
- **Chế độ lưu**:
  - `[1] Tạm thời (active)`: Tự động biến mất và làm sạch khi khởi động lại máy.
  - `[2] Vĩnh viễn (persistent)`: Vẫn giữ nguyên sau khi reboot.
- **SkipAsSource**: Chọn `y` (để Windows không dùng các IP này cho duyệt web thông thường, chỉ dùng khi công cụ yêu cầu).
- **Kết quả**: Tool chạy đa luồng gán cực nhanh vào card mạng, đồng thời xuất ra 1 file text (ví dụ: `ipv6_20260930_123456_50.txt`) và lưu vào `assigned_ips.json`.

---

### 2. [Chức năng 3] Khởi Chạy Local Rotating Proxy Server
Sau khi đã gán IP (hoặc có sẵn file text IP):
- Chọn nguồn IP: từ đợt gán gần nhất hoặc file text.
- Nhập Port HTTP: mặc định `10808`.
- Server sẽ tự động mở 2 cổng Proxy:
  - **HTTP / HTTPS Proxy**: `http://127.0.0.1:10808`
  - **SOCKS5 Proxy**: `socks5://127.0.0.1:10809`
- Mỗi khi gửi 1 request qua Proxy, chương trình sẽ tự động xoay vòng sang 1 địa chỉ IPv6 khác nhau trong danh sách!

---

### 3. [Chức năng 4] Kiểm Tra Kết Nối Internet Thực Tế
- Tool sẽ gửi request test qua các IP vừa gán đến dịch vụ kiểm tra IP quốc tế (`api64.ipify.org`, `ident.me`).
- Giúp bạn biết chính xác dải IPv6 của mình đã thông mạng ra thế giới hay chưa.

---

### 4. [Chức năng 2] Gỡ Bỏ IP Đã Gán (Dọn Dẹp Sạch Sẽ)
- Hiển thị danh sách các đợt IP đã gán kèm thời gian.
- Bạn có thể chọn gỡ bỏ 1 đợt cụ thể hoặc chọn **XÓA TẤT CẢ** chỉ với 1 thao tác.

---

## 🔌 Tích Hợp Vào Ứng Dụng & Tool

### AdsPower / GoLogin / Dolphin Anty / Multilogin:
- **Proxy Type**: `HTTP` hoặc `SOCKS5`
- **Proxy Host**: `127.0.0.1`
- **Proxy Port**: `10808` (cho HTTP) hoặc `10809` (cho SOCKS5)

### Python (Requests):
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

### cURL:
```cmd
curl -x http://127.0.0.1:10808 https://api64.ipify.org
```

---

## 💡 Lưu Ý Quan Trọng
- Để các IPv6 này truy cập được Internet ra thế giới, dải prefix `/64` bắt buộc phải là dải do nhà mạng (Viettel, VNPT, FPT...) cấp cho modem của bạn hoặc dải được định tuyến qua VPN/Tunnel (WireGuard, HE Tunnelbroker).
