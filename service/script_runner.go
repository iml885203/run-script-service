package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScriptExecutor wraps the existing Executor to support context and arguments
type ScriptExecutor struct {
	executor *Executor
}

// NewScriptExecutor creates a new script executor
func NewScriptExecutor(scriptPath, logPath string, maxLines int) *ScriptExecutor {
	return &ScriptExecutor{
		executor: NewExecutor(scriptPath, logPath, maxLines),
	}
}

// Execute executes the script with context support and optional arguments
func (se *ScriptExecutor) Execute(ctx context.Context, args ...string) error {
	// For now, we'll execute synchronously and ignore context cancellation
	// This is a simplified implementation that can be enhanced later
	result := se.executor.ExecuteScript()
	if result.ExitCode != 0 {
		return fmt.Errorf("script exited with code %d", result.ExitCode)
	}
	return nil
}

// ScriptRunner manages the execution of a single script
type ScriptRunner struct {
	config   ScriptConfig
	ticker   *time.Ticker
	cancel   context.CancelFunc
	executor *ScriptExecutor
	running  bool
	mutex    sync.RWMutex
}

// NewScriptRunner creates a new script runner with the given configuration
func NewScriptRunner(config ScriptConfig, logPath string) *ScriptRunner {
	return &ScriptRunner{
		config:   config,
		executor: NewScriptExecutor(config.Path, logPath, config.MaxLogLines),
		running:  false,
	}
}

// Start begins running the script at the configured interval
func (sr *ScriptRunner) Start(ctx context.Context) {
	sr.mutex.Lock()
	if sr.running {
		sr.mutex.Unlock()
		return
	}

	// Create cancellable context
	runCtx, cancel := context.WithCancel(ctx)
	sr.cancel = cancel
	sr.running = true

	// Create ticker for interval execution
	sr.ticker = time.NewTicker(time.Duration(sr.config.Interval) * time.Second)
	sr.mutex.Unlock()

	defer func() {
		sr.mutex.Lock()
		sr.running = false
		sr.ticker.Stop()
		sr.mutex.Unlock()
	}()

	// Run script immediately on start
	sr.RunOnce(runCtx)

	// Then run at intervals
	for {
		select {
		case <-runCtx.Done():
			return
		case <-sr.ticker.C:
			sr.RunOnce(runCtx)
		}
	}
}

// Stop stops the script runner
func (sr *ScriptRunner) Stop() {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if sr.running && sr.cancel != nil {
		sr.cancel()
	}
}

// RunOnce executes the script once with optional arguments
func (sr *ScriptRunner) RunOnce(ctx context.Context, args ...string) error {
	// Create timeout context if timeout is specified
	if sr.config.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(sr.config.Timeout)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	return sr.executor.Execute(ctx, args...)
}

// IsRunning returns whether the script runner is currently running
func (sr *ScriptRunner) IsRunning() bool {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()
	return sr.running
}

// GetConfig returns the script configuration
func (sr *ScriptRunner) GetConfig() ScriptConfig {
	return sr.config
}
