package service

import (
	"context"
	"testing"
)

const testScriptName = "test1"

func TestScriptManager_New(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{
				Name:        testScriptName,
				Path:        "./test1.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)
	if manager == nil {
		t.Fatal("Expected script manager to be created")
	}

	if len(manager.scripts) != 0 {
		t.Errorf("Expected no running scripts initially, got %d", len(manager.scripts))
	}

	if manager.config != config {
		t.Error("Expected config to be set correctly")
	}
}

func TestScriptManager_StartScript(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{
				Name:        "test1",
				Path:        "./test1.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	err := manager.StartScript(ctx, "test1")
	if err != nil {
		t.Fatalf("Expected no error starting script, got: %v", err)
	}

	if len(manager.scripts) != 1 {
		t.Errorf("Expected 1 running script, got %d", len(manager.scripts))
	}

	runner, exists := manager.scripts["test1"]
	if !exists {
		t.Error("Expected script runner to exist for test1")
	}

	if runner.config.Name != testScriptName {
		t.Errorf("Expected script name to be %s, got %s", testScriptName, runner.config.Name)
	}
}

func TestScriptManager_StartScript_NotFound(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	err := manager.StartScript(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when starting non-existent script")
	}
}

func TestScriptManager_StopScript(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{
				Name:        "test1",
				Path:        "./test1.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	// Start the script first
	err := manager.StartScript(ctx, "test1")
	if err != nil {
		t.Fatalf("Error starting script: %v", err)
	}

	// Stop the script
	err = manager.StopScript("test1")
	if err != nil {
		t.Fatalf("Expected no error stopping script, got: %v", err)
	}

	if len(manager.scripts) != 0 {
		t.Errorf("Expected 0 running scripts after stop, got %d", len(manager.scripts))
	}
}

func TestScriptManager_StopScript_NotRunning(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)

	err := manager.StopScript("nonexistent")
	if err == nil {
		t.Error("Expected error when stopping non-running script")
	}
}

func TestScriptManager_StartAllEnabled(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{
				Name:        "enabled1",
				Path:        "./enabled1.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
			{
				Name:        "disabled1",
				Path:        "./disabled1.sh",
				Interval:    60,
				Enabled:     false,
				MaxLogLines: 100,
				Timeout:     30,
			},
			{
				Name:        "enabled2",
				Path:        "./enabled2.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	err := manager.StartAllEnabled(ctx)
	if err != nil {
		t.Fatalf("Expected no error starting all enabled scripts, got: %v", err)
	}

	if len(manager.scripts) != 2 {
		t.Errorf("Expected 2 running scripts (only enabled ones), got %d", len(manager.scripts))
	}

	_, exists1 := manager.scripts["enabled1"]
	if !exists1 {
		t.Error("Expected enabled1 script to be running")
	}

	_, exists2 := manager.scripts["enabled2"]
	if !exists2 {
		t.Error("Expected enabled2 script to be running")
	}

	_, exists3 := manager.scripts["disabled1"]
	if exists3 {
		t.Error("Expected disabled1 script NOT to be running")
	}
}

func TestScriptManager_StopAll(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{
				Name:        "test1",
				Path:        "./test1.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
			{
				Name:        "test2",
				Path:        "./test2.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	// Start all scripts
	err := manager.StartAllEnabled(ctx)
	if err != nil {
		t.Fatalf("Error starting scripts: %v", err)
	}

	// Stop all scripts
	manager.StopAll()

	if len(manager.scripts) != 0 {
		t.Errorf("Expected 0 running scripts after StopAll, got %d", len(manager.scripts))
	}
}
