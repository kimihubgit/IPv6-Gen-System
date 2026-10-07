@echo off
echo ==================================================================
echo  DANG BIEN DICH BAN WINDOWS (ipv6-gen.exe)
echo ==================================================================
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ipv6-gen.exe .
echo ==================================================================
echo  DA BIEN DICH XONG: ipv6-gen.exe
echo ==================================================================
pause
