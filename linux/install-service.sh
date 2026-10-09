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
