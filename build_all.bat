@echo off
echo ==================================================================
echo  BIEN DICH TOAN BO 2 PHIEN BAN: WINDOWS VA LINUX
echo ==================================================================

echo [1/2] Dang bien dich Windows (windows/ipv6-gen.exe)...
cd windows
go build -ldflags="-s -w" -o ipv6-gen.exe .
cd ..

echo [2/2] Dang bien dich Linux (linux/ipv6-gen-linux)...
cd linux
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o ipv6-gen-linux .
cd ..

echo ==================================================================
echo  THANH CONG!
echo  - Windows binary: windows/ipv6-gen.exe
echo  - Linux binary:   linux/ipv6-gen-linux
echo ==================================================================
