# Configuration Management Enhancement Plan

## Overview

研究並改進配置管理系統，探索 Go 語言中類似 `.env` 的配置方式，並考慮將 secret key 等敏感配置整合到現有的配置系統中。

## Current State

### 現有配置方式
- **service_config.json**: 主要配置檔案，包含 scripts、web_port 等設定
- **環境變數**: `WEB_SECRET_KEY` 用於認證密鑰
- **命令行參數**: 部分配置通過命令行傳遞

### 存在的問題
1. **配置分散**: 配置散落在環境變數、JSON 檔案、命令行參數中
2. **缺乏統一管理**: 沒有統一的配置載入機制
3. **敏感資訊暴露**: Secret key 直接顯示在環境變數或日誌中
4. **開發體驗**: 每次重啟需要重新設定環境變數

## Research Areas

### 1. Go 配置管理最佳實踐

#### A. 環境變數載入庫
- **godotenv**: 最受歡迎的 .env 檔案載入庫
- **viper**: 功能強大的配置管理庫，支援多種格式
- **envconfig**: 結構化環境變數載入
- **cleanenv**: 現代化的配置管理庫

#### B. 配置檔案格式
- **.env**: 簡單的 key=value 格式
- **YAML**: 結構化配置，易讀
- **TOML**: Go 社區常用
- **JSON**: 現有格式，程式化友好

### 2. 安全性考量

#### A. 敏感資訊處理
- **配置檔案權限**: 限制檔案存取權限 (600)
- **環境變數優先級**: 環境變數覆蓋檔案設定
- **預設值處理**: 敏感資訊的安全預設值
- **日誌遮罩**: 避免在日誌中暴露敏感資訊

#### B. 部署考量
- **Docker 整合**: 支援 Docker secrets 和環境變數
- **雲端部署**: 支援 AWS/Azure/GCP 的密鑰管理服務
- **開發/生產分離**: 不同環境的配置隔離

## Proposed Solutions

### Option 1: 擴展現有 service_config.json

```json
{
  "scripts": [...],
  "web_port": 8080,
  "security": {
    "secret_key": "auto-generate-if-empty",
    "session_timeout": "24h"
  },
  "logging": {
    "level": "info",
    "max_log_lines": 1000
  }
}
```

**優點**:
- 延續現有架構
- 配置集中化
- JSON 格式程式化友好

**缺點**:
- 敏感資訊明文儲存
- 版本控制風險

### Option 2: 混合配置系統 (.env + JSON)

```bash
# .env 檔案 (敏感資訊)
WEB_SECRET_KEY=your-secret-key
DATABASE_PASSWORD=secret
LOG_LEVEL=debug

# service_config.json (一般配置)
{
  "scripts": [...],
  "web_port": 8080,
  "session_timeout": "24h"
}
```

**優點**:
- 敏感資訊分離
- 開發者友好 (.env 檔案)
- 環境變數優先級

**缺點**:
- 配置分散
- 需要額外依賴

### Option 3: 使用 Viper 統一配置管理

```yaml
# config.yaml
server:
  port: 8080
  timeout: 30s

security:
  secret_key: ${WEB_SECRET_KEY:auto-generate}
  session_timeout: 24h

scripts:
  - name: script1
    path: /path/to/script1.sh
    interval: 300s
```

**優點**:
- 功能強大，支援多種格式
- 環境變數插值
- 配置熱重載
- 廣泛社區支援

**缺點**:
- 學習成本
- 過度工程化風險

## Implementation Plan

### Phase 1: Research & Prototyping (1-2天)

1. **技術調研**
   - [ ] 研究 godotenv, viper, cleanenv 等庫
   - [ ] 比較不同配置格式的優缺點
   - [ ] 分析現有配置載入邏輯

2. **原型開發**
   - [ ] 實現 .env 檔案載入原型
   - [ ] 測試配置優先級 (環境變數 > .env > 預設值)
   - [ ] 驗證向後相容性

### Phase 2: Design & Architecture (1天)

1. **設計配置結構**
   - [ ] 定義統一的配置結構體
   - [ ] 設計配置載入順序和優先級
   - [ ] 規劃敏感資訊處理策略

2. **安全性設計**
   - [ ] 配置檔案權限控制
   - [ ] 日誌敏感資訊遮罩
   - [ ] Secret key 自動生成機制

### Phase 3: Implementation (2-3天)

1. **配置系統重構**
   - [ ] 實現統一配置載入器
   - [ ] 整合 .env 檔案支援
   - [ ] 重構現有配置邏輯

2. **Secret Key 管理**
   - [ ] 實現 secret key 檔案持久化
   - [ ] 添加配置檔案自動生成
   - [ ] 實現安全的預設值機制

3. **測試與驗證**
   - [ ] 單元測試配置載入邏輯
   - [ ] 整合測試不同配置場景
   - [ ] 向後相容性測試

### Phase 4: Documentation & Migration (1天)

1. **文檔更新**
   - [ ] 更新配置指南
   - [ ] 添加最佳實踐說明
   - [ ] 更新 README 配置章節

2. **遷移指南**
   - [ ] 提供現有用戶遷移步驟
   - [ ] 自動配置遷移工具
   - [ ] 範例配置檔案

## Expected Benefits

### 🔧 **開發者體驗**
- 統一的配置管理方式
- 更好的開發環境設定
- 清晰的配置優先級

### 🔐 **安全性**
- 敏感資訊與一般配置分離
- 配置檔案權限控制
- 避免意外暴露敏感資訊

### 📦 **部署靈活性**
- 支援多種部署環境
- 環境變數覆蓋機制
- Docker/雲端友好

### 🔄 **維護性**
- 配置集中管理
- 型別安全的配置載入
- 更好的錯誤處理

## Risk Analysis

### 低風險
- 現有功能向後相容
- 配置載入邏輯相對獨立
- 漸進式遷移可行

### 中風險
- 需要重構部分核心邏輯
- 配置檔案格式變更
- 使用者需要學習新的配置方式

### 高風險
- 敏感資訊處理需特別小心
- 配置載入錯誤可能導致服務無法啟動
- 多種配置來源可能造成混淆

## Decision Criteria

選擇方案時考慮以下因素：
1. **向後相容性**: 不破壞現有用戶的配置
2. **安全性**: 妥善處理敏感資訊
3. **簡潔性**: 避免過度複雜化
4. **社區標準**: 符合 Go 社區最佳實踐
5. **維護成本**: 長期維護的複雜度

## Recommended Approach

基於分析，建議採用 **Option 2: 混合配置系統**：

1. **保持現有 service_config.json** 用於一般配置
2. **引入 .env 檔案** 用於敏感資訊和環境特定配置
3. **使用 godotenv** 庫載入 .env 檔案
4. **維持環境變數優先級** 用於部署彈性

這種方式平衡了功能需求、安全性和實現複雜度。
