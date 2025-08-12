package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScriptRunner_New(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "./test1.sh",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	if runner == nil {
		t.Fatal("Expected script runner to be created")
	}

	if runner.config.Name != "test1" {
		t.Errorf("Expected script name to be test1, got %s", runner.config.Name)
	}

	if runner.executor == nil {
		t.Error("Expected executor to be initialized")
	}
}

func TestScriptRunner_Start(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "./test1.sh",
		Interval:    1, // 1 second for quick test
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the runner
	done := make(chan bool)
	go func() {
		runner.Start(ctx)
		done <- true
	}()

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop the runner
	runner.Stop()
	cancel()

	// Wait for completion
	select {
	case <-done:
		// Good, runner stopped
	case <-time.After(2 * time.Second):
		t.Error("Runner did not stop within timeout")
	}
}

func TestScriptRunner_RunOnce(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "echo",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     5,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	ctx := context.Background()
	err := runner.RunOnce(ctx)

	if err != nil {
		t.Errorf("Expected no error running script once, got: %v", err)
	}
}

func TestScriptRunner_RunOnce_WithTimeout(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "sleep",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     1, // 1 second timeout
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	ctx := context.Background()
	err := runner.RunOnce(ctx, "10") // sleep for 10 seconds, should timeout

	if err == nil {
		t.Error("Expected timeout error when script runs longer than timeout")
	}
}

func TestScriptRunner_IsRunning(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "./test1.sh",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	if runner.IsRunning() {
		t.Error("Expected runner to not be running initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the runner
	go func() {
		runner.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	if !runner.IsRunning() {
		t.Error("Expected runner to be running after start")
	}

	// Stop the runner
	runner.Stop()
	cancel()

	// Give it a moment to stop
	time.Sleep(50 * time.Millisecond)

	if runner.IsRunning() {
		t.Error("Expected runner to not be running after stop")
	}
}

func TestScriptRunner_WithLogManager(t *testing.T) {
	tempDir := t.TempDir()
	logsDir := filepath.Join(tempDir, "logs")

	config := ScriptConfig{
		Name:        "test1",
		Path:        "echo",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     5,
	}

	// Create LogManager
	logManager := NewLogManager(logsDir)

	// Create script runner with LogManager integration
	runner := NewScriptRunnerWithLogManager(config, logManager)

	ctx := context.Background()
	err := runner.RunOnce(ctx, "test output")

	if err != nil {
		t.Errorf("Expected no error running script once, got: %v", err)
	}

	// Verify log entry was created
	logger := logManager.GetLogger("test1")
	entries := logger.GetEntries()

	if len(entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(entries))
	}

	if len(entries) > 0 {
		entry := entries[0]
		if entry.ScriptName != "test1" {
			t.Errorf("Expected script name 'test1', got %s", entry.ScriptName)
		}
		if entry.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", entry.ExitCode)
		}
		if entry.Stdout != "test output" {
			t.Errorf("Expected stdout 'test output', got '%s'", entry.Stdout)
		}
	}
}

func TestScriptRunner_WithEventBroadcaster(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "echo",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     5,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")

	// Create event broadcaster
	broadcaster := NewEventBroadcaster()
	events := make(chan *ScriptStatusEvent, 10)
	unsubscribe := broadcaster.Subscribe(events)
	defer unsubscribe()

	// Create script runner with event broadcaster
	runner := NewScriptRunnerWithEventBroadcaster(config, logPath, broadcaster)

	ctx := context.Background()
	err := runner.RunOnce(ctx, "test output")

	if err != nil {
		t.Errorf("Expected no error running script once, got: %v", err)
	}

	// Should receive two events: starting and completed
	receivedEvents := make([]*ScriptStatusEvent, 0, 2)

	// Collect events with timeout
	for i := 0; i < 2; i++ {
		select {
		case event := <-events:
			receivedEvents = append(receivedEvents, event)
		case <-time.After(1 * time.Second):
			t.Fatalf("Expected to receive event %d", i+1)
		}
	}

	if len(receivedEvents) != 2 {
		t.Errorf("Expected 2 events, got %d", len(receivedEvents))
	}

	// First event should be "starting"
	startEvent := receivedEvents[0]
	if startEvent.ScriptName != "test1" {
		t.Errorf("Expected script name 'test1', got %s", startEvent.ScriptName)
	}
	if startEvent.Status != "starting" {
		t.Errorf("Expected status 'starting', got %s", startEvent.Status)
	}

	// Second event should be "completed"
	completeEvent := receivedEvents[1]
	if completeEvent.ScriptName != "test1" {
		t.Errorf("Expected script name 'test1', got %s", completeEvent.ScriptName)
	}
	if completeEvent.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", completeEvent.Status)
	}
	if completeEvent.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", completeEvent.ExitCode)
	}
}

func TestScriptRunner_WithEventBroadcaster_FailedScript(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "false", // Command that always fails with exit code 1
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     5,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")

	// Create event broadcaster
	broadcaster := NewEventBroadcaster()
	events := make(chan *ScriptStatusEvent, 10)
	unsubscribe := broadcaster.Subscribe(events)
	defer unsubscribe()

	// Create script runner with event broadcaster
	runner := NewScriptRunnerWithEventBroadcaster(config, logPath, broadcaster)

	ctx := context.Background()
	err := runner.RunOnce(ctx)

	// Should get an error since the script fails
	if err == nil {
		t.Error("Expected error when script fails")
	}

	// Should receive two events: starting and failed
	receivedEvents := make([]*ScriptStatusEvent, 0, 2)

	// Collect events with timeout
	for i := 0; i < 2; i++ {
		select {
		case event := <-events:
			receivedEvents = append(receivedEvents, event)
		case <-time.After(1 * time.Second):
			t.Fatalf("Expected to receive event %d", i+1)
		}
	}

	if len(receivedEvents) != 2 {
		t.Errorf("Expected 2 events, got %d", len(receivedEvents))
	}

	// First event should be "starting"
	startEvent := receivedEvents[0]
	if startEvent.Status != "starting" {
		t.Errorf("Expected status 'starting', got %s", startEvent.Status)
	}

	// Second event should be "failed"
	failedEvent := receivedEvents[1]
	if failedEvent.Status != "failed" {
		t.Errorf("Expected status 'failed', got %s", failedEvent.Status)
	}
	if failedEvent.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", failedEvent.ExitCode)
	}
}

func TestScriptRunner_IsExecuting(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "sleep",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     5,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	if runner.IsExecuting() {
		t.Error("Expected runner to not be executing initially")
	}

	ctx := context.Background()

	// Run a script that sleeps for a short time
	go func() {
		runner.RunOnce(ctx, "0.1") // Sleep for 100ms
	}()

	// Give it a moment to start executing
	time.Sleep(10 * time.Millisecond)

	if !runner.IsExecuting() {
		t.Error("Expected runner to be executing during script run")
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	if runner.IsExecuting() {
		t.Error("Expected runner to not be executing after script completion")
	}
}

func TestScriptRunner_SetRestartPending(t *testing.T) {
	config := ScriptConfig{
		Name:        "test1",
		Path:        "./test1.sh",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	logPath := filepath.Join(t.TempDir(), "test.log")
	runner := NewScriptRunner(config, logPath)

	// Initially no restart should be pending
	if runner.HasRestartPending() {
		t.Error("Expected no restart pending initially")
	}

	newConfig := ScriptConfig{
		Name:        "test1",
		Path:        "./test1_new.sh",
		Interval:    120,
		Enabled:     true,
		MaxLogLines: 200,
		Timeout:     60,
	}

	runner.SetRestartPending(newConfig)

	if !runner.HasRestartPending() {
		t.Error("Expected restart to be pending after setting")
	}

	pendingConfig := runner.GetRestartPendingConfig()
	if pendingConfig == nil {
		t.Fatal("Expected pending config to be available")
	}

	if pendingConfig.Interval != 120 {
		t.Errorf("Expected pending interval 120, got %d", pendingConfig.Interval)
	}
}

// 🔴 Red Phase: Write failing test for ScriptExecutor.Execute() method (0% coverage)
func TestScriptExecutor_Execute(t *testing.T) {
	// Create a temporary directory for test scripts
	tempDir := t.TempDir()

	t.Run("successful script execution should return nil", func(t *testing.T) {
		// Create a test script that exits successfully
		scriptPath := filepath.Join(tempDir, "success_script.sh")
		scriptContent := "#!/bin/bash\necho 'Hello World'\nexit 0\n"

		err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		// Create ScriptExecutor
		executor := NewScriptExecutorWithoutLogging(scriptPath)

		// Execute the script
		ctx := context.Background()
		err = executor.Execute(ctx)

		// This test should fail initially because Execute() has 0% coverage
		if err != nil {
			t.Errorf("Expected successful execution, got error: %v", err)
		}
	})

	t.Run("failing script execution should return error", func(t *testing.T) {
		// Create a test script that exits with error
		scriptPath := filepath.Join(tempDir, "fail_script.sh")
		scriptContent := "#!/bin/bash\necho 'Error occurred'\nexit 1\n"

		err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		// Create ScriptExecutor
		executor := NewScriptExecutorWithoutLogging(scriptPath)

		// Execute the script
		ctx := context.Background()
		err = executor.Execute(ctx)

		// Should return an error for non-zero exit code
		if err == nil {
			t.Error("Expected error for failing script, got nil")
		}
	})

	t.Run("context cancellation should return error", func(t *testing.T) {
		// Create a test script that runs for a long time
		scriptPath := filepath.Join(tempDir, "long_script.sh")
		scriptContent := "#!/bin/bash\nsleep 5\nexit 0\n"

		err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		// Create ScriptExecutor
		executor := NewScriptExecutorWithoutLogging(scriptPath)

		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Execute the script
		err = executor.Execute(ctx)

		// Should return context deadline exceeded error
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}
