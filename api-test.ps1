# API 測試腳本

Write-Host "測試 clipboard-online API" -ForegroundColor Cyan

$baseUrl = "http://localhost:8086"

# 測試1: 沒有API版本標頭 (會失敗)
Write-Host "1. 測試沒有 API 版本標頭："
try {
    $response = Invoke-WebRequest -Uri $baseUrl -Method GET -ErrorAction Stop
    Write-Host "成功 - 狀態碼: $($response.StatusCode)" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "失敗 - 狀態碼: $statusCode" -ForegroundColor Red
    if ($statusCode -eq 400) {
        Write-Host "這是預期的錯誤：缺少 X-API-Version 標頭" -ForegroundColor Yellow
    }
}

Write-Host ""

# 測試2: 有正確的API版本標頭
Write-Host "2. 測試帶有正確 API 版本標頭："
try {
    $headers = @{
        'X-API-Version' = '1'
        'X-Client-Name' = 'PowerShell-Test'
    }
    $response = Invoke-WebRequest -Uri $baseUrl -Method GET -Headers $headers -ErrorAction Stop
    Write-Host "成功 - 狀態碼: $($response.StatusCode)" -ForegroundColor Green
    
    $content = $response.Content | ConvertFrom-Json
    Write-Host "剪貼簿類型: $($content.type)" -ForegroundColor Cyan
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "失敗 - 狀態碼: $statusCode" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 解決方案 ===" -ForegroundColor Green
Write-Host "iPhone 捷徑必須包含標頭: X-API-Version: 1"