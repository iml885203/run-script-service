# Plan 01: 單元測試基礎設施

## 目標
為現有的 Go 服務建立完整的單元測試基礎設施，確保代碼品質和可維護性。

## 前置需求
- 現有的 Go 服務正常運行
- Go 測試工具鏈可用

## 實施步驟

### 1. 重構現有代碼使其可測試
- 將 `Service` 結構體拆分為更小的組件
- 抽象檔案系統操作 (可注入 mock)
- 抽象時間操作 (可注入 mock)
- 分離業務邏輯和 I/O 操作

### 2. 建立測試結構
```
├── main.go
├── main_test.go
├── service/
│   ├── service.go
│   ├── service_test.go
│   ├── config.go
│   ├── config_test.go
│   ├── executor.go
│   └── executor_test.go
├── testdata/
│   ├── sample_config.json
│   └── sample_script.sh
└── mocks/
    ├── filesystem.go
    └── time.go
```

### 3. 編寫核心功能測試
- 配置加載/儲存測試
- 腳本執行測試
- 日誌管理測試
- 信號處理測試

### 4. 設置測試輔助工具
- 測試資料準備
- Mock 生成
- 測試覆蓋率報告

## 驗收標準
- [ ] 測試覆蓋率 > 80%
- [ ] 所有核心功能都有單元測試
- [ ] 測試可以獨立運行
- [ ] 測試運行時間 < 5 秒
- [ ] 測試不依賴外部檔案系統

## 相關檔案
- `main.go` (重構)
- `*_test.go` (新增)
- `testdata/` (新增)
- `mocks/` (新增)

## 後續計劃
- [Plan 02: TDD 開發流程](02-tdd-workflow.md)