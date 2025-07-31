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

func TestScriptManager_UpdateScript(t *testing.T) {
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

	// Update the script
	updatedScript := ScriptConfig{
		Name:        "test1",
		Path:        "./updated-test1.sh",
		Interval:    120,
		Enabled:     false,
		MaxLogLines: 200,
		Timeout:     60,
	}

	err := manager.UpdateScript("test1", updatedScript)
	if err != nil {
		t.Fatalf("Expected no error updating script, got: %v", err)
	}

	// Check the script was updated in config
	for _, script := range manager.config.Scripts {
		if script.Name == "test1" {
			if script.Path != "./updated-test1.sh" {
				t.Errorf("Expected path to be updated to './updated-test1.sh', got %s", script.Path)
			}
			if script.Interval != 120 {
				t.Errorf("Expected interval to be updated to 120, got %d", script.Interval)
			}
			if script.Enabled != false {
				t.Errorf("Expected enabled to be updated to false, got %t", script.Enabled)
			}
			if script.MaxLogLines != 200 {
				t.Errorf("Expected max log lines to be updated to 200, got %d", script.MaxLogLines)
			}
			if script.Timeout != 60 {
				t.Errorf("Expected timeout to be updated to 60, got %d", script.Timeout)
			}
			return
		}
	}
	t.Error("Script test1 not found in configuration after update")
}

func TestScriptManager_UpdateScript_NotFound(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)

	updatedScript := ScriptConfig{
		Name:        "nonexistent",
		Path:        "./nonexistent.sh",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	err := manager.UpdateScript("nonexistent", updatedScript)
	if err == nil {
		t.Error("Expected error when updating non-existent script")
	}
}

func TestScriptManager_RemoveScript(t *testing.T) {
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

	// Remove test1
	err := manager.RemoveScript("test1")
	if err != nil {
		t.Fatalf("Expected no error removing script, got: %v", err)
	}

	// Check that test1 is gone and test2 remains
	if len(manager.config.Scripts) != 1 {
		t.Errorf("Expected 1 script remaining after removal, got %d", len(manager.config.Scripts))
	}

	if manager.config.Scripts[0].Name != "test2" {
		t.Errorf("Expected remaining script to be test2, got %s", manager.config.Scripts[0].Name)
	}
}

func TestScriptManager_RemoveScript_NotFound(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)

	err := manager.RemoveScript("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent script")
	}
}

func TestScriptManager_RemoveScript_StopsRunningScript(t *testing.T) {
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

	// Start the script
	err := manager.StartScript(ctx, "test1")
	if err != nil {
		t.Fatalf("Error starting script: %v", err)
	}

	if !manager.IsScriptRunning("test1") {
		t.Error("Expected script to be running before removal")
	}

	// Remove the script
	err = manager.RemoveScript("test1")
	if err != nil {
		t.Fatalf("Expected no error removing script, got: %v", err)
	}

	// Check that script is no longer running
	if manager.IsScriptRunning("test1") {
		t.Error("Expected script to be stopped after removal")
	}

	// Check that script is removed from config
	if len(manager.config.Scripts) != 0 {
		t.Errorf("Expected 0 scripts in config after removal, got %d", len(manager.config.Scripts))
	}
}
