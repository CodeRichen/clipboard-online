# clipboard-online iPhone 捷徑修復指南

## 問題分析
根據檢測，clipboard-online 服務正常運行，但 iPhone 捷徑沒有發送必需的 API 版本標頭。

## 解決方案

### 步驟 1：下載最新版捷徑
請重新下載官方提供的最新版 iPhone 捷徑：

**複製捷徑（Copy）**：
https://www.icloud.com/shortcuts/f463a1e431c94c60b8a5c65305eb819f

**貼上捷徑（Paste）**：
https://www.icloud.com/shortcuts/90e7a2af70df4707a17dece8c263afc5

### 步驟 2：檢查捷徑設定
開啟 iPhone 捷徑應用程式，編輯下載的捷徑，確保 HTTP 請求包含以下標頭：

```
X-API-Version: 1
X-Client-Name: iPhone
```

### 步驟 3：設定 IP 地址
在捷徑中設定您的電腦 IP 地址：
- 電腦 IP：192.168.0.[您的電腦IP]
- 埠號：8086
- 完整網址：http://192.168.0.[您的電腦IP]:8086

### 步驟 4：驗證設定
1. 確認 Windows 防火牆允許埠號 8086
2. 確認電腦和 iPhone 在同一個區域網路
3. 在 iPhone 上測試捷徑

### 故障排解

**如果仍然無法使用：**
1. 檢查電腦 IP 地址：在 Windows 上執行 `ipconfig`
2. 測試網路連通性：在 iPhone Safari 開啟 http://[電腦IP]:8086
3. 檢查防火牆設定：確保允許埠號 8086
4. 重啟 clipboard-online 應用程式

**當前服務狀態：**
- ✅ clipboard-online 正在執行
- ✅ API 功能正常
- ✅ 埠號 8086 可用
- ❌ iPhone 捷徑缺少 X-API-Version 標頭

## 技術細節

根據測試結果：
- 沒有 API 版本標頭的請求會返回 HTTP 400 錯誤
- 包含正確標頭的請求能成功獲取剪貼簿內容
- 服務器要求 X-API-Version 值必須為 "1"

請按照上述步驟重新設定 iPhone 捷徑即可解決問題。