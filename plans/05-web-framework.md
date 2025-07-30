# Plan 05: Web 框架設置

## 目標
集成 Gin Web 框架，建立 REST API 基礎架構，為後續的 Web UI 功能做準備。

## 前置需求
- [Plan 04: 多日誌管理](04-multi-log-management.md) 完成

## 實施步驟

### 1. 更新 go.mod 依賴
```bash
go mod tidy
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
go get github.com/gin-contrib/static
```

### 2. 設計 API 架構 (TDD)
```go
type WebServer struct {
    router      *gin.Engine
    service     *Service
    logManager  *LogManager
    port        int
}

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### 3. 實現 REST API 端點

#### 腳本管理 API
```
GET    /api/scripts          - 獲取所有腳本列表
POST   /api/scripts          - 新增腳本
GET    /api/scripts/{name}   - 獲取特定腳本資訊
PUT    /api/scripts/{name}   - 更新腳本配置
DELETE /api/scripts/{name}   - 刪除腳本
POST   /api/scripts/{name}/run - 手動執行腳本
POST   /api/scripts/{name}/enable - 啟用腳本
POST   /api/scripts/{name}/disable - 停用腳本
```

#### 日誌管理 API
```
GET    /api/logs             - 獲取所有日誌 (支援查詢參數)
GET    /api/logs/{script}    - 獲取特定腳本日誌
DELETE /api/logs/{script}    - 清理特定腳本日誌
```

#### 系統資訊 API
```
GET    /api/status           - 系統狀態
GET    /api/config           - 獲取系統配置
PUT    /api/config           - 更新系統配置
```

### 4. 實現中間件
- CORS 支援
- 請求日誌記錄
- 錯誤處理
- 認證中間件 (可選)

### 5. 更新服務啟動邏輯
```go
func (s *Service) StartWithWeb() error {
    // 啟動背景腳本服務
    go s.runBackground()

    // 啟動 Web 服務器
    webServer := NewWebServer(s, s.logManager, s.config.WebPort)
    return webServer.Start()
}
```

### 6. 更新 CLI 介面
```bash
# 啟動服務 (包含 Web UI)
./run-script-service run --web

# 設置 Web 端口
./run-script-service set-web-port 8080

# 純背景模式 (無 Web UI)
./run-script-service run --background
```

## 驗收標準
- [ ] Web 服務可以正常啟動
- [ ] 所有 API 端點回應正確
- [ ] API 支援 JSON 格式請求/回應
- [ ] CORS 配置正確
- [ ] 錯誤處理統一且友善
- [ ] Web 服務與背景服務可以同時運行
- [ ] 所有 API 都有單元測試

## 相關檔案
- `web/server.go` (新增)
- `web/handlers.go` (新增)
- `web/middleware.go` (新增)
- `go.mod` (修改)
- `main.go` (修改)

## API 測試範例
```bash
# 測試腳本列表
curl http://localhost:8080/api/scripts

# 新增腳本
curl -X POST http://localhost:8080/api/scripts \
  -H "Content-Type: application/json" \
  -d '{"name":"test","path":"./test.sh","interval":60}'

# 執行腳本
curl -X POST http://localhost:8080/api/scripts/test/run
```

## 測試案例
- API 端點正確性
- 錯誤狀況處理
- 並發請求處理
- 大量日誌查詢性能

## 後續計劃
- [Plan 06: 基礎 Web UI](06-web-ui-basic.md)
