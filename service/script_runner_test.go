package service

import (
	"context"
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
