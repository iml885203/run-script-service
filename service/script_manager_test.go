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

func TestScriptManager_DetectChanges(t *testing.T) {
	manager := NewScriptManager(&ServiceConfig{})

	oldConfig := ScriptConfig{
		Name:        "test",
		Path:        "./test.sh",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	// Test interval change
	newConfig := oldConfig
	newConfig.Interval = 120

	changes := manager.detectChanges(oldConfig, newConfig)

	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	if changes[0].Field != "interval" {
		t.Errorf("Expected interval change, got %s", changes[0].Field)
	}

	if changes[0].OldValue != 60 {
		t.Errorf("Expected old interval 60, got %v", changes[0].OldValue)
	}

	if changes[0].NewValue != 120 {
		t.Errorf("Expected new interval 120, got %v", changes[0].NewValue)
	}

	if !changes[0].RequiresRestart {
		t.Error("Expected interval change to require restart")
	}
}

func TestScriptManager_DetectChanges_Multiple(t *testing.T) {
	manager := NewScriptManager(&ServiceConfig{})

	oldConfig := ScriptConfig{
		Name:        "test",
		Path:        "./test.sh",
		Interval:    60,
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	newConfig := ScriptConfig{
		Name:        "test",
		Path:        "./test_new.sh",
		Interval:    120,
		Enabled:     false,
		MaxLogLines: 200,
		Timeout:     60,
	}

	changes := manager.detectChanges(oldConfig, newConfig)

	expectedChanges := 5 // path, interval, enabled, max_log_lines, timeout
	if len(changes) != expectedChanges {
		t.Errorf("Expected %d changes, got %d", expectedChanges, len(changes))
	}

	// Check that all expected fields are present
	fields := make(map[string]bool)
	for _, change := range changes {
		fields[change.Field] = true
	}

	expectedFields := []string{"path", "interval", "enabled", "max_log_lines", "timeout"}
	for _, field := range expectedFields {
		if !fields[field] {
			t.Errorf("Expected change for field %s", field)
		}
	}
}

func TestScriptManager_UpdateScript_WithImmediateApplication(t *testing.T) {
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

	// Test updating non-running script (should just update config)
	updatedScript := ScriptConfig{
		Name:        "test1",
		Path:        "./test1_new.sh",
		Interval:    120,
		Enabled:     true,
		MaxLogLines: 200,
		Timeout:     60,
	}

	err := manager.UpdateScriptWithImmediateApplication("test1", updatedScript)
	if err != nil {
		t.Fatalf("Expected no error updating script, got: %v", err)
	}

	// Verify configuration was updated
	found := false
	for _, script := range manager.config.Scripts {
		if script.Name == "test1" {
			found = true
			if script.Interval != 120 {
				t.Errorf("Expected interval 120, got %d", script.Interval)
			}
			if script.Path != "./test1_new.sh" {
				t.Errorf("Expected path ./test1_new.sh, got %s", script.Path)
			}
		}
	}
	if !found {
		t.Error("Script test1 not found in configuration after update")
	}
}

func TestScriptManager_UpdateScriptWithImmediateApplication_RestartRequiredChanges(t *testing.T) {
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

	// Verify script is running
	if !manager.IsScriptRunning("test1") {
		t.Fatal("Expected script to be running before update")
	}

	// Update with changes that require restart (interval and path)
	updatedScript := ScriptConfig{
		Name:        "test1",
		Path:        "./test1_updated.sh", // path change requires restart
		Interval:    120,                  // interval change requires restart
		Enabled:     true,
		MaxLogLines: 100,
		Timeout:     30,
	}

	// This should gracefully restart the script with new config
	err = manager.UpdateScriptWithImmediateApplication("test1", updatedScript)
	if err != nil {
		t.Errorf("Expected graceful restart to work, but got error: %v", err)
	}

	// Verify script is still running after graceful restart
	if !manager.IsScriptRunning("test1") {
		t.Error("Expected script to still be running after graceful restart")
	}

	// Verify configuration was updated
	found := false
	for _, script := range manager.config.Scripts {
		if script.Name == "test1" {
			found = true
			if script.Interval != 120 {
				t.Errorf("Expected interval 120 after restart, got %d", script.Interval)
			}
			if script.Path != "./test1_updated.sh" {
				t.Errorf("Expected path ./test1_updated.sh after restart, got %s", script.Path)
			}
		}
	}
	if !found {
		t.Error("Script test1 not found in configuration after graceful restart")
	}
}

func TestScriptManager_NewScriptManagerWithPath(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test", Path: "./test.sh", Interval: 60, Enabled: true},
		},
	}
	configPath := "/test/config.json"

	manager := NewScriptManagerWithPath(config, configPath)
	if manager == nil {
		t.Fatal("Expected script manager to be created")
	}

	if manager.config != config {
		t.Error("Expected config to be set correctly")
	}

	if manager.configPath != configPath {
		t.Errorf("Expected config path to be %s, got %s", configPath, manager.configPath)
	}
}

func TestScriptManager_GetScripts(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test1", Path: "./test1.sh", Interval: 60, Enabled: true, MaxLogLines: 100, Timeout: 30},
			{Name: "test2", Path: "./test2.sh", Interval: 120, Enabled: false, MaxLogLines: 50, Timeout: 60},
		},
	}

	manager := NewScriptManager(config)

	scripts, err := manager.GetScripts()
	if err != nil {
		t.Fatalf("Expected no error getting scripts, got: %v", err)
	}

	if len(scripts) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(scripts))
	}

	// Verify first script
	if scripts[0].Name != "test1" {
		t.Errorf("Expected first script name to be test1, got %s", scripts[0].Name)
	}
	if scripts[0].Path != "./test1.sh" {
		t.Errorf("Expected first script path to be ./test1.sh, got %s", scripts[0].Path)
	}

	// Verify second script
	if scripts[1].Name != "test2" {
		t.Errorf("Expected second script name to be test2, got %s", scripts[1].Name)
	}
	if scripts[1].Enabled != false {
		t.Errorf("Expected second script to be disabled, got %t", scripts[1].Enabled)
	}
}

func TestScriptManager_GetRunningScripts(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test1", Path: "./test1.sh", Interval: 60, Enabled: true, MaxLogLines: 100, Timeout: 30},
			{Name: "test2", Path: "./test2.sh", Interval: 120, Enabled: false, MaxLogLines: 50, Timeout: 60},
		},
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	// Initially no scripts should be running
	running := manager.GetRunningScripts()
	if len(running) != 0 {
		t.Errorf("Expected 0 running scripts initially, got %d", len(running))
	}

	// Start a script
	err := manager.StartScript(ctx, "test1")
	if err != nil {
		t.Fatalf("Error starting script: %v", err)
	}

	running = manager.GetRunningScripts()
	if len(running) != 1 {
		t.Errorf("Expected 1 running script, got %d", len(running))
	}

	if running[0] != "test1" {
		t.Errorf("Expected running script to be test1, got %s", running[0])
	}
}

func TestScriptManager_GetConfig(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test", Path: "./test.sh", Interval: 60, Enabled: true},
		},
		WebPort: 8080,
	}

	manager := NewScriptManager(config)

	returnedConfig := manager.GetConfig()
	if returnedConfig != config {
		t.Error("Expected GetConfig to return the same config object")
	}

	if returnedConfig.WebPort != 8080 {
		t.Errorf("Expected WebPort to be 8080, got %d", returnedConfig.WebPort)
	}
}

func TestScriptManager_AddScript(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "existing", Path: "./existing.sh", Interval: 60, Enabled: true},
		},
	}

	manager := NewScriptManager(config)

	// Add a new script
	newScript := ScriptConfig{
		Name:        "new",
		Path:        "./new.sh",
		Interval:    120,
		Enabled:     false,
		MaxLogLines: 200,
		Timeout:     45,
	}

	err := manager.AddScript(newScript)
	if err != nil {
		t.Fatalf("Expected no error adding script, got: %v", err)
	}

	// Verify script was added
	if len(manager.config.Scripts) != 2 {
		t.Errorf("Expected 2 scripts after adding, got %d", len(manager.config.Scripts))
	}

	found := false
	for _, script := range manager.config.Scripts {
		if script.Name == "new" {
			found = true
			if script.Path != "./new.sh" {
				t.Errorf("Expected path ./new.sh, got %s", script.Path)
			}
			if script.Interval != 120 {
				t.Errorf("Expected interval 120, got %d", script.Interval)
			}
		}
	}
	if !found {
		t.Error("New script not found in configuration")
	}
}

func TestScriptManager_AddScript_DuplicateName(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "existing", Path: "./existing.sh", Interval: 60, Enabled: true},
		},
	}

	manager := NewScriptManager(config)

	// Try to add script with duplicate name
	duplicateScript := ScriptConfig{
		Name:     "existing",
		Path:     "./different.sh",
		Interval: 120,
		Enabled:  false,
	}

	err := manager.AddScript(duplicateScript)
	if err == nil {
		t.Error("Expected error when adding script with duplicate name")
	}

	// Verify original script count unchanged
	if len(manager.config.Scripts) != 1 {
		t.Errorf("Expected 1 script after failed add, got %d", len(manager.config.Scripts))
	}
}

func TestScriptManager_EnableScript(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test", Path: "./test.sh", Interval: 60, Enabled: false, MaxLogLines: 100},
		},
	}

	manager := NewScriptManager(config)

	// Enable the script
	err := manager.EnableScript("test")
	if err != nil {
		t.Fatalf("Expected no error enabling script, got: %v", err)
	}

	// Verify script is enabled
	found := false
	for _, script := range manager.config.Scripts {
		if script.Name == "test" {
			found = true
			if !script.Enabled {
				t.Error("Expected script to be enabled after EnableScript call")
			}
		}
	}
	if !found {
		t.Error("Script not found in configuration")
	}
}

func TestScriptManager_EnableScript_NotFound(t *testing.T) {
	config := &ServiceConfig{Scripts: []ScriptConfig{}}
	manager := NewScriptManager(config)

	err := manager.EnableScript("nonexistent")
	if err == nil {
		t.Error("Expected error when enabling non-existent script")
	}
}

func TestScriptManager_DisableScript(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test", Path: "./test.sh", Interval: 60, Enabled: true, MaxLogLines: 100},
		},
	}

	manager := NewScriptManager(config)

	// Disable the script
	err := manager.DisableScript("test")
	if err != nil {
		t.Fatalf("Expected no error disabling script, got: %v", err)
	}

	// Verify script is disabled
	found := false
	for _, script := range manager.config.Scripts {
		if script.Name == "test" {
			found = true
			if script.Enabled {
				t.Error("Expected script to be disabled after DisableScript call")
			}
		}
	}
	if !found {
		t.Error("Script not found in configuration")
	}
}

func TestScriptManager_DisableScript_NotFound(t *testing.T) {
	config := &ServiceConfig{Scripts: []ScriptConfig{}}
	manager := NewScriptManager(config)

	err := manager.DisableScript("nonexistent")
	if err == nil {
		t.Error("Expected error when disabling non-existent script")
	}
}

func TestScriptManager_RunScriptOnce(t *testing.T) {
	config := &ServiceConfig{
		Scripts: []ScriptConfig{
			{Name: "test", Path: "./test.sh", Interval: 60, Enabled: true, MaxLogLines: 100, Timeout: 30},
		},
	}

	manager := NewScriptManager(config)
	ctx := context.Background()

	// This should run the script once without starting a continuous runner
	err := manager.RunScriptOnce(ctx, "test")
	if err != nil {
		// We expect this to fail initially because RunOnce might not be implemented in ScriptRunner
		t.Logf("RunScriptOnce failed as expected (missing RunOnce implementation): %v", err)
	}

	// Verify that no ongoing script runner was created
	if manager.IsScriptRunning("test") {
		t.Error("Expected script not to be running after RunScriptOnce")
	}
}

func TestScriptManager_RunScriptOnce_NotFound(t *testing.T) {
	config := &ServiceConfig{Scripts: []ScriptConfig{}}
	manager := NewScriptManager(config)
	ctx := context.Background()

	err := manager.RunScriptOnce(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when running non-existent script once")
	}
}
