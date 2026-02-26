# iPhone 捷徑問題測試腳本

Write-Host "測試 clipboard-online 是否能處理文字請求" -ForegroundColor Cyan

$baseUrl = "http://localhost:8086"

# 測試正確的文字請求
Write-Host "發送正確的文字請求..." -ForegroundColor Yellow
$headers = @{
    'X-API-Version' = '1'
    'X-Client-Name' = 'iPhone-Test'
    'X-Content-Type' = 'text'
    'Content-Type' = 'application/json'
}

$bodyContent = @{ data = "測試文字：這不是HTTP開頭的內容" }
$jsonBody = $bodyContent | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri $baseUrl -Method POST -Headers $headers -Body $jsonBody -UseBasicParsing
    Write-Host "成功！狀態碼: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "請檢查電腦剪貼簿是否出現測試文字" -ForegroundColor Cyan
} catch {
    Write-Host "失敗：$($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "解決方案：確保iPhone捷徑包含 X-Content-Type: text 標頭" -ForegroundColor Yellow