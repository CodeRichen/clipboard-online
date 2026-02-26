# Test script for iPhone shortcut

Write-Host "Testing clipboard-online text handling" -ForegroundColor Cyan

$url = "http://localhost:8086"
$headers = @{
    'X-API-Version' = '1'
    'X-Client-Name' = 'iPhone-Test'
    'X-Content-Type' = 'text'
    'Content-Type' = 'application/json'
}

$body = '{"data":"Test text from iPhone - not HTTP URL"}'

try {
    $response = Invoke-WebRequest -Uri $url -Method POST -Headers $headers -Body $body -UseBasicParsing
    Write-Host "SUCCESS! Status Code: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Check your computer clipboard for the test text" -ForegroundColor Cyan
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "SOLUTION: Ensure iPhone shortcut includes X-Content-Type: text header" -ForegroundColor Yellow