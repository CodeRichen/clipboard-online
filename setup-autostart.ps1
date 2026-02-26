# clipboard-online 設定自動啟動腳本

Write-Host "=== clipboard-online 自動啟動設定 ===" -ForegroundColor Cyan
Write-Host ""

# 取得程式路徑
$currentDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = Join-Path $currentDir "release\clipboard-online.exe"

Write-Host "程式位置: $exePath"

# 檢查程式是否存在
if (-not (Test-Path $exePath)) {
    Write-Host "錯誤：找不到 clipboard-online.exe" -ForegroundColor Red
    Write-Host "請確認檔案位於 release 資料夾中" -ForegroundColor Yellow
    Read-Host "按 Enter 結束"
    exit 1
}

Write-Host ""
Write-Host "正在設定 Windows 自動啟動..." -ForegroundColor Green

# 設定註冊表鍵值
$regPath = "HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run"
$regName = "ClipboardOnline"

try {
    # 新增到自動啟動
    Set-ItemProperty -Path $regPath -Name $regName -Value $exePath -Force
    Write-Host "✓ 已成功設定開機自動啟動" -ForegroundColor Green
    
    # 驗證設定
    $regValue = Get-ItemProperty -Path $regPath -Name $regName -ErrorAction SilentlyContinue
    if ($regValue) {
        Write-Host "✓ 註冊表設定確認完成" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ 設定自動啟動時發生錯誤: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 手動啟動指引 ===" -ForegroundColor Cyan
Write-Host "1. 雙擊執行 release\clipboard-online.exe"
Write-Host "2. 程式會在系統托盤顯示圖示"  
Write-Host "3. 右鍵點擊托盤圖示可以設定選項"
Write-Host "4. 勾選 '開機啟動' 選項"
Write-Host ""
Write-Host "=== 故障排解 ===" -ForegroundColor Yellow
Write-Host "如果程式無法正常運行，請嘗試："
Write-Host "- 重新建編程式 (執行 .\build.ps1)"
Write-Host "- 檢查 Windows Defender 是否阻擋了程式"
Write-Host "- 以系統管理員身份執行程式"
Write-Host "- 檢查 config.json 設定檔案"
Write-Host ""

# 嘗試啟動程式
Write-Host "現在嘗試啟動程式..."
try {
    Start-Process -FilePath $exePath -WorkingDirectory (Split-Path $exePath)
    Write-Host "✓ 程式啟動命令已執行" -ForegroundColor Green
    Write-Host "請檢查系統托盤是否出現 clipboard-online 圖示" -ForegroundColor Cyan
} catch {
    Write-Host "✗ 啟動程式時發生錯誤: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Read-Host "設定完成，按 Enter 結束"