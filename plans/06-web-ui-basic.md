# Plan 06: 基礎 Web UI

## 目標
建立簡潔的 Web UI 界面，提供腳本管理、日誌查看和系統監控功能。

## 前置需求
- [Plan 05: Web 框架設置](05-web-framework.md) 完成

## 實施步驟

### 1. 設計 UI 架構
- 使用 Vanilla JavaScript + HTML/CSS (無需複雜框架)
- 響應式設計 (支援桌面和行動裝置)
- 單頁應用 (SPA) 設計

### 2. 建立靜態檔案結構
```
web/
├── static/
│   ├── index.html
│   ├── css/
│   │   ├── main.css
│   │   └── components.css
│   ├── js/
│   │   ├── app.js
│   │   ├── api.js
│   │   └── components.js
│   └── assets/
│       └── favicon.ico
```

### 3. 實現主要頁面組件

#### Dashboard (首頁)
- 系統狀態概覽
- 活躍腳本數量
- 最近執行結果
- 系統資源使用情況

#### 腳本管理頁面
- 腳本列表 (表格形式)
- 新增/編輯/刪除腳本
- 啟用/停用腳本
- 手動執行腳本
- 腳本狀態指示

#### 日誌查看頁面
- 腳本選擇下拉選單
- 日誌列表 (分頁顯示)
- 日誌搜尋和過濾
- 即時日誌更新 (WebSocket 或輪詢)
- 日誌匯出功能

#### 設定頁面
- 系統配置編輯
- Web 端口設定
- 日誌保留設定

### 4. 實現前端 JavaScript 功能
```javascript
// API 客戶端
class APIClient {
    async getScripts() { ... }
    async addScript(script) { ... }
    async runScript(name) { ... }
    async getLogs(scriptName, options) { ... }
}

// 組件管理
class ComponentManager {
    renderScriptList(scripts) { ... }
    renderLogViewer(logs) { ... }
    showNotification(message, type) { ... }
}
```

### 5. 樣式設計
- 使用 CSS Grid/Flexbox 布局
- 現代扁平化設計風格
- 深色/淺色主題支援
- 一致的顏色配置和字體

### 6. 更新 Web 服務器
```go
// 靜態檔案服務
router.Static("/static", "./web/static")
router.GET("/", func(c *gin.Context) {
    c.File("./web/static/index.html")
})

// WebSocket 支援 (即時日誌)
router.GET("/ws/logs", websocketHandler)
```

## 頁面功能規劃

### 導覽列
- Dashboard
- Scripts (腳本管理)
- Logs (日誌)
- Settings (設定)

### 腳本管理功能
- ✅ 查看所有腳本狀態
- ✅ 新增腳本 (表單)
- ✅ 編輯腳本配置
- ✅ 刪除腳本 (確認對話框)
- ✅ 手動執行腳本
- ✅ 啟用/停用腳本

### 日誌查看功能
- ✅ 選擇腳本查看日誌
- ✅ 日誌分頁載入
- ✅ 日誌搜尋 (關鍵字、時間範圍)
- ✅ 自動重新整理
- ✅ 日誌匯出 (JSON/CSV)

## 驗收標準
- [ ] Web UI 可以透過瀏覽器正常存取
- [ ] 所有腳本管理操作都可以透過 UI 完成
- [ ] 日誌可以即時查看和搜尋
- [ ] UI 在不同螢幕尺寸下都能正常顯示
- [ ] 操作回饋清楚 (成功/錯誤訊息)
- [ ] UI 回應速度良好 (< 2 秒)

## 相關檔案
- `web/static/index.html` (新增)
- `web/static/css/main.css` (新增)
- `web/static/js/app.js` (新增)
- `web/server.go` (修改 - 新增靜態檔案路由)

## UI 原型圖
```
┌─────────────────────────────────────┐
│ Run Script Service        [Settings]│
├─────────────────────────────────────┤
│ Dashboard | Scripts | Logs          │
├─────────────────────────────────────┤
│                                     │
│  ┌─ Scripts ──────────────────────┐ │
│  │ Name    Status   Last Run  [▶] │ │
│  │ backup  ✅ Running  2m ago  [⏸] │ │
│  │ cleanup ❌ Stopped  1h ago  [▶] │ │
│  │                    [+ Add New] │ │
│  └─────────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

## 測試案例
- 所有頁面載入正常
- CRUD 操作功能正確
- 錯誤處理顯示適當
- 行動裝置相容性

## 後續計劃
- [Plan 07: Web 編輯功能](07-web-editing.md)