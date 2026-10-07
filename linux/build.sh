#!/bin/bash
cd "$(dirname "$0")"
echo "Dang bien dich ipv6-gen-linux..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o ipv6-gen-linux .
chmod +x ipv6-gen-linux
echo "Hoan tat bien dich: ipv6-gen-linux"
