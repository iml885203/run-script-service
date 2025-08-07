# Streaming Logging Implementation Plan

## Overview
Convert the current batch logging system to streaming logging for real-time log output during script execution.

## Current Problem
- Logs are only written after script execution completes
- No real-time visibility into long-running scripts
- Users must wait for script completion to see any output
- Poor user experience for monitoring script progress

## Requirements
- Real-time log streaming during script execution
- Preserve existing log management features (rotation, trimming)
- Maintain compatibility with existing log formats
- Support both file logging and web interface streaming
- Handle partial line buffering correctly

## Implementation Plan

### 1. Core Streaming Architecture

#### 1.1 Stream-based Executor
- Replace `io.ReadAll()` with streaming scanners
- Process stdout/stderr line-by-line in real-time
- Maintain separate goroutines for stdout/stderr streaming
- Buffer partial lines until complete

#### 1.2 Real-time Log Writer
- Write log entries immediately as lines are received
- Implement thread-safe concurrent writing
- Handle interleaved stdout/stderr properly
- Maintain existing log format compatibility

### 2. Backend Changes

#### 2.1 Executor Modifications (`service/executor.go`)

**Current Flow:**
```go
// Read all output after completion
stdoutBytes, _ := io.ReadAll(stdout)
stderrBytes, _ := io.ReadAll(stderr)
err = cmd.Wait()
// Write to log after completion
```

**New Streaming Flow:**
```go
// Stream output in real-time
go streamAndLog(stdout, "STDOUT", logWriter)
go streamAndLog(stderr, "STDERR", logWriter)
err = cmd.Wait()
// Finalize log entry
```

#### 2.2 New Streaming Components

##### 2.2.1 `StreamingExecutor` Interface
```go
type StreamingExecutor interface {
    ExecuteWithStreaming(ctx context.Context, args ...string) *ExecutionResult
    SetLogHandler(handler LogHandler)
}

type LogHandler interface {
    HandleLogLine(timestamp time.Time, stream string, line string)
    HandleExecutionStart(timestamp time.Time)
    HandleExecutionEnd(timestamp time.Time, exitCode int)
}
```

##### 2.2.2 `StreamingLogWriter`
```go
type StreamingLogWriter struct {
    logPath     string
    file        *os.File
    mutex       sync.Mutex
    buffer      *bufio.Writer
    flushTicker *time.Ticker
}
```

##### 2.2.3 Line Streaming Function
```go
func (e *Executor) streamOutput(reader io.Reader, streamType string, logWriter *StreamingLogWriter) {
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        line := scanner.Text()
        timestamp := time.Now()
        logWriter.WriteStreamLine(timestamp, streamType, line)

        // Optional: Send to web interface via WebSocket
        e.notifyWebSocket(streamType, line)
    }
}
```

#### 2.3 Log Manager Updates (`service/log_manager.go`)

##### 2.3.1 Streaming Support
- Add `StreamingMode` flag to `ScriptLogger`
- Implement `WriteStreamingEntry()` method
- Handle partial entries during execution
- Maintain backward compatibility with batch mode

##### 2.3.2 Real-time Entry Management
```go
type StreamingLogEntry struct {
    *LogEntry
    InProgress bool
    Lines      []LogLine
}

type LogLine struct {
    Timestamp time.Time
    Stream    string
    Content   string
}
```

### 3. Web Interface Integration

#### 3.1 WebSocket Streaming
- Extend existing WebSocket to support log streaming
- Add real-time log subscription by script name
- Implement log line broadcasting to connected clients
- Handle client connection/disconnection gracefully

#### 3.2 Frontend Updates (`web/frontend/`)

##### 3.2.1 Real-time Log Component
```typescript
// composables/useStreamingLogs.ts
export function useStreamingLogs(scriptName: string) {
  const logLines = ref<LogLine[]>([])
  const isStreaming = ref(false)

  const subscribe = () => {
    // Subscribe to real-time log stream
  }

  const unsubscribe = () => {
    // Unsubscribe from stream
  }

  return { logLines, isStreaming, subscribe, unsubscribe }
}
```

##### 3.2.2 Enhanced Log View
- Auto-scroll to bottom during streaming
- Pause/resume streaming capability
- Toggle between streaming and historical view
- Visual indicators for active streaming

### 4. File Structure Changes

```
service/
├── executor.go              # Modified for streaming
├── streaming_executor.go    # New streaming implementation
├── streaming_logger.go      # New streaming log writer
├── log_manager.go          # Modified for streaming support
└── stream_handler.go       # New stream processing logic

web/
├── streaming_websocket.go   # Enhanced WebSocket for streaming
└── frontend/src/
    ├── composables/
    │   └── useStreamingLogs.ts  # New streaming composable
    └── views/
        └── Logs.vue             # Enhanced with streaming
```

### 5. Implementation Steps

#### Phase 1: Core Streaming Infrastructure
1. **Create streaming executor base**
   - Implement `StreamingExecutor` interface
   - Create `StreamingLogWriter` with buffered writing
   - Add line-by-line processing logic

2. **Modify existing executor**
   - Replace `io.ReadAll()` with streaming scanners
   - Add goroutines for stdout/stderr processing
   - Maintain backward compatibility

#### Phase 2: Log Management Integration
1. **Update log manager**
   - Add streaming mode support
   - Implement real-time entry writing
   - Handle partial entries correctly

2. **File handling improvements**
   - Implement buffered writing with periodic flush
   - Add proper file locking for concurrent access
   - Maintain log rotation during streaming

#### Phase 3: Web Interface Streaming
1. **WebSocket enhancements**
   - Add log streaming channels
   - Implement client subscription management
   - Handle connection state changes

2. **Frontend streaming UI**
   - Create real-time log components
   - Add streaming controls (pause/resume)
   - Implement auto-scroll and search

#### Phase 4: Testing & Optimization
1. **Performance testing**
   - Test with high-volume output scripts
   - Optimize buffering and flush intervals
   - Memory usage optimization

2. **Integration testing**
   - Test streaming with existing features
   - Verify log rotation during streaming
   - Test WebSocket reliability

### 6. Configuration Options

#### 6.1 Streaming Settings
```json
{
  "streaming": {
    "enabled": true,
    "buffer_size": 4096,
    "flush_interval": "100ms",
    "max_line_length": 10000,
    "websocket_enabled": true
  }
}
```

#### 6.2 Backward Compatibility
- Default to streaming mode for new installations
- Provide configuration flag to disable streaming
- Maintain existing log format for compatibility

### 7. Error Handling & Edge Cases

#### 7.1 Stream Errors
- Handle broken pipes gracefully
- Recover from temporary write failures
- Log streaming errors without disrupting execution

#### 7.2 Performance Considerations
- Limit streaming rate for high-volume output
- Implement backpressure for WebSocket clients
- Buffer management for memory efficiency

#### 7.3 Concurrency Safety
- Thread-safe log writing with proper locking
- Atomic operations for shared state
- Prevent race conditions in stream processing

### 8. Testing Strategy

#### 8.1 Unit Tests
- Test streaming executor with mock outputs
- Test log writer with various line patterns
- Test WebSocket streaming with simulated clients

#### 8.2 Integration Tests
- Test with actual shell scripts
- Test log rotation during active streaming
- Test WebSocket client disconnect scenarios

#### 8.3 Performance Tests
- Benchmark streaming vs batch logging
- Memory usage profiling
- High-volume output stress testing

### 9. Migration Strategy

#### 9.1 Gradual Rollout
- Implement as optional feature initially
- Default to streaming for new services
- Provide migration path for existing installations

#### 9.2 Fallback Mechanisms
- Auto-fallback to batch mode on streaming errors
- Configuration option to disable streaming
- Graceful degradation for unsupported scenarios

## Success Criteria

- ✅ Real-time log visibility during script execution
- ✅ Preserve all existing log management features
- ✅ Web interface shows streaming logs in real-time
- ✅ No performance degradation for normal scripts
- ✅ Proper handling of long-running scripts
- ✅ Thread-safe concurrent log writing
- ✅ Backward compatibility with existing logs
- ✅ Comprehensive error handling and recovery
- ✅ Complete test coverage for streaming functionality

## Additional Feature Enhancement: Simplified Log Display

### Current Log Display Issues

The current log display system tries to parse log content into structured format, which causes problems:

1. **Log Format Inconsistency** - Different scripts produce different log formats
2. **Parsing Complexity** - Unnecessary parsing overhead and potential errors
3. **Display Limitations** - Structured parsing may miss important raw output
4. **User Experience** - Users often need to see raw log content for debugging

### Proposed Solution: Raw Log Content Display

#### **Frontend Changes**

##### 1. Simplify Log Entry Display
```vue
<!-- OLD: Complex structured parsing -->
<div class="log-entry">
  <div class="log-timestamp">{{ entry.timestamp }}</div>
  <div class="log-level">{{ entry.level }}</div>
  <div class="log-message">{{ entry.message }}</div>
</div>

<!-- NEW: Simple raw content display -->
<div class="log-entry">
  <div class="log-content">{{ entry.rawContent }}</div>
</div>
```

##### 2. Enhanced Raw Log Viewer
```vue
<template>
  <div class="logs-container">
    <!-- Log Controls -->
    <div class="log-controls">
      <select v-model="selectedScript">
        <option value="">All scripts</option>
        <option v-for="script in scripts" :key="script" :value="script">
          {{ script }}
        </option>
      </select>

      <input
        v-model="searchQuery"
        placeholder="Search logs..."
        class="search-input"
      />

      <button @click="clearLogs" class="btn-danger">
        Clear Logs
      </button>
    </div>

    <!-- Raw Log Display -->
    <div class="log-viewer">
      <pre
        v-for="(entry, index) in filteredLogs"
        :key="`${entry.timestamp}-${index}`"
        class="log-line"
        :class="{ 'highlight': isHighlighted(entry.content) }"
      >{{ entry.content }}</pre>
    </div>
  </div>
</template>
```

##### 3. Improved Log Styling
```css
.log-viewer {
  font-family: 'Courier New', monospace;
  background: #1e1e1e;
  color: #f0f0f0;
  padding: 1rem;
  border-radius: 4px;
  max-height: 600px;
  overflow-y: auto;
  border: 1px solid #333;
}

.log-line {
  margin: 0;
  padding: 2px 0;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.4;
}

.log-line.highlight {
  background-color: rgba(255, 255, 0, 0.2);
}

/* Color coding for different content types */
.log-line:has-text("ERROR") { color: #ff6b6b; }
.log-line:has-text("WARN") { color: #ffa726; }
.log-line:has-text("INFO") { color: #42a5f5; }
.log-line:has-text("DEBUG") { color: #ab47bc; }
```

#### **Backend Changes**

##### 1. Simplified Log Entry Structure
```go
// OLD: Complex structured log entry
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`
    Message   string    `json:"message"`
    Script    string    `json:"script"`
    // ... many other fields
}

// NEW: Simple raw content log entry
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Script    string    `json:"script"`
    Content   string    `json:"content"`     // Raw log content
    Type      string    `json:"type"`        // "stdout", "stderr", "system"
}
```

##### 2. Remove Complex Parsing Logic
```go
// Remove complex log parsing functions
// func parseLogEntry(line string) *LogEntry { ... }
// func extractLogLevel(content string) string { ... }
// func parseTimestamp(content string) time.Time { ... }

// Replace with simple content storage
func (lm *LogManager) WriteRawEntry(script, content, entryType string) {
    entry := &LogEntry{
        Timestamp: time.Now(),
        Script:    script,
        Content:   content,
        Type:      entryType,
    }
    lm.writeEntryToFile(entry)
}
```

##### 3. Updated API Endpoints
```go
// Simplified log retrieval
func (ws *WebServer) handleGetLogs(c *gin.Context) {
    script := c.Query("script")
    limit := c.DefaultQuery("limit", "50")

    logs, err := ws.logManager.GetRawLogs(script, limit)
    if err != nil {
        c.JSON(500, APIResponse{Success: false, Error: err.Error()})
        return
    }

    c.JSON(200, APIResponse{Success: true, Data: logs})
}
```

#### **Implementation Benefits**

1. **Simplified Architecture**
   - Remove complex parsing logic
   - Reduce code maintenance burden
   - Fewer potential parsing errors

2. **Better User Experience**
   - See exact script output
   - No information loss from parsing
   - Faster log loading (no parsing overhead)

3. **Format Flexibility**
   - Works with any log format
   - Supports colored output (ANSI codes)
   - Handles multiline outputs correctly

4. **Performance Improvement**
   - No CPU overhead for parsing
   - Faster log retrieval
   - Reduced memory usage

#### **Migration Strategy**

1. **Backward Compatibility**
   - Keep existing parsed fields for old logs
   - New logs use simplified structure
   - Gradual migration of display components

2. **Configuration Option**
   ```json
   {
     "logging": {
       "display_mode": "raw",        // "raw" or "parsed"
       "preserve_formatting": true,
       "max_line_length": 10000
     }
   }
   ```

3. **Progressive Enhancement**
   - Phase 1: Add raw content field to existing entries
   - Phase 2: Update frontend to use raw display
   - Phase 3: Remove complex parsing (optional)

#### **Implementation Steps**

1. **Backend Updates** (1 day)
   - [ ] Add raw content field to LogEntry
   - [ ] Update log writing to store raw content
   - [ ] Modify API to return raw content
   - [ ] Add configuration for display mode

2. **Frontend Updates** (1 day)
   - [ ] Create raw log viewer component
   - [ ] Update Logs.vue to use raw display
   - [ ] Add search functionality for raw content
   - [ ] Implement syntax highlighting (optional)

3. **Testing & Polish** (0.5 day)
   - [ ] Test with various script outputs
   - [ ] Verify search functionality
   - [ ] Test performance with large logs
   - [ ] Update documentation

## Benefits

1. **Improved User Experience**
   - Real-time visibility into script execution
   - Better monitoring of long-running processes
   - Immediate feedback on script progress
   - **Raw log content display without parsing complexity**

2. **Better Debugging**
   - See output as it happens
   - Identify hang points in scripts
   - Real-time error detection
   - **Exact script output preservation**

3. **Enhanced Web Interface**
   - Live log streaming in browser
   - Better user engagement
   - Professional monitoring experience
   - **Simplified and faster log display**

4. **Operational Benefits**
   - Faster issue detection
   - Reduced waiting time for feedback
   - Better system observability
   - **Reduced parsing overhead and complexity**
