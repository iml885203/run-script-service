# Dashboard Metrics Display Fix Plan

## Problem Description

System Dashboard 中的系統監控指標存在以下問題：

### 🐛 **Issue 1: 不穩定的資料顯示**
- CPU/Memory/Disk Usage 有時候會出現，有時候不會
- 資料載入不一致，影響使用者體驗
- 可能涉及 WebSocket 連線、API 呼叫時序、或前端狀態管理問題

### 🐛 **Issue 2: 數值格式問題**
- Memory Usage 小數點位數過多 (例如: 45.23423432%)
- Disk Usage 小數點位數過多 (例如: 78.87654321%)
- 造成版面跑版，數值難以閱讀
- 缺乏適當的數值格式化

### 🐛 **Issue 3: 版面配置問題**
- 長數值導致 UI 元素變形
- 可能影響響應式設計
- 整體視覺呈現不佳

## Root Cause Analysis

### 可能的問題來源

#### 1. **後端資料收集問題**
```go
// 可能的問題：
// - SystemMonitor 資料收集不穩定
// - API 回傳數據不一致
// - 錯誤處理不當
```

#### 2. **前端資料處理問題**
```typescript
// 可能的問題：
// - WebSocket 連接不穩定
// - 非同步載入競爭條件
// - 狀態更新時序問題
```

#### 3. **數值格式化問題**
```typescript
// 問題：缺乏數值格式化
const usage = 45.234234234; // 原始數值
// 需要：45.2% 或 45%
```

## Investigation Plan

### Phase 1: 問題重現與分析 (1天)

#### A. 重現問題
- [ ] 多次重新整理 Dashboard 觀察資料顯示行為
- [ ] 檢查瀏覽器開發者工具的 Network/Console/WebSocket 訊息
- [ ] 記錄問題出現的頻率和模式
- [ ] 測試不同瀏覽器的表現

#### B. 資料流分析
- [ ] 檢查後端 SystemMonitor API 回應時間和資料結構
- [ ] 追蹤 WebSocket 即時更新機制
- [ ] 分析前端資料載入和狀態管理流程
- [ ] 確認 API 錯誤處理機制

### Phase 2: 後端問題診斷 (0.5天)

#### A. SystemMonitor 檢查
```go
// 檢查項目：
// 1. service/monitor.go - GetSystemMetrics() 方法
// 2. 資料收集的錯誤處理
// 3. CPU/Memory/Disk 計算邏輯
// 4. API response 格式一致性
```

#### B. API 端點驗證
- [ ] 測試 `/api/status` 端點穩定性
- [ ] 檢查回應時間和資料完整性
- [ ] 驗證錯誤處理和預設值
- [ ] 確認數值精度設定

### Phase 3: 前端問題診斷 (0.5天)

#### A. WebSocket 連接檢查
```typescript
// 檢查項目：
// 1. WebSocket 連接狀態管理
// 2. 重連機制
// 3. 資料更新頻率
// 4. 錯誤處理
```

#### B. 狀態管理分析
- [ ] 檢查 Dashboard.vue 的資料載入邏輯
- [ ] 分析非同步操作的時序
- [ ] 驗證響應式資料更新
- [ ] 檢查條件渲染邏輯

## Solution Design

### 1. **穩定性修復**

#### A. 後端改進
```go
type SystemMetrics struct {
    CPU    float64 `json:"cpu"`    // 保證數值有效性
    Memory float64 `json:"memory"` // 添加預設值處理
    Disk   float64 `json:"disk"`   // 添加錯誤處理
    // 添加時間戳和狀態欄位
    Timestamp time.Time `json:"timestamp"`
    Status    string    `json:"status"`
}

// 改進數值收集邏輯
func (sm *SystemMonitor) GetSystemMetrics() (*SystemMetrics, error) {
    // 添加重試機制
    // 添加預設值處理
    // 改進錯誤處理
}
```

#### B. 前端改進
```typescript
interface SystemMetrics {
  cpu: number
  memory: number
  disk: number
  timestamp?: string
  status?: string
}

// 添加載入狀態管理
const systemStats = ref<SystemMetrics | null>(null)
const isLoading = ref(true)
const error = ref<string | null>(null)
```

### 2. **數值格式化修復**

#### A. 前端格式化函數
```typescript
// 新增 utils/formatters.ts
export function formatPercentage(value: number, decimals = 1): string {
  if (typeof value !== 'number' || isNaN(value)) {
    return 'N/A'
  }
  return `${value.toFixed(decimals)}%`
}

export function formatMemoryUsage(value: number): string {
  return formatPercentage(value, 1) // 保留 1 位小數
}

export function formatDiskUsage(value: number): string {
  return formatPercentage(value, 1) // 保留 1 位小數
}

export function formatCpuUsage(value: number): string {
  return formatPercentage(value, 1) // 保留 1 位小數
}
```

#### B. Dashboard 組件更新
```vue
<template>
  <div class="metrics-card">
    <div v-if="isLoading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else class="metrics">
      <div class="metric">
        <h3>CPU Usage</h3>
        <p>{{ formatCpuUsage(systemStats.cpu) }}</p>
      </div>
      <div class="metric">
        <h3>Memory Usage</h3>
        <p>{{ formatMemoryUsage(systemStats.memory) }}</p>
      </div>
      <div class="metric">
        <h3>Disk Usage</h3>
        <p>{{ formatDiskUsage(systemStats.disk) }}</p>
      </div>
    </div>
  </div>
</template>
```

### 3. **版面配置修復**

#### A. CSS 改進
```css
.metric {
  min-width: 120px; /* 確保最小寬度 */
  max-width: 150px; /* 限制最大寬度 */
}

.metric p {
  font-size: 1.5rem;
  font-weight: bold;
  white-space: nowrap; /* 防止換行 */
  overflow: hidden;
  text-overflow: ellipsis; /* 處理過長文字 */
}

.metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 1rem;
}
```

## Implementation Plan

### Phase 1: 後端穩定性修復 (1天)

1. **改進 SystemMonitor**
   - [ ] 添加錯誤處理和重試機制
   - [ ] 實現預設值和數值驗證
   - [ ] 添加時間戳和狀態資訊
   - [ ] 優化資料收集性能

2. **API 端點改進**
   - [ ] 確保 `/api/status` 回應穩定性
   - [ ] 添加適當的 HTTP 錯誤碼
   - [ ] 實現快取機制避免頻繁計算
   - [ ] 添加資料驗證

3. **測試後端修復**
   - [ ] 單元測試數值計算邏輯
   - [ ] 集成測試 API 端點
   - [ ] 壓力測試穩定性
   - [ ] 錯誤情境測試

### Phase 2: 前端顯示修復 (1天)

1. **創建格式化工具**
   - [ ] 實現 `utils/formatters.ts`
   - [ ] 添加數值驗證和錯誤處理
   - [ ] 實現多種格式化選項
   - [ ] 添加單元測試

2. **更新 Dashboard 組件**
   - [ ] 整合格式化函數
   - [ ] 改進載入狀態管理
   - [ ] 添加錯誤顯示機制
   - [ ] 優化 WebSocket 資料處理

3. **版面配置優化**
   - [ ] 修復 CSS 版面問題
   - [ ] 確保響應式設計
   - [ ] 改進視覺呈現
   - [ ] 添加載入動畫

### Phase 3: 整合測試與優化 (0.5天)

1. **端到端測試**
   - [ ] 測試資料顯示穩定性
   - [ ] 驗證格式化效果
   - [ ] 檢查版面響應性
   - [ ] 不同瀏覽器相容性測試

2. **性能優化**
   - [ ] WebSocket 更新頻率調整
   - [ ] 前端渲染性能優化
   - [ ] 減少不必要的 API 呼叫
   - [ ] 添加適當的快取機制

3. **用戶體驗改進**
   - [ ] 添加平滑的數值變化動畫
   - [ ] 改進載入體驗
   - [ ] 添加工具提示說明
   - [ ] 優化錯誤訊息顯示

## Testing Strategy

### 1. **單元測試**
```typescript
// formatters.test.ts
describe('formatPercentage', () => {
  it('should format percentage with specified decimals', () => {
    expect(formatPercentage(45.6789, 1)).toBe('45.7%')
    expect(formatPercentage(45.6789, 2)).toBe('45.68%')
  })

  it('should handle invalid values', () => {
    expect(formatPercentage(NaN)).toBe('N/A')
    expect(formatPercentage(undefined as any)).toBe('N/A')
  })
})
```

### 2. **集成測試**
```go
func TestSystemMonitor_GetSystemMetrics(t *testing.T) {
    monitor := NewSystemMonitor()
    metrics, err := monitor.GetSystemMetrics()

    assert.NoError(t, err)
    assert.NotNil(t, metrics)
    assert.True(t, metrics.CPU >= 0 && metrics.CPU <= 100)
    assert.True(t, metrics.Memory >= 0 && metrics.Memory <= 100)
    assert.True(t, metrics.Disk >= 0 && metrics.Disk <= 100)
}
```

### 3. **E2E 測試**
```typescript
test('Dashboard metrics display correctly', async ({ page }) => {
  await page.goto('/dashboard')

  // 等待資料載入
  await page.waitForSelector('.metrics')

  // 檢查格式化
  const cpuText = await page.textContent('.metric:has-text("CPU") p')
  expect(cpuText).toMatch(/^\d+\.\d%$/) // 例如: 45.7%

  // 檢查穩定性
  await page.reload()
  await page.waitForSelector('.metrics')
  const newCpuText = await page.textContent('.metric:has-text("CPU") p')
  expect(newCpuText).toMatch(/^\d+\.\d%$/)
})
```

## Expected Results

### 📈 **修復後的預期效果**

1. **穩定的資料顯示**
   - ✅ CPU/Memory/Disk 數值每次都正確顯示
   - ✅ WebSocket 連接穩定，即時更新正常
   - ✅ 頁面重新整理後資料載入一致

2. **清晰的數值格式**
   - ✅ Memory Usage: `45.7%` (保留 1 位小數)
   - ✅ Disk Usage: `78.9%` (保留 1 位小數)
   - ✅ CPU Usage: `23.4%` (保留 1 位小數)

3. **良好的版面配置**
   - ✅ 數值不會造成版面跑版
   - ✅ 響應式設計在各種螢幕尺寸下正常
   - ✅ 視覺呈現專業且一致

4. **改進的用戶體驗**
   - ✅ 載入狀態清楚顯示
   - ✅ 錯誤處理機制完善
   - ✅ 數值更新平滑自然

## Risk Assessment

### 低風險
- 數值格式化修復 (僅影響顯示)
- CSS 版面調整 (視覺改進)
- 前端載入狀態改進

### 中風險
- WebSocket 連接邏輯調整
- 後端 SystemMonitor 修改
- API 回應格式變更

### 高風險
- 系統監控核心邏輯變更 (需要充分測試)
- 效能影響 (需要監控資源使用)

## Success Metrics

1. **功能指標**
   - [ ] 資料顯示成功率 > 99%
   - [ ] 頁面載入時間 < 2 秒
   - [ ] WebSocket 連接穩定性 > 95%

2. **用戶體驗指標**
   - [ ] 數值格式一致性 100%
   - [ ] 版面配置零跑版問題
   - [ ] 錯誤處理覆蓋率 100%

3. **技術指標**
   - [ ] 單元測試覆蓋率 > 90%
   - [ ] E2E 測試通過率 100%
   - [ ] 性能回歸測試通過

## 總結

這個修復計劃將系統性地解決 Dashboard 中系統監控顯示的所有問題，確保數據穩定顯示、格式化美觀、版面配置合理。通過前後端的協調改進，提供更好的用戶體驗和系統穩定性。
