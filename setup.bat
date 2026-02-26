@echo off
echo =================================
echo clipboard-online 自動啟動設定
echo =================================
echo.

set "exePath=%~dp0release\clipboard-online.exe"
echo 程式位置: %exePath%

if not exist "%exePath%" (
    echo 錯誤：找不到 clipboard-online.exe 檔案
    echo 請確認檔案位於 release 資料夾中
    pause
    exit /b 1
)

echo.
echo 正在設定 Windows 自動啟動...

:: 設定註冊表自動啟動
reg add "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run" /v "ClipboardOnline" /t REG_SZ /d "%exePath%" /f >nul 2>&1

if %errorlevel% equ 0 (
    echo ✓ 已成功設定開機自動啟動
) else (
    echo ✗ 設定自動啟動時發生錯誤
)

echo.
echo 正在啟動程式...
start "" "%exePath%"

echo ✓ 程式啟動命令已執行
echo.
echo =================================
echo 使用說明：
echo 1. 程式會在系統托盤顯示圖示
echo 2. 右鍵點擊托盤圖示可以設定選項  
echo 3. 勾選 '開機啟動' 選項
echo =================================
echo.
echo 如果程式無法正常運行，請：
echo - 檢查系統托盤是否有程式圖示
echo - 重新建編程式
echo - 檢查防毒軟體是否阻擋
echo.
pause