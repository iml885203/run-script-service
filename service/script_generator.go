package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScriptTemplate represents the input data for script generation
type ScriptTemplate struct {
	Type        string           `json:"type"`                   // "pure" or "claude-code"
	Name        string           `json:"name"`                   // Script name
	ProjectPath string           `json:"project_path,omitempty"` // For claude-code scripts
	Content     string           `json:"content,omitempty"`      // For pure scripts
	Prompts     []string         `json:"prompts,omitempty"`      // For claude-code scripts
	Config      ScriptConfigData `json:"config"`                 // Script configuration
}

// ScriptConfigData represents script configuration settings
type ScriptConfigData struct {
	Interval    string `json:"interval"`
	Timeout     int    `json:"timeout"`
	MaxLogLines int    `json:"max_log_lines"`
}

// GeneratedScript represents the output of script generation
type GeneratedScript struct {
	Name    string           `json:"name"`
	Path    string           `json:"path"`
	Content string           `json:"content"`
	Config  ScriptConfigData `json:"config"`
}

// ScriptGenerator handles script generation from templates
type ScriptGenerator struct {
	outputDir string
}

// NewScriptGenerator creates a new script generator instance
func NewScriptGenerator() *ScriptGenerator {
	// Get current working directory as default output directory
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	return &ScriptGenerator{
		outputDir: dir,
	}
}

// GenerateScript generates a script from a template
func (sg *ScriptGenerator) GenerateScript(template *ScriptTemplate) (*GeneratedScript, error) {
	// Validate template first
	if err := sg.ValidateTemplate(template); err != nil {
		return nil, fmt.Errorf("template validation failed: %v", err)
	}

	var content string
	var err error

	switch template.Type {
	case "pure":
		content, err = sg.generatePureScript(template)
	case "claude-code":
		content, err = sg.generateClaudeCodeScript(template)
	default:
		return nil, fmt.Errorf("unsupported script type: %s", template.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("script generation failed: %v", err)
	}

	// Generate script file path
	scriptPath := filepath.Join(sg.outputDir, fmt.Sprintf("%s.sh", template.Name))

	result := &GeneratedScript{
		Name:    template.Name,
		Path:    scriptPath,
		Content: content,
		Config:  template.Config,
	}

	return result, nil
}

// ValidateTemplate validates a script template
func (sg *ScriptGenerator) ValidateTemplate(template *ScriptTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	if template.Name == "" {
		return fmt.Errorf("script name is required")
	}

	if template.Type == "" {
		return fmt.Errorf("script type is required")
	}

	switch template.Type {
	case "pure":
		if template.Content == "" {
			return fmt.Errorf("content is required for pure scripts")
		}
	case "claude-code":
		if template.ProjectPath == "" {
			return fmt.Errorf("project path is required for Claude Code scripts")
		}
		if len(template.Prompts) == 0 {
			return fmt.Errorf("at least one prompt is required for Claude Code scripts")
		}
		if len(template.Prompts) > 5 {
			return fmt.Errorf("maximum of 5 prompts allowed for Claude Code scripts")
		}
		// Validate prompts are not empty
		for i, prompt := range template.Prompts {
			if strings.TrimSpace(prompt) == "" {
				return fmt.Errorf("prompt %d cannot be empty", i+1)
			}
		}
	default:
		return fmt.Errorf("unsupported script type: %s (must be 'pure' or 'claude-code')", template.Type)
	}

	return nil
}

// generatePureScript generates a pure shell script
func (sg *ScriptGenerator) generatePureScript(template *ScriptTemplate) (string, error) {
	script := `#!/bin/bash

# Auto-generated Script: ` + template.Name + `
# Type: Pure Shell Script

` + template.Content + `
`
	return script, nil
}

// generateClaudeCodeScript generates a Claude Code automation script
func (sg *ScriptGenerator) generateClaudeCodeScript(template *ScriptTemplate) (string, error) {
	script := `#!/bin/bash

# Auto-generated Claude Code Script
# Script Name: ` + template.Name + `
# Project Path: ` + template.ProjectPath + `
export SKIP_CLAUDE_HOOKS=1

echo "$(date): Starting ` + template.Name + `..."
cd ` + template.ProjectPath + `
`

	// Add each prompt as a phase
	for i, prompt := range template.Prompts {
		phaseNum := i + 1
		script += fmt.Sprintf(`
# Phase %d: %s
echo "$(date): Phase %d - Executing..."
/home/logan/.claude/local/claude --dangerously-skip-permissions -p "%s" --output-format stream-json --verbose
PHASE%d_EXIT=$?
echo "$(date): Phase %d completed with exit code: $PHASE%d_EXIT"
`, phaseNum, prompt, phaseNum, prompt, phaseNum, phaseNum, phaseNum)
	}

	script += `
echo "$(date): ` + template.Name + ` completed successfully"
`

	return script, nil
}
