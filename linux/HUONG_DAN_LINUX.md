# 🐧 Hướng Dẫn Triển Khai IPv6 Rotating Proxy Trên Linux VPS (ANY-IP + ndppd)

Tài liệu hướng dẫn chi tiết từ A-Z cách xây dựng hệ thống **IPv6 Rotating Proxy đỉnh cao trên Linux VPS**:
* **Không cần gán IP vào card mạng (0 IP rác trên card mạng `eth0`).**
* **Xoay vô hạn 18 tỷ tỷ IPv6** trong dải `/64` (mỗi request gửi từ máy Windows sẽ tự động mang một IPv6 ngẫu nhiên hoàn toàn mới).
* **Hoạt động 24/7** dưới dạng dịch vụ hệ thống `systemd`.
* Tương thích 100% với các trình duyệt Antidetect (**AdsPower, GoLogin, Hidemyacc, Dolphin**) và tool automation (**Python, Node.js, cURL**).

---

## 🧠 1. Nguyên Lý Hoạt Động (Cốt Lõi Kỹ Thuật)

Hầu hết các nhà cung cấp VPS (Linode, Hetzner, Vultr, OVH, FPT, CMC, Viettel...) cấp cho VPS một dải mạng IPv6 `/64` nhưng định tuyến thông qua giao thức **NDP (Neighbor Discovery Protocol)**:
1. Khi VPS gửi request ra ngoài bằng một IPv6 ngẫu nhiên (ví dụ `2403:6a40:0:15:xxxx:xxxx:xxxx:xxxx`), gói tin đến server đích bình thường.
2. Khi server đích phản hồi về, Router của nhà mạng sẽ hỏi trên mạng cục bộ: *"Ai đang giữ địa chỉ IPv6 này?"*
3. Nếu VPS **không gán IP** đó lên card mạng, Kernel sẽ không trả lời -> Router hủy gói tin -> Dẫn đến lỗi `i/o timeout` (503 Service Unavailable).

### 💡 Giải Pháp Tối Thượng: Kết Hợp `ndppd` + `ip_nonlocal_bind` + `Go Proxy`
* **`ndppd` (NDP Proxy Daemon)**: Đứng ở card `eth0` "nghe" bản tin hỏi của Router nhà mạng và **tự động trả lời `static`** xác nhận VPS sở hữu toàn bộ 18 tỷ tỷ IP trong dải `/64`.
* **`net.ipv6.ip_nonlocal_bind = 1`**: Cho phép ứng dụng Proxy Go bind vào bất kỳ IPv6 nào trong dải mà không cần IP đó phải tồn tại trên card mạng.
* **`ip -6 route add local <prefix> dev lo`**: Báo cho Kernel biết mọi gói tin thuộc dải `/64` này khi nhận về đều chuyển tiếp vào các socket nội bộ xử lý.
* **`ipv6-gen-linux (Go Proxy)`**: Lắng nghe cổng HTTP (`10808`) & SOCKS5 (`10809`), mỗi kết nối từ Windows tới sẽ tự động sinh 1 IPv6 ngẫu nhiên mới toanh để gửi ra Internet.

---

## 🚀 2. Các Bước Triển Khai Trên VPS (Làm 1 Lần Dùng Mãi Mãi)

### Bước 1: Kiểm tra card mạng và dải IPv6 của VPS
Đăng nhập SSH vào VPS bằng quyền `root`:
```bash
ip -6 addr show eth0
```
Bạn sẽ thấy thông tin dải IPv6 được cấp, ví dụ:
* Tên card mạng: `eth0`
* Dải Prefix `/64`: `2403:6a40:0:15::/64` *(4 nhóm số đầu tiên kết thúc bằng `::/64`)*

---

### Bước 2: Cài đặt và cấu hình `ndppd`
Chạy lệnh cài đặt `ndppd`:
```bash
apt-get update && apt-get install -y ndppd
```

Tạo file cấu hình `/etc/ndppd.conf` *(thay `eth0` và dải `/64` bằng thông số VPS của bạn)*:
```bash
cat <<EOF > /etc/ndppd.conf
proxy eth0 {
   router yes
   timeout 500
   ttl 30000
   rule 2403:6a40:0:15::/64 {
      static
   }
}
EOF
```
> **Lưu ý quan trọng**: Phải dùng `static` (không dùng `auto`) để `ndppd` luôn trả lời Router ngay lập tức cho các IP nội bộ.

Khởi động và kích hoạt `ndppd` tự chạy cùng hệ thống:
```bash
systemctl restart ndppd
systemctl enable ndppd
systemctl status ndppd
```
*(Trạng thái hiện màu xanh `active (running)` là thành công)*.

---

### Bước 3: Cấu hình Kernel Linux (`sysctl` & `ip route`)
Chạy các lệnh cấu hình Kernel:
```bash
# Bật cho phép bind IP ảo không cần gán card mạng
sysctl -w net.ipv6.ip_nonlocal_bind=1
sysctl -w net.ipv6.conf.all.forwarding=1

# Định tuyến dải IPv6 về loopback lo
ip -6 route add local 2403:6a40:0:15::/64 dev lo
```

#### Giữ cấu hình vĩnh viễn không bị mất khi Reboot VPS:
Thêm vào `/etc/sysctl.conf`:
```bash
echo "net.ipv6.ip_nonlocal_bind = 1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding = 1" >> /etc/sysctl.conf
sysctl -p
```

Thêm route tự động kích hoạt khi bật máy vào `/etc/rc.local` (hoặc crontab `@reboot`):
```bash
crontab -l | { cat; echo "@reboot ip -6 route add local 2403:6a40:0:15::/64 dev lo"; } | crontab -
```

---

### Bước 4: Test thử cơ chế ANY-IP trên VPS
Bịa ra một IPv6 bất kỳ trong dải (chưa từng gán vào máy) và test thử:
```bash
curl -6 --interface 2403:6a40:0:15:9999:aaaa:bbbb:cccc https://api64.ipify.org
```
Nếu màn hình in ra đúng IP `2403:6a40:0:15:9999:aaaa:bbbb:cccc`, hệ thống ANY-IP đã hoạt động hoàn hảo 100%!

---

### Bước 5: Cài đặt Proxy Go Service chạy ngầm 24/7
Tải mã nguồn từ GitHub hoặc copy thư mục `linux` lên VPS:
```bash
cd /root
git clone https://github.com/kimihubgit/IPv6-Gen-System.git
cd /root/IPv6-Gen-System/linux
chmod +x run.sh install-service.sh ipv6-gen-linux
```

Chạy script cài đặt dịch vụ tự động:
```bash
./install-service.sh
```
* **Card mạng [eth0]:** Bấm `Enter`
* **IPv6 Prefix (/64):** Dán `2403:6a40:0:15::/64`
* **Cổng HTTP Proxy [10808]:** Bấm `Enter`

Dịch vụ `ipv6-proxy.service` sẽ được cài đặt vào `systemd` và chạy ngầm mãi mãi.

---

### Bước 6: Mở cổng Firewall trên VPS
Mở cổng HTTP (`10808`) và SOCKS5 (`10809`) để máy tính từ xa kết nối được:
```bash
ufw allow 10808/tcp
ufw allow 10809/tcp
```
*(Nếu dùng AWS, Oracle Cloud, Google Cloud, Linode, Hetzner... nhớ vào trang web quản lý thêm Inbound Rule cho 2 port này)*.

---

## 🪟 3. Cách Kết Nối & Sử Dụng Từ Máy Tính Windows

Thông số Proxy của bạn:
* **Host / IP:** `<IP_V4_VPS>` (Ví dụ: `42.96.15.130`)
* **HTTP Proxy Port:** `10808`
* **SOCKS5 Proxy Port:** `10809`
* **Username / Password:** Để trống (Không cần đăng nhập)

### 1. Test trên PowerShell Windows (Xem IP xoay trực tiếp)
```powershell
# Test HTTP Proxy
curl.exe -s -x http://<IP_VPS>:10808 https://api64.ipify.org

# Test SOCKS5 Proxy
curl.exe -s -x socks5h://<IP_VPS>:10809 https://api64.ipify.org
```
*Mỗi lần bạn chạy lệnh, nó sẽ in ra một địa chỉ IPv6 ngẫu nhiên hoàn toàn mới!*

### 2. Sử dụng trong Antidetect Browser (AdsPower, GoLogin, Hidemyacc...)
1. Mở phần mềm quản lý Profile.
2. Thêm Proxy mới:
   - **Loại:** `HTTP` hoặc `SOCKS5`
   - **Host:** `<IP_VPS>`
   - **Port:** `10808` (hoặc `10809` cho SOCKS5)
3. Bấm **Check Proxy** -> Sẽ báo xanh và hiển thị IPv6 ngẫu nhiên.

### 3. Sử dụng trong Python Automation
```python
import requests

proxies = {
    'http': 'http://42.96.15.130:10808',
    'https': 'http://42.96.15.130:10808',
}

# Mỗi request gửi đi sẽ mang một IPv6 mới
for i in range(5):
    res = requests.get('https://api64.ipify.org?format=json', proxies=proxies, timeout=10)
    print(f"Lần {i+1}: {res.json()['ip']}")
```

---

## 🛠️ 4. Quản Lý & Lệnh Thường Dùng Trên VPS

### Quản lý dịch vụ Proxy:
```bash
# Xem trạng thái:
systemctl status ipv6-proxy

# Xem log xoay IP trực tiếp khi Windows gửi request:
journalctl -u ipv6-proxy -f

# Khởi động lại:
systemctl restart ipv6-proxy

# Dừng:
systemctl stop ipv6-proxy
```

### Quản lý dịch vụ `ndppd`:
```bash
# Xem trạng thái ndppd:
systemctl status ndppd

# Khởi động lại ndppd:
systemctl restart ndppd
```

---

## ❓ 5. Xử Lý Sự Cố (Troubleshooting)

| Lỗi thường gặp | Nguyên nhân | Cách khắc phục |
| :--- | :--- | :--- |
| **`curl: (7) Failed to connect... Could not connect to server`** | Dịch vụ proxy đang tắt hoặc chưa mở port Firewall. | Chạy `systemctl status ipv6-proxy` kiểm tra và mở port `ufw allow 10808/tcp`. |
| **`CONNECT tunnel failed, response 503` hoặc `i/o timeout`** | `ndppd` chưa chạy hoặc file config để `auto` thay vì `static`. | Sửa file `/etc/ndppd.conf` thành `rule ... { static }` và `systemctl restart ndppd`. |
| **Request trả về IPv4 của VPS thay vì IPv6** | Bảng định tuyến local chưa có route `dev lo` nên bị fallback. | Chạy `ip -6 route add local <prefix> dev lo`. |
| **Proxy báo `File exists` khi thêm route** | Route đã tồn tại sẵn trong bảng Kernel. | Hoàn toàn bình thường, bỏ qua. |
