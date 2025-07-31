// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"context"
	"fmt"
	"sync"
)

// ScriptManager manages multiple script runners
type ScriptManager struct {
	scripts map[string]*ScriptRunner
	config  *ServiceConfig
	mutex   sync.RWMutex
}

// NewScriptManager creates a new script manager with the given configuration
func NewScriptManager(config *ServiceConfig) *ScriptManager {
	return &ScriptManager{
		scripts: make(map[string]*ScriptRunner),
		config:  config,
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
