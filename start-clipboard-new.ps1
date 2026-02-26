# clipboard-online 自動啟動腳本
# 此腳本會啟動 clipboard-online 程式並設定為背景執行

# 設定程式路徑
$currentDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = Join-Path $currentDir "release\clipboard-online.exe"

# 檢查程式是否存在
if (-not (Test-Path $exePath)) {
    Write-Host "錯誤：找不到 clipboard-online.exe 檔案" -ForegroundColor Red
    Write-Host "路徑：$exePath" -ForegroundColor Red
    Read-Host "按任意鍵繼續..."
    exit 1
}

# 檢查程式是否已經在執行
$runningProcess = Get-Process -Name "clipboard-online" -ErrorAction SilentlyContinue
if ($runningProcess) {
    Write-Host "clipboard-online 已經在執行中 (PID: $($runningProcess.Id))" -ForegroundColor Yellow
    exit 0
}

# 啟動程式
Write-Host "正在啟動 clipboard-online..." -ForegroundColor Green

try {
    # 使用 Start-Process 以最小化視窗啟動程式
    $process = Start-Process -FilePath $exePath -WorkingDirectory (Split-Path $exePath) -WindowStyle Minimized -PassThru
    
    # 等待一下讓程式完全啟動
    Start-Sleep -Seconds 2
    
    # 檢查程式是否成功啟動
    $newProcess = Get-Process -Name "clipboard-online" -ErrorAction SilentlyContinue
    if ($newProcess) {
        Write-Host "clipboard-online 已成功啟動 (PID: $($newProcess.Id))" -ForegroundColor Green
        Write-Host "程式現在在系統托盤中執行" -ForegroundColor Cyan
        Write-Host "可以在系統托盤中右鍵點擊圖示來設定開機啟動" -ForegroundColor Cyan
    } else {
        Write-Host "程式可能沒有成功啟動" -ForegroundColor Yellow
        Write-Host "請檢查系統托盤是否有 clipboard-online 圖示" -ForegroundColor Yellow
    }
}  
catch {
    Write-Host "錯誤：無法啟動程式" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}