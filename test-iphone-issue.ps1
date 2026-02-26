# 測試 iPhone 捷徑問題的 PowerShell 腳本

Write-Host "=== clipboard-online iPhone 捷徑問題診斷 ===" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8086"

# 測試 1: 模擬 iPhone 發送錯誤的請求（沒有 Content-Type 標頭）
Write-Host "1. 模擬錯誤的iPhone請求（沒有X-Content-Type標頭）：" -ForegroundColor Yellow
try {
    $headers = @{
        'X-API-Version' = '1'
        'X-Client-Name' = 'iPhone'
        'Content-Type' = 'application/json'
    }
    $body = @{
        data = "這是一段普通文字，不是HTTP開頭"
    }
    $bodyJson = $body | ConvertTo-Json
    
    Write-Host "發送請求..." -ForegroundColor Gray
    $response = Invoke-WebRequest -Uri $baseUrl -Method POST -Headers $headers -Body $body -ErrorAction Stop
    Write-Host "✓ 成功 - 狀態碼: $($response.StatusCode)" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✗ 失敗 - 狀態碼: $statusCode" -ForegroundColor Red
    Write-Host "  ► 這展示了問題：沒有 X-Content-Type 標頭導致文字被當作文件處理" -ForegroundColor Yellow
}

Write-Host ""

# 測試 2: 發送正確的請求
Write-Host "2. 發送正確的iPhone請求（包含X-Content-Type: text）：" -ForegroundColor Yellow
try {
    $headers = @{
        'X-API-Version' = '1'
        'X-Client-Name' = 'iPhone'
        'X-Content-Type' = 'text'
        'Content-Type' = 'application/json'
    }
    $body = @{
        data = "這是正確處理的文字內容"
    }
    $bodyJson = $body | ConvertTo-Json
    
    Write-Host "發送請求..." -ForegroundColor Gray
    $response = Invoke-WebRequest -Uri $baseUrl -Method POST -Headers $headers -Body $body -ErrorAction Stop
    Write-Host "✓ 成功 - 狀態碼: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "  ► 現在檢查電腦剪貼簿應該有這段文字" -ForegroundColor Cyan
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✗ 失敗 - 狀態碼: $statusCode" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 診斷結果 ===" -ForegroundColor Green
Write-Host "問題原因：iPhone捷徑沒有在HTTP請求中包含 'X-Content-Type: text' 標頭"
Write-Host "解決方案：重新下載最新版iPhone捷徑或手動添加該標頭"
Write-Host ""
Write-Host "檢查日誌檔案以確認："
Write-Host "Get-Content 'D:\2.programm2\github-star\clipboard-online\release\log.txt' -Tail 10"