// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"context"
	"fmt"
	"sync"
)

// ConfigChange represents a configuration change
type ConfigChange struct {
	Field           string
	OldValue        interface{}
	NewValue        interface{}
	RequiresRestart bool
}

// ScriptManager manages multiple script runners
type ScriptManager struct {
	scripts    map[string]*ScriptRunner
	config     *ServiceConfig
	configPath string
	mutex      sync.RWMutex
}

// NewScriptManager creates a new script manager with the given configuration
func NewScriptManager(config *ServiceConfig) *ScriptManager {
	return &ScriptManager{
		scripts: make(map[string]*ScriptRunner),
		config:  config,
	}
}

// NewScriptManagerWithPath creates a new script manager with configuration and config path
func NewScriptManagerWithPath(config *ServiceConfig, configPath string) *ScriptManager {
	return &ScriptManager{
		scripts:    make(map[string]*ScriptRunner),
		config:     config,
		configPath: configPath,
	}
}

// StartScript starts a script by name
func (sm *ScriptManager) StartScript(ctx context.Context, name string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if script is already running
	if _, exists := sm.scripts[name]; exists {
		return fmt.Errorf("script %s is already running", name)
	}

	// Find the script config
	var scriptConfig *ScriptConfig
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			scriptConfig = &sm.config.Scripts[i]
			break
		}
	}

	if scriptConfig == nil {
		return fmt.Errorf("script %s not found in configuration", name)
	}

	// Create and start the script runner
	logPath := fmt.Sprintf("%s.log", name) // Simple log path for now
	runner := NewScriptRunner(*scriptConfig, logPath)
	sm.scripts[name] = runner

	// Start the runner in a goroutine
	go func() {
		runner.Start(ctx)
		// Clean up when runner stops
		sm.mutex.Lock()
		delete(sm.scripts, name)
		sm.mutex.Unlock()
	}()

	return nil
}

// StopScript stops a script by name
func (sm *ScriptManager) StopScript(name string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	runner, exists := sm.scripts[name]
	if !exists {
		return fmt.Errorf("script %s is not running", name)
	}

	runner.Stop()
	delete(sm.scripts, name)
	return nil
}

// StartAllEnabled starts all enabled scripts
func (sm *ScriptManager) StartAllEnabled(ctx context.Context) error {
	for _, scriptConfig := range sm.config.Scripts {
		if scriptConfig.Enabled {
			if err := sm.StartScript(ctx, scriptConfig.Name); err != nil {
				return fmt.Errorf("failed to start script %s: %v", scriptConfig.Name, err)
			}
		}
	}
	return nil
}

// StopAll stops all running scripts
func (sm *ScriptManager) StopAll() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for name, runner := range sm.scripts {
		runner.Stop()
		delete(sm.scripts, name)
	}
}

// GetScripts returns all configured scripts
func (sm *ScriptManager) GetScripts() ([]ScriptConfig, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy to prevent external modification
	scripts := make([]ScriptConfig, len(sm.config.Scripts))
	copy(scripts, sm.config.Scripts)

	return scripts, nil
}

// GetRunningScripts returns a list of currently running script names
func (sm *ScriptManager) GetRunningScripts() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var running []string
	for name := range sm.scripts {
		running = append(running, name)
	}
	return running
}

// IsScriptRunning checks if a specific script is running
func (sm *ScriptManager) IsScriptRunning(name string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	_, exists := sm.scripts[name]
	return exists
}

// GetConfig returns the script manager's configuration
func (sm *ScriptManager) GetConfig() *ServiceConfig {
	return sm.config
}

// SaveConfig saves the current configuration to file
func (sm *ScriptManager) SaveConfig() error {
	if sm.configPath == "" {
		return fmt.Errorf("config path not set - cannot save configuration")
	}
	return SaveServiceConfig(sm.configPath, sm.config)
}

// AddScript adds a new script configuration
func (sm *ScriptManager) AddScript(scriptConfig ScriptConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if script with same name already exists
	for _, existing := range sm.config.Scripts {
		if existing.Name == scriptConfig.Name {
			return fmt.Errorf("script with name %s already exists", scriptConfig.Name)
		}
	}

	// Add the script to configuration
	sm.config.Scripts = append(sm.config.Scripts, scriptConfig)
	return nil
}

// RunScriptOnce executes a script once by name
func (sm *ScriptManager) RunScriptOnce(ctx context.Context, name string) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Find the script config
	var scriptConfig *ScriptConfig
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			scriptConfig = &sm.config.Scripts[i]
			break
		}
	}

	if scriptConfig == nil {
		return fmt.Errorf("script %s not found in configuration", name)
	}

	// Create a temporary script runner for one-time execution
	logPath := fmt.Sprintf("%s.log", name)
	runner := NewScriptRunner(*scriptConfig, logPath)

	return runner.RunOnce(ctx)
}

// EnableScript enables a script by name
func (sm *ScriptManager) EnableScript(name string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Find the script config
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			sm.config.Scripts[i].Enabled = true
			return nil
		}
	}

	return fmt.Errorf("script %s not found in configuration", name)
}

// DisableScript disables a script by name
func (sm *ScriptManager) DisableScript(name string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Find the script config
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			sm.config.Scripts[i].Enabled = false
			return nil
		}
	}

	return fmt.Errorf("script %s not found in configuration", name)
}

// UpdateScript updates an existing script configuration
func (sm *ScriptManager) UpdateScript(name string, updatedConfig ScriptConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Find the script config and update it
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			// Ensure the name matches the parameter
			updatedConfig.Name = name
			sm.config.Scripts[i] = updatedConfig
			return nil
		}
	}

	return fmt.Errorf("script %s not found in configuration", name)
}

// RemoveScript removes a script from configuration and stops it if running
func (sm *ScriptManager) RemoveScript(name string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Stop the script if it's running
	if runner, exists := sm.scripts[name]; exists {
		runner.Stop()
		delete(sm.scripts, name)
	}

	// Find and remove the script from configuration
	found := false
	newScripts := make([]ScriptConfig, 0, len(sm.config.Scripts))
	for _, sc := range sm.config.Scripts {
		if sc.Name != name {
			newScripts = append(newScripts, sc)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("script %s not found in configuration", name)
	}

	sm.config.Scripts = newScripts
	return nil
}

// detectChanges detects differences between old and new script configurations
func (sm *ScriptManager) detectChanges(old, new ScriptConfig) []ConfigChange {
	var changes []ConfigChange

	if old.Interval != new.Interval {
		changes = append(changes, ConfigChange{
			Field:           "interval",
			OldValue:        old.Interval,
			NewValue:        new.Interval,
			RequiresRestart: true,
		})
	}

	if old.Enabled != new.Enabled {
		changes = append(changes, ConfigChange{
			Field:           "enabled",
			OldValue:        old.Enabled,
			NewValue:        new.Enabled,
			RequiresRestart: true,
		})
	}

	if old.Path != new.Path {
		changes = append(changes, ConfigChange{
			Field:           "path",
			OldValue:        old.Path,
			NewValue:        new.Path,
			RequiresRestart: true,
		})
	}

	if old.MaxLogLines != new.MaxLogLines {
		changes = append(changes, ConfigChange{
			Field:           "max_log_lines",
			OldValue:        old.MaxLogLines,
			NewValue:        new.MaxLogLines,
			RequiresRestart: false,
		})
	}

	if old.Timeout != new.Timeout {
		changes = append(changes, ConfigChange{
			Field:           "timeout",
			OldValue:        old.Timeout,
			NewValue:        new.Timeout,
			RequiresRestart: false,
		})
	}

	return changes
}

// UpdateScriptWithImmediateApplication updates a script and applies changes immediately to running scripts
func (sm *ScriptManager) UpdateScriptWithImmediateApplication(name string, updatedConfig ScriptConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Find the current configuration
	var oldConfig *ScriptConfig
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			oldConfig = &sc
			// Update configuration first
			updatedConfig.Name = name
			sm.config.Scripts[i] = updatedConfig
			break
		}
	}

	if oldConfig == nil {
		return fmt.Errorf("script %s not found in configuration", name)
	}

	// Detect changes
	changes := sm.detectChanges(*oldConfig, updatedConfig)

	// Apply changes immediately if script is running
	if runner, exists := sm.scripts[name]; exists {
		return sm.applyConfigChanges(name, runner, *oldConfig, updatedConfig, changes)
	}

	// Script not running, configuration update is sufficient
	return nil
}

// applyConfigChanges applies configuration changes to a running script
func (sm *ScriptManager) applyConfigChanges(name string, runner *ScriptRunner, oldConfig, newConfig ScriptConfig, changes []ConfigChange) error {
	// For now, implement basic logic - this will be enhanced in subsequent steps
	for _, change := range changes {
		switch change.Field {
		case "enabled":
			if newConfig.Enabled && !oldConfig.Enabled {
				// Script was disabled, now enabled - but it's already running, so no action needed
				return nil
			} else if !newConfig.Enabled && oldConfig.Enabled {
				// Script was enabled, now disabled - stop it
				runner.Stop()
				delete(sm.scripts, name)
				return nil
			}
		case "interval", "path":
			// Changes that require restart - implement graceful restart
			return sm.gracefulRestartScript(name, runner, newConfig)
		case "timeout", "max_log_lines":
			// These changes can be applied without restart
			// For now, just log that they would be applied
			continue
		}
	}
	return nil
}

// UpdateScriptWithFeedback updates a script and returns detailed feedback about the changes
func (sm *ScriptManager) UpdateScriptWithFeedback(name string, updatedConfig ScriptConfig) UpdateResponse {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Find the current configuration
	var oldConfig *ScriptConfig
	for i, sc := range sm.config.Scripts {
		if sc.Name == name {
			oldConfig = &sc
			// Update configuration first
			updatedConfig.Name = name
			sm.config.Scripts[i] = updatedConfig
			break
		}
	}

	if oldConfig == nil {
		return UpdateResponse{
			Success:   false,
			Message:   fmt.Sprintf("Script %s not found in configuration", name),
			Applied:   false,
			Scheduled: false,
			Changes:   []ConfigChangeInfo{},
		}
	}

	// Detect changes
	changes := sm.detectChanges(*oldConfig, updatedConfig)

	// Convert ConfigChange to ConfigChangeInfo
	changeInfos := make([]ConfigChangeInfo, len(changes))
	allApplied := true
	anyScheduled := false

	for i, change := range changes {
		applied := false
		reason := ""

		// Determine if change can be applied immediately
		if runner, exists := sm.scripts[name]; exists {
			if runner.IsExecuting() {
				// Script is executing, schedule for later
				applied = false
				anyScheduled = true
				reason = "Script is currently executing, change will be applied after completion"
				runner.SetRestartPending(updatedConfig)
			} else {
				// Script is idle, apply immediately based on change type
				switch change.Field {
				case "timeout", "max_log_lines":
					// These can be applied without restart
					applied = true
					reason = "Applied immediately"
				case "enabled":
					if updatedConfig.Enabled && !oldConfig.Enabled {
						// Re-enabling - already running, no action needed
						applied = true
						reason = "Script already running"
					} else if !updatedConfig.Enabled && oldConfig.Enabled {
						// Disabling - stop the script
						runner.Stop()
						delete(sm.scripts, name)
						applied = true
						reason = "Script stopped successfully"
					}
				case "interval", "path":
					// These require restart
					applied = false
					anyScheduled = true
					reason = "Requires graceful restart, scheduled for next execution cycle"
					runner.SetRestartPending(updatedConfig)
				}
			}
		} else {
			// Script not running, all changes are effectively "applied"
			applied = true
			reason = "Script not currently running, configuration updated"
		}

		if !applied {
			allApplied = false
		}

		changeInfos[i] = ConfigChangeInfo{
			Field:    change.Field,
			OldValue: change.OldValue,
			NewValue: change.NewValue,
			Applied:  applied,
			Reason:   reason,
		}
	}

	// Determine overall status
	message := fmt.Sprintf("Script %s updated successfully", name)
	if anyScheduled {
		message += " (some changes scheduled for next execution cycle)"
	}

	return UpdateResponse{
		Success:   true,
		Message:   message,
		Applied:   allApplied && !anyScheduled,
		Scheduled: anyScheduled,
		Changes:   changeInfos,
	}
}

// gracefulRestartScript stops a running script and starts it again with the new configuration
func (sm *ScriptManager) gracefulRestartScript(name string, runner *ScriptRunner, newConfig ScriptConfig) error {
	// Stop the current script
	runner.Stop()
	delete(sm.scripts, name)

	// Create and start a new script runner with updated configuration
	logPath := fmt.Sprintf("%s.log", name)
	newRunner := NewScriptRunner(newConfig, logPath)
	sm.scripts[name] = newRunner

	// Start the new runner in a goroutine
	go func() {
		ctx := context.Background() // Use background context for restart
		newRunner.Start(ctx)
		// Clean up when runner stops
		sm.mutex.Lock()
		delete(sm.scripts, name)
		sm.mutex.Unlock()
	}()

	return nil
}
