# Plan 07: Web 編輯功能

## 目標
實現進階 Web 編輯功能，包括腳本內容編輯、配置檔案編輯和線上程式碼編輯器。

## 前置需求
- [Plan 06: 基礎 Web UI](06-web-ui-basic.md) 完成

## 實施步驟

### 1. 集成程式碼編輯器
- 選擇 CodeMirror 或 Monaco Editor
- 支援語法高亮 (Shell, JSON, YAML)
- 支援自動完成和語法檢查
- 支援檔案比較 (diff view)

### 2. 擴展 API 端點

#### 檔案操作 API
```
GET    /api/files/{path}     - 讀取檔案內容
PUT    /api/files/{path}     - 寫入檔案內容
POST   /api/files/validate   - 驗證檔案語法
GET    /api/files/diff       - 比較檔案差異
```

#### 系統操作 API
```
POST   /api/system/backup    - 建立系統備份
POST   /api/system/restore   - 還原系統備份
GET    /api/system/health    - 系統健康檢查
```

### 3. 實現檔案編輯功能

#### 腳本檔案編輯器
```javascript
class ScriptEditor {
    constructor(container) {
        this.editor = CodeMirror(container, {
            mode: 'shell',
            theme: 'monokai',
            lineNumbers: true,
            autoCloseBrackets: true,
            matchBrackets: true
        });
    }

    async loadScript(scriptPath) { ... }
    async saveScript(scriptPath, content) { ... }
    validateSyntax() { ... }
}
```

#### 配置檔案編輯器
- JSON 格式驗證
- 即時語法檢查
- 配置項目說明提示
- 安全性驗證 (防止無效配置)

### 4. 實現進階編輯功能

#### 批次操作
- 批次啟用/停用腳本
- 批次修改執行間隔
- 批次匯出/匯入腳本

#### 版本控制
- 編輯歷史記錄
- 變更追蹤和比較
- 一鍵還原到先前版本

#### 範本管理
- 預設腳本範本
- 自訂範本建立
- 範本分享和匯入

### 5. 安全性考慮

#### 檔案存取控制
```go
type FileAccessControl struct {
    allowedPaths []string
    deniedPaths  []string
}

func (fac *FileAccessControl) IsPathAllowed(path string) bool {
    // 實現路徑白名單檢查
    // 防止存取系統敏感檔案
}
```

#### 輸入驗證
- 腳本內容安全檢查
- 檔案路徑驗證
- 執行權限檢查

### 6. 實現實時協作功能 (可選)
- WebSocket 連線管理
- 多用戶編輯衝突解決
- 即時變更同步

## UI 功能擴展

### 腳本編輯器頁面
```
┌─────────────────────────────────────┐
│ Script Editor: backup.sh      [Save]│
├─────────────────────────────────────┤
│ #!/bin/bash                   │1    │
│ # Backup script               │2    │
│ echo "Starting backup..."     │3    │
│ rsync -av /data/ /backup/     │4    │
│ echo "Backup completed"       │5    │
├─────────────────────────────────────┤
│ [Validate] [Preview] [History]      │
└─────────────────────────────────────┘
```

### 配置編輯器頁面
- 視覺化配置編輯器
- JSON/YAML 切換檢視
- 配置驗證和預覽
- 匯入/匯出功能

### 範本管理頁面
- 範本庫瀏覽
- 範本預覽和編輯
- 範本套用到新腳本

## 驗收標準
- [ ] 可以透過 Web UI 編輯腳本檔案
- [ ] 程式碼編輯器功能完整 (語法高亮、自動完成等)
- [ ] 配置檔案可以安全編輯和驗證
- [ ] 檔案變更有歷史記錄和還原功能
- [ ] 編輯操作有適當的權限控制
- [ ] 批次操作功能正常運作
- [ ] 所有編輯功能都有錯誤處理

## 相關檔案
- `web/static/js/editor.js` (新增)
- `web/static/css/editor.css` (新增)
- `web/handlers_files.go` (新增)
- `service/file_manager.go` (新增)
- `service/template_manager.go` (新增)

## 安全性檢查清單
- [ ] 檔案路徑驗證 (防止 path traversal)
- [ ] 腳本內容檢查 (防止惡意程式碼)
- [ ] 執行權限控制
- [ ] 輸入資料驗證
- [ ] 錯誤訊息不洩漏敏感資訊

## 測試案例
- 程式碼編輯器功能測試
- 檔案儲存和載入測試
- 權限控制測試
- 大檔案處理測試
- 並發編輯測試

## 效能考慮
- 大檔案分塊載入
- 編輯器延遲載入
- 自動儲存功能
- 變更節流 (throttling)

## 後續擴展
- 整合 Git 版本控制
- 支援更多程式語言
- 外掛系統架構
- 多租戶支援
