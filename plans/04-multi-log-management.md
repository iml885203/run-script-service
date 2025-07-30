# Plan 04: 多日誌管理

## 目標
為每個腳本建立獨立的日誌管理系統，支援日誌輪轉、查詢和格式化顯示。

## 前置需求
- [Plan 03: 多腳本支援](03-multi-script-support.md) 完成

## 實施步驟

### 1. 設計日誌管理結構 (TDD)
```go
type LogManager struct {
    loggers map[string]*ScriptLogger
    baseDir string
}

type ScriptLogger struct {
    scriptName  string
    logPath     string
    maxLines    int
    entries     []LogEntry
    mutex       sync.RWMutex
}

type LogEntry struct {
    Timestamp  time.Time `json:"timestamp"`
    ScriptName string    `json:"script_name"`
    ExitCode   int       `json:"exit_code"`
    Stdout     string    `json:"stdout"`
    Stderr     string    `json:"stderr"`
    Duration   int64     `json:"duration_ms"`
}
```

### 2. 實現獨立日誌檔案
- 每個腳本有獨立的日誌檔案: `logs/{script_name}.log`
- 日誌格式標準化 (JSON Lines)
- 支援日誌輪轉 (保持最後 N 行)

### 3. 實現日誌查詢功能
```go
type LogQuery struct {
    ScriptName string    `json:"script_name,omitempty"`
    StartTime  time.Time `json:"start_time,omitempty"`
    EndTime    time.Time `json:"end_time,omitempty"`
    ExitCode   *int      `json:"exit_code,omitempty"`
    Limit      int       `json:"limit,omitempty"`
}

func (lm *LogManager) QueryLogs(query LogQuery) ([]LogEntry, error)
```

### 4. 更新 CLI 介面
```bash
# 查看特定腳本日誌
./run-script-service logs --script=backup

# 查看所有腳本日誌
./run-script-service logs --all

# 查看即時日誌 (類似 tail -f)
./run-script-service logs --script=backup --follow

# 過濾日誌
./run-script-service logs --script=backup --exit-code=0 --limit=10

# 清理日誌
./run-script-service clear-logs --script=backup
```

### 5. 日誌目錄結構
```
logs/
├── main.log
├── backup.log
├── cleanup.log
└── archive/
    ├── main_2024-01-01.log
    └── backup_2024-01-01.log
```

### 6. 日誌輪轉策略
- 每個腳本日誌保持最後 100 行 (可配置)
- 舊日誌歸檔到 `archive/` 目錄
- 定期清理過期歸檔 (可配置保留天數)

## 驗收標準
- [ ] 每個腳本有獨立的日誌檔案
- [ ] 日誌格式統一且結構化 (JSON Lines)
- [ ] 支援日誌查詢和過濾
- [ ] 日誌自動輪轉和歸檔
- [ ] CLI 支援多種日誌查看方式
- [ ] 日誌操作不影響腳本執行性能
- [ ] 所有功能都有單元測試

## 相關檔案
- `service/log_manager.go` (新增)
- `service/script_logger.go` (新增)
- `logs/` 目錄 (新增)
- `main.go` (修改 - 新增日誌相關命令)

## 日誌範例
```json
{"timestamp":"2024-01-15T10:30:00Z","script_name":"backup","exit_code":0,"stdout":"Backup completed successfully","stderr":"","duration_ms":1500}
{"timestamp":"2024-01-15T10:35:00Z","script_name":"backup","exit_code":1,"stdout":"","stderr":"Disk space insufficient","duration_ms":500}
```

## 測試案例
- 並行日誌寫入
- 日誌輪轉邏輯
- 日誌查詢效能
- 日誌檔案損壞處理
- 磁碟空間不足處理

## 後續計劃
- [Plan 05: Web 框架設置](05-web-framework.md)
