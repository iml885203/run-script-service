# Plan 09: Vue.js 前端重構 - 嵌入式單頁應用

## 目標

將現有的原生 JavaScript 前端重構為 Vue.js 3 單頁應用，並完全嵌入到 Go binary 中，實現零外部依賴的部署。

## 前置需求

- ✅ Plan 06: 基礎 Web UI (已完成)
- ✅ Plan 07: Web 編輯功能 (已完成)
- ✅ 統一 Go binary 管理系統 (剛完成)

## 實施步驟

### 階段 1: Vue.js 開發環境設置

1. **建立 Vue 開發目錄結構**
   ```
   web/
   ├── frontend/           # Vue 開發源碼
   │   ├── src/
   │   │   ├── components/
   │   │   ├── views/
   │   │   ├── composables/
   │   │   ├── services/
   │   │   └── main.js
   │   ├── package.json
   │   ├── vite.config.js
   │   └── index.html
   ├── static/            # Go embed 的編譯後文件
   └── server.go          # Go web 服務器
   ```

2. **設置 Vite 建構配置**
   - 配置輸出到 `web/static/` 目錄
   - 設置 base path 為 `/static/`
   - 啟用生產環境優化
   - 設置 asset 內聯門檻

3. **配置 Go embed 集成**
   - 修改 `web/server.go` 使用 `embed.FS`
   - 確保編譯後的 Vue 資源能正確嵌入

### 階段 2: Vue.js 應用架構設計

1. **建立 Vue 3 Composition API 應用**
   ```javascript
   // main.js
   import { createApp } from 'vue'
   import App from './App.vue'
   import router from './router'

   createApp(App).use(router).mount('#app')
   ```

2. **設計組件架構**
   - `App.vue` - 根組件，包含導航和路由出口
   - `Dashboard.vue` - 儀表板視圖
   - `Scripts.vue` - 腳本管理視圖
   - `Logs.vue` - 日誌查看視圖
   - `Settings.vue` - 設置視圖

3. **建立共享服務層**
   ```javascript
   // services/api.js
   export class ApiService {
     static async getScripts() { ... }
     static async addScript(script) { ... }
     static async runScript(name) { ... }
   }
   ```

### 階段 3: 核心功能遷移

1. **Dashboard 組件開發**
   - 系統狀態顯示
   - 運行中腳本列表
   - 實時狀態更新
   - WebSocket 連接管理

2. **Scripts 管理組件**
   - 腳本列表與狀態
   - 添加新腳本表單
   - 編輯腳本配置
   - 啟用/禁用切換
   - 立即執行功能

3. **Logs 查看組件**
   - 實時日誌流
   - 腳本過濾功能
   - 日誌清除操作
   - 搜索與分頁

4. **Settings 配置組件**
   - Web 端口設置
   - 自動刷新間隔
   - 系統監控配置

### 階段 4: 響應式設計與 UX 優化

1. **實現響應式佈局**
   - 移動設備適配
   - 平板設備優化
   - 桌面端多列佈局

2. **提升用戶體驗**
   - Loading 狀態指示器
   - 錯誤處理與提示
   - 操作確認對話框
   - Toast 通知系統

3. **性能優化**
   - 路由懶加載
   - 組件按需導入
   - API 請求去重
   - 虛擬滾動（大量日誌）

### 階段 5: 建構與部署集成

1. **設置建構流程**
   ```bash
   # Makefile 新增目標
   make build-frontend    # 建構 Vue 應用
   make embed-frontend    # 重新生成 Go embed
   make build-all         # 完整建構流程
   ```

2. **自動化工作流**
   - Git pre-commit hook 自動建構
   - CI/CD 集成
   - 版本號同步

3. **Go binary 整合**
   - 確保單一 binary 包含所有資源
   - 測試各平台編譯
   - 驗證嵌入資源正確性

## 技術架構

### Vue.js 技術棧

```javascript
{
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.2.0",
    "@vueuse/core": "^10.7.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "@vitejs/plugin-vue": "^4.5.0",
    "sass": "^1.69.0"
  }
}
```

### Go 嵌入配置

```go
//go:embed frontend/dist/*
var staticFiles embed.FS

func NewWebServer() *WebServer {
    // 使用嵌入的靜態文件
    staticFS, _ := fs.Sub(staticFiles, "frontend/dist")
    return &WebServer{
        staticFS: http.FS(staticFS),
    }
}
```

### 建構配置

```javascript
// vite.config.js
export default {
  base: '/static/',
  build: {
    outDir: '../static',
    assetsInlineLimit: 8192,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router']
        }
      }
    }
  }
}
```

## 驗收標準

### 功能完整性
- [ ] 所有現有功能完整遷移到 Vue.js
- [ ] WebSocket 實時通信正常工作
- [ ] API 調用全部正常
- [ ] 路由導航無錯誤

### 用戶體驗
- [ ] 響應式設計在各設備上正常顯示
- [ ] 頁面加載時間 < 2 秒
- [ ] 操作響應時間 < 500ms
- [ ] 錯誤處理和用戶反饋完善

### 部署需求
- [ ] 單一 Go binary 包含所有前端資源
- [ ] 無需額外前端建構步驟即可運行
- [ ] 跨平台編譯正常
- [ ] 生產建構大小合理（< 15MB）

### 開發體驗
- [ ] 熱重載開發環境
- [ ] 清晰的錯誤提示
- [ ] 完整的 TypeScript 支持（可選）
- [ ] 組件化和可維護的代碼結構

## 相關文件

### 新建文件
- `web/frontend/package.json` - Node.js 依賴配置
- `web/frontend/vite.config.js` - Vite 建構配置
- `web/frontend/src/main.js` - Vue 應用入口
- `web/frontend/src/App.vue` - 根組件
- `web/frontend/src/router/index.js` - 路由配置
- `web/frontend/src/components/` - Vue 組件目錄
- `web/frontend/src/views/` - 視圖組件目錄
- `web/frontend/src/services/api.js` - API 服務層

### 修改文件
- `web/server.go` - 更新靜態文件服務邏輯
- `Makefile` - 添加前端建構目標
- `README.md` - 更新建構說明
- `.gitignore` - 忽略 node_modules 和建構文件

## 測試案例

### 單元測試
- [ ] Vue 組件渲染測試
- [ ] API 服務調用測試
- [ ] 路由導航測試
- [ ] 狀態管理測試

### 集成測試
- [ ] 前後端 API 集成測試
- [ ] WebSocket 連接測試
- [ ] 文件上傳下載測試
- [ ] 實時更新功能測試

### 端到端測試
- [ ] 完整用戶流程測試
- [ ] 跨瀏覽器兼容性測試
- [ ] 移動設備響應式測試
- [ ] 性能負載測試

## 風險評估

### 技術風險
- **建構複雜度**: Vue.js 建構可能增加部署複雜度
  - 緩解: 完全自動化建構流程，確保一鍵部署
- **Bundle 大小**: Vue.js 可能增加 binary 大小
  - 緩解: 樹搖優化，按需加載，gzip 壓縮

### 遷移風險
- **功能丟失**: 遷移過程中可能遺漏現有功能
  - 緩解: 詳細的功能清單和測試覆蓋
- **性能回退**: Vue.js 可能比原生 JS 慢
  - 緩解: 性能基準測試，優化關鍵路徑

### 維護風險
- **技術棧複雜化**: 增加 Node.js 生態依賴
  - 緩解: 最小化依賴，固定版本，離線建構

## 時程估算

- **階段 1**: 環境設置 (1-2 天)
- **階段 2**: 架構設計 (2-3 天)
- **階段 3**: 功能遷移 (5-7 天)
- **階段 4**: UX 優化 (3-4 天)
- **階段 5**: 集成部署 (2-3 天)

**總估算**: 13-19 工作天

## 後續計劃

完成此計劃後，可以考慮：
- **Plan 10**: TypeScript 支持
- **Plan 11**: PWA 功能（離線使用）
- **Plan 12**: 多主題支持
- **Plan 13**: 國際化 (i18n)
