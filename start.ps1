# clipboard-online 啟動腳本

Write-Host "正在啟動 clipboard-online..." -ForegroundColor Green

# 設定程式路徑
$currentDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = Join-Path $currentDir "release\clipboard-online.exe"

Write-Host "程式路徑: $exePath"

# 檢查程式是否存在
if (-not (Test-Path $exePath)) {
    Write-Host "錯誤：找不到 clipboard-online.exe 檔案" -ForegroundColor Red
    exit 1
}

# 檢查程式是否已經在執行
$runningProcess = Get-Process -Name "clipboard-online" -ErrorAction SilentlyContinue
if ($runningProcess) {
    Write-Host "clipboard-online 已經在執行中" -ForegroundColor Yellow
    exit 0
}

# 啟動程式
Write-Host "正在啟動程式..."
$process = Start-Process -FilePath $exePath -WorkingDirectory (Split-Path $exePath) -WindowStyle Minimized -PassThru

# 等待程式啟動
Start-Sleep -Seconds 3

# 檢查是否成功啟動
$newProcess = Get-Process -Name "clipboard-online" -ErrorAction SilentlyContinue
if ($newProcess) {
    Write-Host "clipboard-online 已成功啟動！" -ForegroundColor Green
    Write-Host "程式現在在系統托盤中執行" -ForegroundColor Cyan
} else {
    Write-Host "程式可能沒有成功啟動，請檢查系統托盤" -ForegroundColor Yellow
}