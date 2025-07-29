# Plan 02: TDD 開發流程

## 目標
建立 Test-Driven Development (TDD) 開發流程，包括自動化工具鏈和開發規範。

## 前置需求
- [Plan 01: 單元測試基礎設施](01-unit-testing.md) 完成

## 實施步驟

### 1. 建立 Makefile
```makefile
.PHONY: test build clean coverage lint format

# 測試相關
test:
	go test ./... -v

test-watch:
	find . -name "*.go" | entr -c go test ./... -v

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 建構相關
build:
	go build -o run-script-service main.go

clean:
	rm -f run-script-service coverage.out coverage.html

# 代碼品質
lint:
	golangci-lint run

format:
	go fmt ./...
	goimports -w .

# TDD 循環
tdd: test-watch

# CI 管道
ci: format lint test build
```

### 2. 設置開發工具
- 安裝 `entr` (檔案監控)
- 安裝 `golangci-lint` (代碼檢查)
- 安裝 `goimports` (import 整理)
- 設置 IDE/編輯器整合

### 3. 建立 TDD 開發規範
- Red-Green-Refactor 循環
- 測試命名規範
- 提交訊息規範
- 代碼審查檢查清單

### 4. 設置持續整合
- GitHub Actions 配置
- Pre-commit hooks
- 自動化測試報告

## 驗收標準
- [ ] `make test` 可以運行所有測試
- [ ] `make tdd` 可以監控檔案變化自動測試
- [ ] `make ci` 可以執行完整的 CI 管道
- [ ] 代碼覆蓋率報告自動生成
- [ ] 所有代碼檢查工具正常運作

## 相關檔案
- `Makefile` (新增)
- `.github/workflows/ci.yml` (新增)
- `.golangci.yml` (新增)
- `docs/development.md` (新增)

## TDD 開發範例
```bash
# 1. 寫測試 (Red)
make test  # 失敗

# 2. 寫最少代碼使測試通過 (Green)  
make test  # 成功

# 3. 重構 (Refactor)
make test  # 仍然成功

# 4. 重複循環
```

## 後續計劃
- [Plan 03: 多腳本支援](03-multi-script-support.md)