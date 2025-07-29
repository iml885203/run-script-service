# Plan 03: 多腳本支援

## 目標
擴展服務以支援同時管理和執行多個腳本，每個腳本可以有獨立的配置和執行週期。

## 前置需求
- [Plan 01: 單元測試基礎設施](01-unit-testing.md) 完成
- [Plan 02: TDD 開發流程](02-tdd-workflow.md) 完成

## 實施步驟

### 1. 重新設計配置結構 (TDD)
```go
type ScriptConfig struct {
    Name        string `json:"name"`
    Path        string `json:"path"`
    Interval    int    `json:"interval"`    // 秒
    Enabled     bool   `json:"enabled"`
    MaxLogLines int    `json:"max_log_lines"`
    Timeout     int    `json:"timeout"`     // 秒，0 表示無限制
}

type ServiceConfig struct {
    Scripts []ScriptConfig `json:"scripts"`
    WebPort int           `json:"web_port"`
}
```

### 2. 實現腳本管理器 (TDD)
```go
type ScriptManager struct {
    scripts map[string]*ScriptRunner
    config  *ServiceConfig
}

type ScriptRunner struct {
    config   ScriptConfig
    ticker   *time.Ticker
    cancel   context.CancelFunc
    executor *ScriptExecutor
}
```

### 3. 實現並行執行邏輯
- 每個腳本獨立的 goroutine
- 腳本生命週期管理 (啟動/停止/重啟)
- 腳本狀態追蹤
- 錯誤隔離 (一個腳本失敗不影響其他)

### 4. 更新 CLI 介面
```bash
# 新增腳本
./run-script-service add-script --name="backup" --path="./backup.sh" --interval="1h"

# 列出所有腳本
./run-script-service list-scripts

# 啟用/停用腳本
./run-script-service enable-script backup
./run-script-service disable-script backup

# 手動執行腳本
./run-script-service run-script backup

# 移除腳本
./run-script-service remove-script backup
```

### 5. 配置檔案範例
```json
{
  "scripts": [
    {
      "name": "main",
      "path": "./run.sh", 
      "interval": 3600,
      "enabled": true,
      "max_log_lines": 100,
      "timeout": 300
    },
    {
      "name": "backup",
      "path": "./backup.sh",
      "interval": 86400,
      "enabled": true, 
      "max_log_lines": 50,
      "timeout": 1800
    }
  ],
  "web_port": 8080
}
```

## 驗收標準
- [ ] 可以同時執行多個腳本
- [ ] 每個腳本有獨立的執行週期
- [ ] 可以動態新增/移除/啟用/停用腳本
- [ ] 腳本執行互不干擾
- [ ] 支援腳本執行超時設定
- [ ] 向後相容原有單腳本配置
- [ ] 所有功能都有單元測試

## 相關檔案
- `service/script_manager.go` (新增)
- `service/script_runner.go` (新增)
- `service/config.go` (修改)
- `main.go` (修改)
- 對應的測試檔案

## 測試案例
- 多腳本並行執行
- 腳本執行失敗隔離
- 動態配置更新
- 腳本超時處理
- 配置檔案向後相容性

## 後續計劃
- [Plan 04: 多日誌管理](04-multi-log-management.md)