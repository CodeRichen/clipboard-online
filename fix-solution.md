# clipboard-online 手機複製問題解決方案

## 🔍 問題分析

根據您提供的日誌，問題出現在兩個方面：

### 1. 臨時檔案刪除錯誤
```
time="2026-02-26T12:02:05+08:00" level=warning msg="failed to delete specify path" 
delPath="D:\\2.programm2\\github-star\\clipboard-online\\release\\temp\\Clipboard 2026年2月26日 上午11.59.txt" 
error="remove D:\\2.programm2\\github-star\\clipboard-online\\release\\temp\\Clipboard 2026年2月26日 上午11.59.txt: The system cannot find the file specified."
```

### 2. iPhone 捷徑內容類型判斷問題
當手機複製的文字不是 HTTP 開頭時，iPhone 捷徑可能沒有正確設定 `X-Content-Type: text` 標頭，導致服務器將文本內容錯誤地當作文件處理。

## 🛠️ 解決方案

### 方案 1：修復 iPhone 捷徑設定（推薦）

1. **重新下載最新版捷徑**：
   - 複製捷徑：https://www.icloud.com/shortcuts/f463a1e431c94c60b8a5c65305eb819f
   - 貼上捷徑：https://www.icloud.com/shortcuts/90e7a2af70df4707a17dece8c263afc5

2. **檢查捷徑設定**：
   開啟iPhone捷徑應用，編輯 "clipboard-online 貼上" 捷徑：
   
   **必需的 HTTP 標頭**：
   ```
   X-API-Version: 1
   X-Content-Type: text    ← 重要！確保這個標頭設定為 "text"
   X-Client-Name: iPhone
   ```

   **網址設定**：
   ```
   http://192.168.0.201:8086
   或
   http://192.168.0.181:8086
   ```

3. **測試步驟**：
   - 在 iPhone 上複製任意文字（非 HTTP 開頭）
   - 執行 "clipboard-online 貼上" 捷徑
   - 檢查電腦剪貼簿是否正確接收

### 方案 2：手動修改捷徑（進階）

如果上述方案無效，請手動修改捷徑：

1. 開啟捷徑應用 → 編輯 "clipboard-online 貼上" 捷徑
2. 找到 "取得網頁內容" 或 "HTTP請求" 動作
3. 確保標頭設定包含：
   ```
   X-API-Version: 1
   X-Content-Type: text
   X-Client-Name: iPhone
   ```
4. 確保請求方法為 POST
5. 確保請求主體格式為：
   ```json
   {
     "data": "剪貼簿內容"
   }
   ```

### 方案 3：清除臨時檔案（解決警告）

刪除有問題的臨時檔案：

```powershell
Remove-Item -Path "D:\2.programm2\github-star\clipboard-online\release\temp\*" -Force -ErrorAction SilentlyContinue
```

或者，設定程式保留歷史檔案，修改 `config.json`：

```json
{
  "port": "8086",
  "authkey": "",
  "authkeyExpiredTimeout": 30,
  "logLevel": "warning",
  "tempDir": "./temp",
  "reserveHistory": true,     ← 改為 true
  "notify": {
    "copy": true,
    "paste": true
  }
}
```

## 📱 iPhone 捷徑檢查清單

請確認您的 iPhone 捷徑包含以下設定：

- ✅ **X-API-Version**: `1`
- ✅ **X-Content-Type**: `text`（對於文字內容）
- ✅ **X-Client-Name**: `iPhone`
- ✅ **請求方法**: `POST`
- ✅ **網址**: `http://[您的電腦IP]:8086`
- ✅ **請求主體**: `{"data": "剪貼簿內容"}`

## 🔧 故障排解

### 測試連通性
在 iPhone Safari 開啟：`http://[您的電腦IP]:8086`
- 如果顯示錯誤頁面但有回應 → 連通性正常
- 如果無法載入 → 檢查防火牆和網路設定

### 檢查日誌
觀察 `release/log.txt` 檔案：
- **HTTP 200**：成功
- **HTTP 400**：缺少 API 版本標頭或內容類型錯誤
- **HTTP 403**：驗證失敗（如有設定 authkey）

### 常見問題
1. **文字無法貼上**：檢查 `X-Content-Type: text` 標頭
2. **檔案無法處理**：確認臨時目錄可寫入
3. **連線失敗**：檢查電腦 IP 和防火牆設定

## 📞 技術支援

如問題持續存在，請提供：
1. iPhone 捷徑的螢幕截圖（顯示 HTTP 請求設定）
2. `release/log.txt` 的最新日誌
3. 您的網路設定（電腦和 iPhone 的 IP）

重點是確保 iPhone 捷徑正確設定 `X-Content-Type: text` 標頭來處理純文字內容。