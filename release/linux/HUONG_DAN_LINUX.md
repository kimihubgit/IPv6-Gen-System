# 🐧 Hướng Dẫn Sử Dụng Tool IPv6 Trên Linux / VPS

Bản thực thi độc lập (**Statically Linked Binary**), tương thích 100% với mọi hệ điều hành Linux (Ubuntu, Debian, CentOS, AlmaLinux, Rocky Linux, Alpine, Fedora, Arch...).

---

## 🚀 Cách Cài Đặt & Chạy Nhanh Nhất (Chỉ 2 Bước)

### Bước 1: Tải thư mục `release/linux` lên VPS / Server
Bạn có thể dùng SCP, SFTP (FileZilla / WinSCP) hoặc git clone:
```bash
# Di chuyển vào thư mục tool
cd release/linux
```

### Bước 2: Chạy công cụ (1-Click)
```bash
chmod +x run.sh
./run.sh
```
> **Ghi chú**: Script `run.sh` sẽ tự động cấp quyền thực thi cho `ipv6-gen-linux`, tự động gọi `sudo` nếu bạn đang dùng tài khoản thường, và tự động cấu hình `sysctl` để gán hàng ngàn IP không bị giới hạn.

---

## 📋 Hướng Dẫn Sử Dụng Menu Tương Tác

Giao diện Menu trực quan trên Linux:
```text
==================================================================
        🌐 LINUX IPV6 ROTATOR & GENERATOR TOOL (GO) 🌐           
   Sinh & Gán hàng loạt IPv6 vào Card Mạng - Tích hợp Rotating Proxy
==================================================================
 [✓] Quyền thực thi: Root / Sudo (Đủ quyền cấu hình card mạng)
------------------------------------------------------------------
  [1] ⚡ Sinh & Gán danh sách IPv6 vào Card mạng
  [2] 🧹 Gỡ bỏ IPv6 đã gán (Xem lịch sử & Clean up)
  [3] 🚀 Khởi chạy Local Rotating Proxy Server (HTTP / SOCKS5)
  [4] 🧪 Kiểm tra kết nối Internet thực tế của IPv6
  [5] 📄 Chỉ sinh danh sách IPv6 ra file .txt (Không can thiệp card mạng)
  [0] 🚪 Thoát
```

### 1. Gán IPv6 vào Card mạng Linux (Mục 1)
1. **Chọn card mạng**: Tool sẽ liệt kê các card mạng (ví dụ: `eth0`, `ens3`, `enp1s0`, `enp3s0`). Bạn chọn card đang nối mạng Internet chính.
2. **Nhập IPv6 Prefix**: Tool tự nhận diện dải `/64` của VPS (nếu VPS đã được cấp phát dải IPv6) và gợi ý sẵn. Bạn chỉ cần bấm `Enter` hoặc nhập dải của bạn (vd: `2a01:4f8:xxxx:xxxx::/64`).
3. **Nhập số lượng IP**: ví dụ `100`, `500` hoặc `1000`.
4. **Chế độ**: Chọn `[1] Random` (khuyên dùng).
5. **Hoàn tất**: Tool gán đồng thời cực nhanh bằng lệnh kernel `ip -6 addr add`, đồng thời lưu file text và lịch sử dọn dẹp.

---

### 2. Chạy Proxy Nền 24/7 (Systemd Service - Khuyên Dùng Cho VPS)
Khi chạy trên VPS, bạn thường không muốn giữ cửa sổ SSH mở liên tục. Chúng tôi đã chuẩn bị sẵn script cài đặt dịch vụ nền:

```bash
sudo ./install-service.sh
```

Script sẽ:
- Hỏi dải IP / số lượng IP muốn tạo (nếu chưa có).
- Hỏi cổng HTTP Proxy (mặc định: `10808`).
- Tự động tạo dịch vụ systemd `/etc/systemd/system/ipv6-proxy.service`.
- Tự động kích hoạt chạy cùng hệ thống (Auto-start khi VPS reboot) và khởi chạy ngay lập tức!

#### Quản lý dịch vụ Proxy:
```bash
# Xem trạng thái hoạt động:
systemctl status ipv6-proxy

# Xem log trực tiếp khi Proxy xoay IP:
journalctl -u ipv6-proxy -f

# Dừng Proxy:
sudo systemctl stop ipv6-proxy

# Khởi động lại:
sudo systemctl restart ipv6-proxy
```

---

## 🧪 Kiểm Tra Kết Nối Proxy

### Kiểm tra xoay IP qua curl:
```bash
# Gọi 3 lần liên tiếp xem IP có tự xoay không:
curl -x http://127.0.0.1:10808 https://api64.ipify.org ; echo ""
curl -x http://127.0.0.1:10808 https://api64.ipify.org ; echo ""
curl -x http://127.0.0.1:10808 https://api64.ipify.org ; echo ""
```

---

## 🌐 Dùng Proxy Từ Máy Tính Cá Nhân (Client) Về VPS

Nếu Proxy đang chạy trên VPS (ví dụ IP VPS là `1.2.3.4`):
### Cách 1: Sử dụng SSH Tunnel (An toàn nhất, không cần mở port)
Trên máy tính cá nhân của bạn, chạy lệnh sau:
```bash
ssh -L 10808:127.0.0.1:10808 root@IP_VPS_CUA_BAN
```
Sau đó trên máy tính của bạn, nhập proxy là `127.0.0.1:10808` vào trình duyệt hoặc AdsPower/GoLogin. Toàn bộ lưu lượng sẽ được mã hóa gửi sang VPS và xoay IPv6 ra ngoài Internet!

---

## 🛠️ Lệnh Dòng Lệnh Không Cần Menu (CLI Headless)

Nếu bạn muốn viết script tự động hóa:
```bash
# 1. Gán 100 IP vào eth0 từ dải:
sudo ./ipv6-gen-linux -action add -iface eth0 -prefix "2001:db8:1234::/64" -count 100 -out my_ips.txt

# 2. Khởi chạy Proxy server:
sudo ./ipv6-gen-linux -action proxy -out my_ips.txt -proxy-port 10808

# 3. Chỉ sinh 200 IP ra file text:
./ipv6-gen-linux -action generate -prefix "2001:db8:1234::/64" -count 200 -out ips.txt
```
