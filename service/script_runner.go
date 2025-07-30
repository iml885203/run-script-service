// Package service provides core functionality for the run-script-service daemon.
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

// NewScriptExecutorWithoutLogging creates a script executor that doesn't log to files
func NewScriptExecutorWithoutLogging(scriptPath string) *ScriptExecutor {
	return &ScriptExecutor{
		executor: NewExecutor(scriptPath, "", 0), // No logging
	}
}

// Execute executes the script with context support and optional arguments
func (se *ScriptExecutor) Execute(ctx context.Context, args ...string) error {
	result, err := se.ExecuteWithResult(ctx, args...)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("script exited with code %d", result.ExitCode)
	}
	return nil
}

// ExecuteWithResult executes the script and returns detailed execution result
func (se *ScriptExecutor) ExecuteWithResult(ctx context.Context, args ...string) (*ExecutionResult, error) {
	// Check if context is already canceled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create a channel to signal completion
	resultChan := make(chan *ExecutionResult, 1)

	// Execute in a goroutine to allow for cancellation
	go func() {
		result := se.executor.ExecuteScript(args...)
		resultChan <- result
	}()

	// Wait for either completion or cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		return result, nil
	}
}

// ScriptRunner manages the execution of a single script
type ScriptRunner struct {
	config     ScriptConfig
	ticker     *time.Ticker
	cancel     context.CancelFunc
	executor   *ScriptExecutor
	logManager *LogManager
	running    bool
	mutex      sync.RWMutex
}

// NewScriptRunner creates a new script runner with the given configuration
func NewScriptRunner(config ScriptConfig, logPath string) *ScriptRunner {
	return &ScriptRunner{
		config:     config,
		executor:   NewScriptExecutor(config.Path, logPath, config.MaxLogLines),
		logManager: nil,
		running:    false,
	}
}

// NewScriptRunnerWithLogManager creates a new script runner with LogManager integration
func NewScriptRunnerWithLogManager(config ScriptConfig, logManager *LogManager) *ScriptRunner {
	return &ScriptRunner{
		config:     config,
		executor:   NewScriptExecutorWithoutLogging(config.Path), // No file logging since we use LogManager
		logManager: logManager,
		running:    false,
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
	if err := sr.RunOnce(runCtx); err != nil {
		// Log error but continue running - this is expected behavior
		_ = err
	}

	// Then run at intervals
	for {
		select {
		case <-runCtx.Done():
			return
		case <-sr.ticker.C:
			if err := sr.RunOnce(runCtx); err != nil {
				// Log error but continue running - this is expected behavior
				_ = err
			}
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

	// If LogManager is available, use it for structured logging
	if sr.logManager != nil {
		startTime := time.Now()
		result, err := sr.executor.ExecuteWithResult(ctx, args...)
		if err != nil {
			return err
		}

		// Create log entry
		logEntry := &LogEntry{
			Timestamp:  result.Timestamp,
			ScriptName: sr.config.Name,
			ExitCode:   result.ExitCode,
			Stdout:     result.Stdout,
			Stderr:     result.Stderr,
			Duration:   time.Since(startTime).Milliseconds(),
		}

		// Add to log manager
		logger := sr.logManager.GetLogger(sr.config.Name)
		if addErr := logger.AddEntry(logEntry); addErr != nil {
			// Log error but don't fail the execution
			fmt.Printf("Failed to add log entry: %v\n", addErr)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("script exited with code %d", result.ExitCode)
		}
		return nil
	}

	// Fallback to old executor method
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
