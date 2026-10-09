#!/bin/bash
# Cai dat proxy chay ngam nhu mot systemd service tren Linux
if [ "$EUID" -ne 0 ]; then
    echo "Vui long chay script voi quyen root: sudo ./install-service.sh"
    exit 1
fi

DIR="$(cd "$(dirname "$0")" && pwd)"
chmod +x "$DIR/ipv6-gen-linux"
mkdir -p "$DIR/data"

echo "=================================================================="
echo "    CHON CHE DO CAI DAT DICH VU NEN (SYSTEMD 24/7):"
echo "  [1] Web Dashboard Hub (Quan ly ban/cho thue proxy da cong, GB, han dung)"
echo "  [2] Single Proxy Don Gian (Chi chay 1 cong 10808 / 10809)"
echo "=================================================================="
read -p "Chon che do [1]: " MODE
MODE=${MODE:-1}

read -p "Nhap card mang [eth0]: " IFACE
IFACE=${IFACE:-eth0}

read -p "Nhap IPv6 Prefix (/64): " PREFIX
if [ -z "$PREFIX" ]; then
    echo "Prefix khong duoc de trong!"
    exit 1
fi

SERVICE_FILE="/etc/systemd/system/ipv6-proxy.service"

if [ "$MODE" -eq 1 ]; then
    read -p "Cong Web Dashboard [9090]: " WEB_PORT
    WEB_PORT=${WEB_PORT:-9090}

    # Tối ưu hóa kernel Linux cho proxy luồng cao
    cat << 'SYSCTL_EOF' > /etc/sysctl.d/99-ipv6-proxy.conf
fs.file-max = 2097152
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv6.neigh.default.gc_thresh1 = 4096
net.ipv6.neigh.default.gc_thresh2 = 8192
net.ipv6.neigh.default.gc_thresh3 = 16384
net.ipv6.ip_nonlocal_bind = 1
net.ipv6.conf.all.forwarding = 1
SYSCTL_EOF
    sysctl --system >/dev/null 2>&1

    cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=IPv6 Proxy Hub - Web Dashboard & Multi-Port Server
After=network.target

[Service]
Type=simple
WorkingDirectory=$DIR
ExecStart=$DIR/ipv6-gen-linux -action server -iface $IFACE -prefix $PREFIX -web-port $WEB_PORT -data $DIR/data/proxies.json
Restart=always
RestartSec=5
LimitNOFILE=1048576
LimitNPROC=512000
TasksMax=infinity

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable ipv6-proxy
    systemctl restart ipv6-proxy

    echo "=================================================================="
    echo " Da cai dat va khoi chay IPv6 Proxy Hub thanh cong!"
    echo " - Web Dashboard:       http://<IP_VPS>:$WEB_PORT"
    echo " - Tai khoan mac dinh:  admin / admin123"
    echo " - Kiem tra trang thai: systemctl status ipv6-proxy"
    echo " - Xem log hoat dong:   journalctl -u ipv6-proxy -f"
    echo " - Luu y Firewall: Mo port Web va cac port Proxy:"
    echo "   sudo ufw allow $WEB_PORT/tcp"
    echo "   sudo ufw allow 10000:10100/tcp"
    echo "=================================================================="

else
    read -p "Cong HTTP Proxy [10808]: " PORT
    PORT=${PORT:-10808}

    cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=IPv6 Any-IP Rotating Proxy Service (Single Port)
After=network.target

[Service]
Type=simple
WorkingDirectory=$DIR
ExecStart=$DIR/ipv6-gen-linux -action any-proxy -iface $IFACE -prefix $PREFIX -proxy-host 0.0.0.0 -proxy-port $PORT
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
    echo " - Luu y Firewall: Neu VPS bat firewall, hay mo port:"
    echo "   sudo ufw allow $PORT/tcp && sudo ufw allow $SOCKS_PORT/tcp"
    echo "=================================================================="
fi
