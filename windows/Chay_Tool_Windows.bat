@echo off
chcp 65001 >nul
title Windows IPv6 Rotator & Generator System

:: Kiem tra quyen Administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ==================================================================
    echo  [!] Dang yeu cau quyen Administrator (UAC)...
    echo ==================================================================
    powershell -NoProfile -Command "Start-Process '%~dp0ipv6-gen.exe' -WorkingDirectory '%~dp0' -Verb RunAs"
    exit /b
)

:: Neu da co quyen Admin thi chay truc tiep
cd /d "%~dp0"
ipv6-gen.exe
pause
