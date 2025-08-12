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
		result := se.executor.ExecuteScriptWithContext(ctx, args...)
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
	config           ScriptConfig
	ticker           *time.Ticker
	cancel           context.CancelFunc
	executor         *ScriptExecutor
	logManager       *LogManager
	eventBroadcaster *EventBroadcaster
	running          bool
	executing        bool
	executionMutex   sync.RWMutex
	restartPending   *ScriptConfig
	restartCallback  func(ScriptConfig) error
	mutex            sync.RWMutex
}

// NewScriptRunner creates a new script runner with the given configuration
func NewScriptRunner(config ScriptConfig, logPath string) *ScriptRunner {
	return &ScriptRunner{
		config:           config,
		executor:         NewScriptExecutor(config.Path, logPath, config.MaxLogLines),
		logManager:       nil,
		eventBroadcaster: nil,
		running:          false,
	}
}

// NewScriptRunnerWithLogManager creates a new script runner with LogManager integration
func NewScriptRunnerWithLogManager(config ScriptConfig, logManager *LogManager) *ScriptRunner {
	return &ScriptRunner{
		config:           config,
		executor:         NewScriptExecutorWithoutLogging(config.Path), // No file logging since we use LogManager
		logManager:       logManager,
		eventBroadcaster: nil,
		running:          false,
	}
}

// NewScriptRunnerWithEventBroadcaster creates a new script runner with event broadcasting
func NewScriptRunnerWithEventBroadcaster(config ScriptConfig, logPath string, broadcaster *EventBroadcaster) *ScriptRunner {
	return &ScriptRunner{
		config:           config,
		executor:         NewScriptExecutor(config.Path, logPath, config.MaxLogLines),
		logManager:       nil,
		eventBroadcaster: broadcaster,
		running:          false,
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
	startTime := time.Now()

	// Set executing state
	sr.setExecuting(true)
	defer sr.setExecuting(false)

	// Broadcast starting event
	if sr.eventBroadcaster != nil {
		startEvent := NewScriptStatusEvent(sr.config.Name, "starting", 0, 0)
		sr.eventBroadcaster.Broadcast(startEvent)
	}

	// Create timeout context if timeout is specified
	if sr.config.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(sr.config.Timeout)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	// Execute the script
	result, err := sr.executor.ExecuteWithResult(ctx, args...)
	duration := time.Since(startTime).Milliseconds()

	// If LogManager is available, use it for structured logging
	if sr.logManager != nil {
		if err != nil {
			// Broadcast failed event if there was an execution error
			if sr.eventBroadcaster != nil {
				failedEvent := NewScriptStatusEvent(sr.config.Name, "failed", -1, duration)
				sr.eventBroadcaster.Broadcast(failedEvent)
			}
			return err
		}

		// Create log entry
		logEntry := &LogEntry{
			Timestamp:  result.Timestamp,
			ScriptName: sr.config.Name,
			ExitCode:   result.ExitCode,
			Stdout:     result.Stdout,
			Stderr:     result.Stderr,
			Duration:   duration,
		}

		// Add to log manager
		logger := sr.logManager.GetLogger(sr.config.Name)
		if addErr := logger.AddEntry(logEntry); addErr != nil {
			// Log error but don't fail the execution
			fmt.Printf("Failed to add log entry: %v\n", addErr)
		}

		// Broadcast completion or failure event
		if sr.eventBroadcaster != nil {
			if result.ExitCode == 0 {
				completedEvent := NewScriptStatusEvent(sr.config.Name, "completed", result.ExitCode, duration)
				sr.eventBroadcaster.Broadcast(completedEvent)
			} else {
				failedEvent := NewScriptStatusEvent(sr.config.Name, "failed", result.ExitCode, duration)
				sr.eventBroadcaster.Broadcast(failedEvent)
			}
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("script exited with code %d", result.ExitCode)
		}
		return nil
	}

	// Handle case with event broadcaster but no log manager
	if err != nil {
		// Broadcast failed event if there was an execution error
		if sr.eventBroadcaster != nil {
			failedEvent := NewScriptStatusEvent(sr.config.Name, "failed", -1, duration)
			sr.eventBroadcaster.Broadcast(failedEvent)
		}
		return err
	}

	// Broadcast completion or failure event
	if sr.eventBroadcaster != nil {
		if result.ExitCode == 0 {
			completedEvent := NewScriptStatusEvent(sr.config.Name, "completed", result.ExitCode, duration)
			sr.eventBroadcaster.Broadcast(completedEvent)
		} else {
			failedEvent := NewScriptStatusEvent(sr.config.Name, "failed", result.ExitCode, duration)
			sr.eventBroadcaster.Broadcast(failedEvent)
		}
	}

	// Check for pending restart after execution
	defer func() {
		if sr.restartPending != nil && sr.restartCallback != nil {
			go sr.restartCallback(*sr.restartPending)
			sr.restartPending = nil
		}
	}()

	// Fallback to old executor method behavior
	if result.ExitCode != 0 {
		return fmt.Errorf("script exited with code %d", result.ExitCode)
	}
	return nil
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

// IsExecuting returns whether the script is currently executing
func (sr *ScriptRunner) IsExecuting() bool {
	sr.executionMutex.RLock()
	defer sr.executionMutex.RUnlock()
	return sr.executing
}

// setExecuting sets the execution state (internal method)
func (sr *ScriptRunner) setExecuting(executing bool) {
	sr.executionMutex.Lock()
	sr.executing = executing
	sr.executionMutex.Unlock()
}

// SetRestartPending sets a pending restart configuration
func (sr *ScriptRunner) SetRestartPending(config ScriptConfig) {
	sr.executionMutex.Lock()
	sr.restartPending = &config
	sr.executionMutex.Unlock()
}

// HasRestartPending returns whether there is a restart pending
func (sr *ScriptRunner) HasRestartPending() bool {
	sr.executionMutex.RLock()
	defer sr.executionMutex.RUnlock()
	return sr.restartPending != nil
}

// GetRestartPendingConfig returns the pending restart configuration
func (sr *ScriptRunner) GetRestartPendingConfig() *ScriptConfig {
	sr.executionMutex.RLock()
	defer sr.executionMutex.RUnlock()
	return sr.restartPending
}
