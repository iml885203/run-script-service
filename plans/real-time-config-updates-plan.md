# Real-time Script Configuration Updates Plan

## Overview
Implement immediate application of script configuration changes made through the web interface to eliminate inconsistency between displayed settings and actual runtime behavior.

## Current Problem
- Web interface updates only modify configuration files
- Running scripts continue with old settings until manually restarted
- Users experience confusion when changes don't take immediate effect
- No visual feedback about whether changes are applied to running scripts
- Potential data inconsistency between UI state and runtime state

## Requirements
- Immediate application of configuration changes to running scripts
- Graceful restart mechanism for scripts with new settings
- Clear UI feedback about configuration update status
- Preserve script execution state and logs during updates
- Handle edge cases like script currently executing during update
- Maintain backward compatibility with existing script management

## Implementation Plan

### 1. Backend Configuration Hot-Reload

#### 1.1 Enhanced Script Manager
```go
// Enhanced UpdateScript method with immediate application
func (sm *ScriptManager) UpdateScript(name string, updatedConfig ScriptConfig) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    oldConfig, err := sm.getScriptConfig(name)
    if err != nil {
        return err
    }

    // Update configuration
    sm.updateScriptConfig(name, updatedConfig)

    // Apply changes immediately if script is running
    if runner, exists := sm.scripts[name]; exists {
        return sm.applyConfigChanges(name, runner, oldConfig, updatedConfig)
    }

    return nil
}
```

#### 1.2 Configuration Change Detection
```go
type ConfigChange struct {
    Field    string
    OldValue interface{}
    NewValue interface{}
    RequiresRestart bool
}

func (sm *ScriptManager) detectChanges(old, new ScriptConfig) []ConfigChange {
    var changes []ConfigChange

    if old.Interval != new.Interval {
        changes = append(changes, ConfigChange{
            Field: "interval",
            OldValue: old.Interval,
            NewValue: new.Interval,
            RequiresRestart: true,
        })
    }

    if old.Enabled != new.Enabled {
        changes = append(changes, ConfigChange{
            Field: "enabled",
            OldValue: old.Enabled,
            NewValue: new.Enabled,
            RequiresRestart: true,
        })
    }

    // Add other field comparisons...
    return changes
}
```

#### 1.3 Smart Configuration Application
```go
func (sm *ScriptManager) applyConfigChanges(name string, runner *ScriptRunner, oldConfig, newConfig ScriptConfig) error {
    changes := sm.detectChanges(oldConfig, newConfig)

    for _, change := range changes {
        switch change.Field {
        case "enabled":
            if newConfig.Enabled && !oldConfig.Enabled {
                // Script was disabled, now enabled - start it
                return sm.gracefulStart(name, newConfig)
            } else if !newConfig.Enabled && oldConfig.Enabled {
                // Script was enabled, now disabled - stop it
                return sm.gracefulStop(name, runner)
            }

        case "interval":
            // Interval changed - graceful restart required
            return sm.gracefulRestart(name, runner, newConfig)

        case "timeout", "max_log_lines":
            // These can be updated without restart
            return sm.updateRunnerConfig(runner, newConfig)

        case "path":
            // Script path changed - full restart required
            return sm.gracefulRestart(name, runner, newConfig)
        }
    }

    return nil
}
```

#### 1.4 Graceful Script Management
```go
func (sm *ScriptManager) gracefulRestart(name string, runner *ScriptRunner, newConfig ScriptConfig) error {
    // Check if script is currently executing
    if runner.IsExecuting() {
        // Schedule restart after current execution
        return sm.scheduleRestart(name, runner, newConfig)
    }

    // Stop current runner
    runner.Stop()
    delete(sm.scripts, name)

    // Start with new configuration
    return sm.StartScript(context.Background(), name)
}

func (sm *ScriptManager) scheduleRestart(name string, runner *ScriptRunner, newConfig ScriptConfig) error {
    // Set a flag to restart after current execution
    runner.SetRestartPending(newConfig)

    // Return immediately - restart will happen automatically
    return nil
}
```

### 2. Enhanced Script Runner

#### 2.1 Execution State Management
```go
type ScriptRunner struct {
    // ... existing fields
    executing        bool
    executionMutex   sync.RWMutex
    restartPending   *ScriptConfig
    restartCallback  func(ScriptConfig) error
}

func (sr *ScriptRunner) IsExecuting() bool {
    sr.executionMutex.RLock()
    defer sr.executionMutex.RUnlock()
    return sr.executing
}

func (sr *ScriptRunner) SetRestartPending(config ScriptConfig) {
    sr.executionMutex.Lock()
    sr.restartPending = &config
    sr.executionMutex.Unlock()
}
```

#### 2.2 Post-Execution Config Check
```go
func (sr *ScriptRunner) RunOnce(ctx context.Context, args ...string) error {
    // Set executing state
    sr.setExecuting(true)
    defer sr.setExecuting(false)

    // ... existing execution logic

    // Check for pending restart after execution
    defer func() {
        if sr.restartPending != nil && sr.restartCallback != nil {
            go sr.restartCallback(*sr.restartPending)
            sr.restartPending = nil
        }
    }()

    return nil
}
```

### 3. Web Interface Enhancements

#### 3.1 Real-time Update Feedback
```go
type UpdateResponse struct {
    Success      bool                   `json:"success"`
    Message      string                 `json:"message"`
    Applied      bool                   `json:"applied"`
    Scheduled    bool                   `json:"scheduled"`
    Changes      []ConfigChangeInfo     `json:"changes"`
    NextExecution *time.Time            `json:"next_execution,omitempty"`
}

type ConfigChangeInfo struct {
    Field       string      `json:"field"`
    OldValue    interface{} `json:"old_value"`
    NewValue    interface{} `json:"new_value"`
    Applied     bool        `json:"applied"`
    Reason      string      `json:"reason,omitempty"`
}

func (ws *WebServer) handleUpdateScript(c *gin.Context) {
    // ... existing validation logic

    // Apply updates with detailed response
    updateResult := ws.scriptManager.UpdateScriptWithFeedback(scriptName, updateData)

    c.JSON(http.StatusOK, updateResult)
}
```

#### 3.2 WebSocket Configuration Updates
```go
type ConfigUpdateEvent struct {
    Type        string              `json:"type"`
    ScriptName  string              `json:"script_name"`
    Status      string              `json:"status"` // "applied", "scheduled", "failed"
    Changes     []ConfigChangeInfo  `json:"changes"`
    Timestamp   time.Time           `json:"timestamp"`
}

// Broadcast config updates to all connected clients
func (ws *WebServer) broadcastConfigUpdate(event ConfigUpdateEvent) {
    ws.websocketManager.Broadcast(event)
}
```

### 4. Frontend Real-time Feedback

#### 4.1 Configuration Update Status
```typescript
// composables/useScriptUpdates.ts
export function useScriptUpdates() {
  const pendingUpdates = ref<Map<string, ConfigUpdate>>(new Map());

  const updateScript = async (name: string, config: Partial<ScriptConfig>) => {
    try {
      const response = await api.updateScript(name, config);

      if (response.scheduled) {
        // Show pending update indicator
        pendingUpdates.value.set(name, {
          config,
          status: 'scheduled',
          reason: 'Script is currently executing'
        });
      } else if (response.applied) {
        // Show immediate success
        showNotification('Script updated successfully', 'success');
      }

      return response;
    } catch (error) {
      showNotification('Failed to update script', 'error');
      throw error;
    }
  };

  return { updateScript, pendingUpdates };
}
```

#### 4.2 Visual Status Indicators
```vue
<!-- components/ScriptStatusIndicator.vue -->
<template>
  <div class="script-status">
    <!-- Running status -->
    <span
      class="status-indicator"
      :class="{
        running: script.running,
        stopped: !script.running,
        updating: pendingUpdate
      }"
    >
      {{ getStatusText() }}
    </span>

    <!-- Pending update indicator -->
    <div v-if="pendingUpdate" class="pending-update">
      <icon name="clock" />
      <span>Update pending ({{ pendingUpdate.reason }})</span>
      <button @click="forceUpdate" class="force-update-btn">
        Apply Now
      </button>
    </div>

    <!-- Configuration diff -->
    <div v-if="showChanges" class="config-changes">
      <h4>Pending Changes:</h4>
      <ul>
        <li v-for="change in pendingUpdate.changes" :key="change.field">
          <strong>{{ change.field }}:</strong>
          {{ change.old_value }} → {{ change.new_value }}
        </li>
      </ul>
    </div>
  </div>
</template>
```

#### 4.3 Real-time Update Notifications
```vue
<!-- components/ConfigUpdateToast.vue -->
<template>
  <div class="update-toast" v-if="visible">
    <div class="toast-content">
      <icon :name="getStatusIcon()" />
      <div class="message">
        <h4>{{ getTitle() }}</h4>
        <p>{{ getMessage() }}</p>
      </div>
      <button @click="dismiss" class="close-btn">×</button>
    </div>

    <!-- Progress bar for scheduled updates -->
    <div v-if="update.scheduled" class="progress-bar">
      <div class="progress" :style="{ width: progressPercent + '%' }"></div>
    </div>
  </div>
</template>
```

### 5. Configuration Persistence

#### 5.1 Atomic Configuration Updates
```go
func (sm *ScriptManager) atomicConfigUpdate(name string, updater func(*ScriptConfig) error) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    // Create backup of current config
    backup := make([]ScriptConfig, len(sm.config.Scripts))
    copy(backup, sm.config.Scripts)

    // Apply update
    if err := updater(sm.getScriptConfigPointer(name)); err != nil {
        // Restore backup on error
        sm.config.Scripts = backup
        return err
    }

    // Persist to disk
    if err := sm.saveConfig(); err != nil {
        // Restore backup on save error
        sm.config.Scripts = backup
        return err
    }

    return nil
}
```

#### 5.2 Configuration History
```go
type ConfigHistory struct {
    Timestamp time.Time     `json:"timestamp"`
    Script    string        `json:"script"`
    Changes   []ConfigChange `json:"changes"`
    User      string        `json:"user,omitempty"`
}

func (sm *ScriptManager) recordConfigChange(script string, changes []ConfigChange) {
    history := ConfigHistory{
        Timestamp: time.Now(),
        Script:    script,
        Changes:   changes,
    }

    sm.configHistory = append(sm.configHistory, history)

    // Keep only last 100 changes
    if len(sm.configHistory) > 100 {
        sm.configHistory = sm.configHistory[1:]
    }
}
```

### 6. Error Handling & Edge Cases

#### 6.1 Update Conflict Resolution
```go
func (sm *ScriptManager) handleUpdateConflicts(name string, changes []ConfigChange) error {
    runner, exists := sm.scripts[name]
    if !exists {
        return nil // No conflicts if not running
    }

    for _, change := range changes {
        switch {
        case change.Field == "path" && runner.IsExecuting():
            return fmt.Errorf("cannot change script path while executing")

        case change.Field == "enabled" && !change.NewValue.(bool) && runner.IsExecuting():
            // Allow disabling but schedule stop after execution
            runner.SetStopAfterExecution(true)

        default:
            // Other changes can be scheduled
        }
    }

    return nil
}
```

#### 6.2 Recovery Mechanisms
```go
func (sm *ScriptManager) recoverFromFailedUpdate(name string, originalConfig ScriptConfig) error {
    // Stop any partially updated runner
    if runner, exists := sm.scripts[name]; exists {
        runner.Stop()
        delete(sm.scripts, name)
    }

    // Restore original configuration
    sm.updateScriptConfig(name, originalConfig)

    // Restart with original config if it was enabled
    if originalConfig.Enabled {
        return sm.StartScript(context.Background(), name)
    }

    return nil
}
```

### 7. Implementation Steps

#### Phase 1: Backend Hot-Reload Foundation
1. **Enhanced Script Manager**
   - Implement configuration change detection
   - Add graceful restart mechanisms
   - Create execution state tracking

2. **Script Runner Enhancements**
   - Add execution state management
   - Implement pending restart functionality
   - Create post-execution config checks

#### Phase 2: Web Interface Integration
1. **API Response Enhancements**
   - Return detailed update status
   - Include change information
   - Provide scheduling feedback

2. **WebSocket Update Events**
   - Broadcast configuration changes
   - Send real-time status updates
   - Handle client synchronization

#### Phase 3: Frontend Real-time UI
1. **Update Status Components**
   - Visual pending update indicators
   - Configuration change display
   - Real-time status updates

2. **User Experience Polish**
   - Toast notifications
   - Progress indicators
   - Error handling feedback

#### Phase 4: Testing & Validation
1. **Comprehensive Testing**
   - Update scenarios testing
   - Edge case validation
   - Performance impact assessment

2. **User Acceptance Testing**
   - Real-world usage scenarios
   - UI/UX feedback collection
   - Documentation updates

### 8. Configuration Impact Matrix

| Configuration Change | Immediate Effect | Restart Required | User Impact |
|---------------------|------------------|------------------|-------------|
| `enabled: true→false` | Stop after current execution | No | Medium |
| `enabled: false→true` | Start immediately | No | Low |
| `interval` | Apply to next cycle | Yes (graceful) | Medium |
| `timeout` | Apply to next execution | No | Low |
| `max_log_lines` | Apply immediately | No | Low |
| `path` | Full restart required | Yes (immediate) | High |

### 9. Testing Strategy

#### 9.1 Unit Tests
- Configuration change detection logic
- Graceful restart mechanisms
- Update conflict resolution

#### 9.2 Integration Tests
- End-to-end update workflows
- WebSocket event broadcasting
- Configuration persistence

#### 9.3 User Experience Tests
- Update feedback clarity
- Performance impact measurement
- Error scenario handling

### 10. Migration & Rollback

#### 10.1 Gradual Deployment
- Feature flag for hot-reload functionality
- Fallback to old behavior if issues occur
- Progressive rollout to different script types

#### 10.2 Rollback Strategy
- Quick disable of hot-reload features
- Restore original update behavior
- Data consistency verification

## Success Criteria

- ✅ Script configuration changes apply immediately when safe
- ✅ Clear visual feedback about update status in web interface
- ✅ Graceful handling of updates during script execution
- ✅ No data loss or corruption during configuration updates
- ✅ Consistent state between UI and runtime
- ✅ Performance impact < 5% for normal operations
- ✅ Comprehensive error handling and recovery
- ✅ Real-time WebSocket updates for all connected clients

## Benefits

1. **Improved User Experience**
   - Immediate feedback on configuration changes
   - Clear understanding of when changes take effect
   - Reduced confusion about system state

2. **Better System Reliability**
   - Consistent state between UI and runtime
   - Graceful handling of configuration updates
   - Reduced need for manual service restarts

3. **Enhanced Productivity**
   - Faster iteration on script configurations
   - Real-time validation of changes
   - Better debugging experience

4. **Professional Interface**
   - Modern, responsive configuration management
   - Clear status indicators and feedback
   - Intuitive update workflow
