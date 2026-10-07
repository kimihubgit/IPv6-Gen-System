#!/bin/bash
# Cai dat proxy chay ngam nhu mot systemd service tren Linux
if [ "$EUID" -ne 0 ]; then
    echo "Vui long chay script voi quyen root: sudo ./install-service.sh"
    exit 1
fi

DIR="$(cd "$(dirname "$0")" && pwd)"
chmod +x "$DIR/ipv6-gen-linux"

read -p "Nhap card mang [eth0]: " IFACE
IFACE=${IFACE:-eth0}

read -p "Nhap IPv6 Prefix (/64): " PREFIX
if [ -z "$PREFIX" ]; then
    echo "Prefix khong duoc de trong!"
    exit 1
fi

read -p "So luong IP duy tri trong pool [200]: " POOL_SIZE
POOL_SIZE=${POOL_SIZE:-200}

read -p "Cong HTTP Proxy [10808]: " PORT
PORT=${PORT:-10808}

SERVICE_FILE="/etc/systemd/system/ipv6-proxy.service"

cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=IPv6 Rotating Proxy Service
After=network.target

[Service]
Type=simple
WorkingDirectory=$DIR
ExecStart=$DIR/ipv6-gen-linux -action dynamic-proxy -iface $IFACE -prefix $PREFIX -pool-size $POOL_SIZE -proxy-host 0.0.0.0 -proxy-port $PORT
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ipv6-proxy
systemctl restart ipv6-proxy

SOCKS_PORT=$((PORT + 1))
echo "=================================================================="
echo " Da cai dat va khoi chay dich vu ipv6-proxy thanh cong!"
echo " - HTTP Proxy:          http://0.0.0.0:$PORT (Tu Windows: http://<IP_VPS>:$PORT)"
echo " - SOCKS5 Proxy:        socks5://0.0.0.0:$SOCKS_PORT (Tu Windows: socks5://<IP_VPS>:$SOCKS_PORT)"
echo " - Kiem tra trang thai: systemctl status ipv6-proxy"
echo " - Xem log hoat dong:   journalctl -u ipv6-proxy -f"
echo " - Dung dich vu:        systemctl stop ipv6-proxy"
echo " - Luu y Firewall: Neu VPS bat firewall, hay mo port:"
echo "   sudo ufw allow $PORT/tcp && sudo ufw allow $SOCKS_PORT/tcp"
echo "=================================================================="
