# clipboard-online API 測試腳本

Write-Host "=== clipboard-online API 測試 ===" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8086"

# 測試沒有 API 版本標頭的請求（會返回 400）
Write-Host "1. 測試沒有 API 版本標頭的請求：" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri $baseUrl -Method GET -ErrorAction Stop
    Write-Host "✓ 成功 - 狀態碼: $($response.StatusCode)" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✗ 失敗 - 狀態碼: $statusCode (預期是 400)" -ForegroundColor Red
    if ($statusCode -eq 400) {
        Write-Host "  ► 這是預期的錯誤：缺少 X-API-Version 標頭" -ForegroundColor Yellow
    }
}

Write-Host ""

# 測試帶有正確 API 版本標頭的請求
Write-Host "2. 測試帶有正確 API 版本標頭的請求：" -ForegroundColor Yellow
try {
    $headers = @{
        'X-API-Version' = '1'
        'X-Client-Name' = 'PowerShell-Test'
    }
    $response = Invoke-WebRequest -Uri $baseUrl -Method GET -Headers $headers -ErrorAction Stop
    Write-Host "✓ 成功 - 狀態碼: $($response.StatusCode)" -ForegroundColor Green
    
    # 解析回應內容
    $content = $response.Content | ConvertFrom-Json
    Write-Host "  ► 類型: $($content.type)" -ForegroundColor Cyan
    if ($content.type -eq "text") {
        $dataPreview = if ($content.data.Length -gt 50) { $content.data.Substring(0, 50) + "..." } else { $content.data }
        Write-Host "  ► 內容: $dataPreview" -ForegroundColor Cyan
    }
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✗ 失敗 - 狀態碼: $statusCode" -ForegroundColor Red
    Write-Host "  ► 錯誤: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== iPhone 捷徑設定指引 ===" -ForegroundColor Cyan
Write-Host "iPhone 捷徑必須包含以下 HTTP 標頭："
Write-Host "• X-API-Version: 1" -ForegroundColor Yellow
Write-Host "• X-Client-Name: iPhone (可選)" -ForegroundColor Yellow
Write-Host ""
Write-Host "如果您的 config.json 有設定 authkey，還需要："
Write-Host "• X-Auth: [MD5 雜湊值]" -ForegroundColor Yellow
Write-Host ""

# 讀取當前設定
$configPath = "config.json"
if (Test-Path $configPath) {
    Write-Host "當前設定 (config.json)：" -ForegroundColor Green
    $config = Get-Content $configPath | ConvertFrom-Json
    Write-Host "• 埠號: $($config.port)" -ForegroundColor Cyan
    Write-Host "• 驗證金鑰: $(if($config.authkey -eq '') {'未設定'} else {'已設定'})" -ForegroundColor Cyan
    Write-Host "• 日誌等級: $($config.logLevel)" -ForegroundColor Cyan
} else {
    Write-Host "找不到 config.json 設定檔" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 解決方案 ===" -ForegroundColor Green
Write-Host "1. 重新下載最新版的 iPhone 捷徑"
Write-Host "2. 確保捷徑包含正確的 HTTP 標頭"
Write-Host "3. 檢查 IP 地址設定是否正確"
Write-Host "4. 如果問題持續，請檢查防火牆設定"