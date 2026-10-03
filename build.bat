@echo off
echo ==================================================================
echo  BAT DAU BIEN DICH 2 BAN: WINDOWS VA LINUX
echo ==================================================================

if not exist "release\windows" mkdir "release\windows"
if not exist "release\linux" mkdir "release\linux"

echo [1/2] Dang bien dich ban Windows (amd64)...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o release\windows\ipv6-gen.exe .
copy /y "release\windows\ipv6-gen.exe" "ipv6-gen.exe" >nul
copy /y "release\windows\Chay_Tool_Windows.bat" "Chay_Tool_Windows.bat" >nul

echo [2/2] Dang bien dich ban Linux (amd64)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o release\linux\ipv6-gen-linux .
copy /y "release\linux\ipv6-gen-linux" "ipv6-gen-linux" >nul
copy /y "release\linux\run.sh" "run.sh" >nul
copy /y "release\linux\install-service.sh" "install-service.sh" >nul

echo ==================================================================
echo  HOAN TAT BIEN DICH CA 2 BAN!
echo  - Ban Windows: release\windows\ (Chay_Tool_Windows.bat, ipv6-gen.exe)
echo  - Ban Linux:   release\linux\   (run.sh, ipv6-gen-linux, install-service.sh)
echo ==================================================================
pause
