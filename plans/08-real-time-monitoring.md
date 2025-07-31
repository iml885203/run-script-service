# Plan 08: Real-time System Monitoring

## 目標 (Objectives)

實施即時系統監控功能，提供實時腳本執行狀態、系統資源使用情況和執行統計的即時更新。

Implement real-time system monitoring features providing live updates of script execution status, system resource usage, and execution statistics.

## 前置需求 (Prerequisites)

- ✅ Plan 01: Unit Testing Infrastructure
- ✅ Plan 02: TDD Workflow
- ✅ Plan 03: Multi-Script Support
- ✅ Plan 04: Multi-Log Management
- ✅ Plan 05: Web Framework Setup
- ✅ Plan 06: Basic Web UI
- ✅ Plan 07: Web Editing Features

## 實施步驟 (Implementation Steps)

### Step 1: WebSocket Infrastructure (TDD)
- [x] **測試**: Create WebSocket connection handler tests
- [x] **實作**: Implement WebSocket server endpoint `/ws`
- [x] **測試**: Test real-time message broadcasting
- [x] **實作**: Add WebSocket client connection management

### Step 2: Real-time Script Status Updates (TDD)
- [ ] **測試**: Create tests for script status change events
- [ ] **實作**: Implement script status event system
- [ ] **測試**: Test WebSocket broadcasting of script events
- [ ] **實作**: Add script start/stop/error event broadcasting

### Step 3: System Resource Monitoring (TDD)
- [ ] **測試**: Create tests for system metrics collection
- [ ] **實作**: Implement CPU, memory, disk usage monitoring
- [ ] **測試**: Test metrics data structure and formatting
- [ ] **實作**: Add periodic system metrics broadcasting

### Step 4: Live Dashboard Updates (TDD)
- [ ] **測試**: Create frontend WebSocket client tests
- [ ] **實作**: Implement WebSocket client in JavaScript
- [ ] **測試**: Test real-time UI updates
- [ ] **實作**: Add live dashboard metrics display

## 驗收標準 (Acceptance Criteria)

1. **WebSocket Connection**: Web clients can establish stable WebSocket connections
2. **Real-time Updates**: Script status changes are immediately reflected in the web UI
3. **System Metrics**: CPU, memory, and disk usage are displayed and updated in real-time
4. **Connection Management**: Multiple clients can connect simultaneously without issues
5. **Error Handling**: WebSocket disconnections are handled gracefully with auto-reconnect
6. **Performance**: Monitoring doesn't significantly impact script execution performance

## 相關檔案 (Related Files)

### New Files
- `web/websocket.go` - WebSocket server implementation
- `web/websocket_test.go` - WebSocket server tests
- `service/monitor.go` - System monitoring implementation
- `service/monitor_test.go` - System monitoring tests
- `web/static/js/websocket.js` - WebSocket client implementation

### Modified Files
- `web/server.go` - Add WebSocket endpoint
- `web/server_test.go` - Add WebSocket tests
- `service/service.go` - Add monitoring integration
- `service/script_runner.go` - Add status event broadcasting
- `web/static/js/app.js` - Add real-time UI updates
- `web/static/index.html` - Add real-time status indicators

## 技術細節 (Technical Details)

### WebSocket Message Format
```json
{
  "type": "script_status|system_metrics|error",
  "timestamp": "2025-07-31T19:00:00Z",
  "data": {
    // Type-specific payload
  }
}
```

### System Metrics Structure
```json
{
  "cpu_percent": 45.2,
  "memory_percent": 67.8,
  "disk_percent": 23.1,
  "active_scripts": 3,
  "total_executions": 156
}
```

### Script Status Events
```json
{
  "script_name": "backup.sh",
  "status": "running|completed|failed",
  "exit_code": 0,
  "duration": 1234
}
```
