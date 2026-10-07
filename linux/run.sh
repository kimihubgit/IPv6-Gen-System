#!/bin/bash
# Script chạy tool IPv6 trên Linux (tự động cấp quyền và gọi sudo)
cd "$(dirname "$0")"

chmod +x ./ipv6-gen-linux

# Bật cấu hình sysctl cho phép nhiều IP
if [ "$EUID" -ne 0 ]; then
    echo "Dang khoi dong voi quyen sudo..."
    sudo sysctl -w net.ipv6.conf.all.max_addresses=0 >/dev/null 2>&1
    sudo ./ipv6-gen-linux "$@"
else
    sysctl -w net.ipv6.conf.all.max_addresses=0 >/dev/null 2>&1
    ./ipv6-gen-linux "$@"
fi
