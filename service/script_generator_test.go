package service

import (
	"fmt"
	"strings"
	"testing"
)

// TestScriptGenerator_GeneratePureScript tests generating a pure shell script
func TestScriptGenerator_GeneratePureScript(t *testing.T) {
	generator := NewScriptGenerator()

	template := &ScriptTemplate{
		Type:        "pure",
		Name:        "test-script",
		ProjectPath: "/path/to/project",
		Content:     "echo 'Hello World'\nls -la",
		Config: ScriptConfigData{
			Interval:    "5m",
			Timeout:     30,
			MaxLogLines: 100,
		},
	}

	result, err := generator.GenerateScript(template)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != template.Name {
		t.Errorf("Expected name %s, got %s", template.Name, result.Name)
	}

	if result.Path == "" {
		t.Error("Expected generated script path to be set")
	}

	if !strings.Contains(result.Content, "echo 'Hello World'") {
		t.Error("Expected script content to contain user content")
	}

	if !strings.Contains(result.Content, "#!/bin/bash") {
		t.Error("Expected script to start with bash shebang")
	}
}

// TestScriptGenerator_GenerateClaudeCodeScript tests generating a Claude Code script
func TestScriptGenerator_GenerateClaudeCodeScript(t *testing.T) {
	generator := NewScriptGenerator()

	template := &ScriptTemplate{
		Type:        "claude-code",
		Name:        "claude-test-script",
		ProjectPath: "/home/user/project",
		Prompts:     []string{"Fix any bugs", "Update dependencies", "Run tests"},
		Config: ScriptConfigData{
			Interval:    "1h",
			Timeout:     300,
			MaxLogLines: 200,
		},
	}

	result, err := generator.GenerateScript(template)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != template.Name {
		t.Errorf("Expected name %s, got %s", template.Name, result.Name)
	}

	if !strings.Contains(result.Content, "#!/bin/bash") {
		t.Error("Expected script to start with bash shebang")
	}

	if !strings.Contains(result.Content, "cd /home/user/project") {
		t.Error("Expected script to change to project directory")
	}

	if !strings.Contains(result.Content, "SKIP_CLAUDE_HOOKS=1") {
		t.Error("Expected script to set SKIP_CLAUDE_HOOKS environment variable")
	}

	// Check for all three prompts
	for i, prompt := range template.Prompts {
		phaseText := fmt.Sprintf("Phase %d", i+1)
		if !strings.Contains(result.Content, phaseText) {
			t.Errorf("Expected script to contain %s", phaseText)
		}

		if !strings.Contains(result.Content, prompt) {
			t.Errorf("Expected script to contain prompt: %s", prompt)
		}
	}

	// Check for Claude CLI command
	if !strings.Contains(result.Content, "/home/logan/.claude/local/claude") {
		t.Error("Expected script to contain Claude CLI path")
	}
}

// TestScriptGenerator_ValidateTemplate tests template validation
func TestScriptGenerator_ValidateTemplate(t *testing.T) {
	generator := NewScriptGenerator()

	// Test missing name
	template := &ScriptTemplate{
		Type: "pure",
		Name: "",
	}
	err := generator.ValidateTemplate(template)
	if err == nil {
		t.Error("Expected error for missing name")
	}

	// Test invalid type
	template = &ScriptTemplate{
		Type: "invalid",
		Name: "test",
	}
	err = generator.ValidateTemplate(template)
	if err == nil {
		t.Error("Expected error for invalid type")
	}

	// Test claude-code without project path
	template = &ScriptTemplate{
		Type:        "claude-code",
		Name:        "test",
		ProjectPath: "",
	}
	err = generator.ValidateTemplate(template)
	if err == nil {
		t.Error("Expected error for claude-code without project path")
	}

	// Test claude-code without prompts
	template = &ScriptTemplate{
		Type:        "claude-code",
		Name:        "test",
		ProjectPath: "/path",
		Prompts:     []string{},
	}
	err = generator.ValidateTemplate(template)
	if err == nil {
		t.Error("Expected error for claude-code without prompts")
	}

	// Test pure script without content
	template = &ScriptTemplate{
		Type:    "pure",
		Name:    "test",
		Content: "",
	}
	err = generator.ValidateTemplate(template)
	if err == nil {
		t.Error("Expected error for pure script without content")
	}

	// Test valid pure script
	template = &ScriptTemplate{
		Type:    "pure",
		Name:    "test",
		Content: "echo hello",
	}
	err = generator.ValidateTemplate(template)
	if err != nil {
		t.Errorf("Expected no error for valid pure script, got %v", err)
	}

	// Test valid claude-code script
	template = &ScriptTemplate{
		Type:        "claude-code",
		Name:        "test",
		ProjectPath: "/path",
		Prompts:     []string{"test prompt"},
	}
	err = generator.ValidateTemplate(template)
	if err != nil {
		t.Errorf("Expected no error for valid claude-code script, got %v", err)
	}
}

// TestScriptGenerator_TooManyPrompts tests prompt limit validation
func TestScriptGenerator_TooManyPrompts(t *testing.T) {
	generator := NewScriptGenerator()

	template := &ScriptTemplate{
		Type:        "claude-code",
		Name:        "test",
		ProjectPath: "/path",
		Prompts:     []string{"1", "2", "3", "4", "5", "6"}, // 6 prompts, limit is 5
	}

	err := generator.ValidateTemplate(template)
	if err == nil {
		t.Error("Expected error for too many prompts")
	}
}
