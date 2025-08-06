# Enhanced Script Creation Experience Plan

## Overview
Improve the script creation UI/UX by adding directory selection for Git projects, script type selection, and automatic script generation based on type and prompts.

## Current Problem
- Script creation requires manual path input
- No guidance for different script types
- No automation for Claude Code script generation
- Poor user experience for discovering available projects

## Requirements
- Directory selection UI showing home directory Git projects
- Script type selection (Pure Script vs Claude Code Script)
- Dynamic form based on script type selection
- Automatic script generation and file creation
- Support for multiple prompts (up to 5) for Claude Code scripts
- Integration with existing script management system

## Implementation Plan

### 1. Backend Enhancements

#### 1.1 Git Project Discovery API
```go
// New endpoint: GET /api/git-projects
type GitProject struct {
    Name        string `json:"name"`
    Path        string `json:"path"`
    Description string `json:"description,omitempty"`
    LastCommit  string `json:"last_commit,omitempty"`
}

func (ws *WebServer) handleGitProjects(c *gin.Context) {
    projects, err := discoverGitProjects(homeDir)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"projects": projects})
}
```

#### 1.2 Script Generation Service
```go
type ScriptTemplate struct {
    Type        string            `json:"type"`
    Name        string            `json:"name"`
    ProjectPath string            `json:"project_path"`
    Content     string            `json:"content,omitempty"`
    Prompts     []string          `json:"prompts,omitempty"`
    Config      ScriptConfigData  `json:"config"`
}

type ScriptGenerator interface {
    GenerateScript(template *ScriptTemplate) (*GeneratedScript, error)
    ValidateTemplate(template *ScriptTemplate) error
}
```

#### 1.3 Claude Code Script Generator
Based on `/home/logan/run-script-service/run-service-develop-plan-cycle.sh`:
```go
func (sg *ScriptGenerator) generateClaudeCodeScript(template *ScriptTemplate) string {
    script := `#!/bin/bash

# Auto-generated Claude Code Script
# Script Name: ` + template.Name + `
# Project Path: ` + template.ProjectPath + `
export SKIP_CLAUDE_HOOKS=1

echo "$(date): Starting ` + template.Name + `..."
cd ` + template.ProjectPath

    for i, prompt := range template.Prompts {
        script += fmt.Sprintf(`

# Phase %d: %s
echo "$(date): Phase %d - Executing..."
/home/logan/.claude/local/claude --dangerously-skip-permissions -p "%s" --output-format stream-json --verbose
PHASE%d_EXIT=$?
echo "$(date): Phase %d completed with exit code: $PHASE%d_EXIT"
`, i+1, prompt, i+1, prompt, i+1, i+1, i+1)
    }

    script += `

echo "$(date): ` + template.Name + ` completed successfully"
`
    return script
}
```

#### 1.4 Enhanced Script Creation API
```go
// Modified: POST /api/scripts
type CreateScriptRequest struct {
    Name        string   `json:"name" binding:"required"`
    Type        string   `json:"type" binding:"required"` // "pure" or "claude-code"
    ProjectPath string   `json:"project_path,omitempty"`
    Content     string   `json:"content,omitempty"`
    Prompts     []string `json:"prompts,omitempty"`
    Interval    string   `json:"interval" binding:"required"`
    Timeout     int      `json:"timeout"`
    MaxLogLines int      `json:"max_log_lines"`
}
```

### 2. Frontend Enhancements

#### 2.1 Git Projects Service
```typescript
// services/gitService.ts
export interface GitProject {
  name: string;
  path: string;
  description?: string;
  lastCommit?: string;
}

export class GitService {
  static async getGitProjects(): Promise<GitProject[]> {
    const response = await api.get('/api/git-projects');
    return response.data.projects;
  }
}
```

#### 2.2 Enhanced Script Creation Form
```vue
<!-- components/CreateScriptForm.vue -->
<template>
  <div class="create-script-form">
    <!-- Basic Info -->
    <div class="form-section">
      <h3>Basic Information</h3>
      <input v-model="form.name" placeholder="Script Name" required />
      <select v-model="form.interval" required>
        <option value="">Select Interval</option>
        <option value="5m">5 minutes</option>
        <option value="30m">30 minutes</option>
        <option value="1h">1 hour</option>
        <option value="6h">6 hours</option>
        <option value="24h">24 hours</option>
      </select>
    </div>

    <!-- Script Type Selection -->
    <div class="form-section">
      <h3>Script Type</h3>
      <div class="script-type-selector">
        <label class="type-option">
          <input type="radio" v-model="form.type" value="pure" />
          <div class="option-card">
            <h4>Pure Script</h4>
            <p>Traditional shell script with custom content</p>
          </div>
        </label>
        <label class="type-option">
          <input type="radio" v-model="form.type" value="claude-code" />
          <div class="option-card">
            <h4>Claude Code Script</h4>
            <p>AI-powered development workflow with prompts</p>
          </div>
        </label>
      </div>
    </div>

    <!-- Directory Selection (for Claude Code) -->
    <div v-if="form.type === 'claude-code'" class="form-section">
      <h3>Project Directory</h3>
      <div class="project-selector">
        <div
          v-for="project in gitProjects"
          :key="project.path"
          class="project-item"
          :class="{ active: form.projectPath === project.path }"
          @click="form.projectPath = project.path"
        >
          <div class="project-info">
            <h4>{{ project.name }}</h4>
            <p class="project-path">{{ project.path }}</p>
            <p v-if="project.lastCommit" class="last-commit">
              Last commit: {{ project.lastCommit }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Pure Script Content -->
    <div v-if="form.type === 'pure'" class="form-section">
      <h3>Script Content</h3>
      <textarea
        v-model="form.content"
        placeholder="#!/bin/bash&#10;&#10;echo 'Hello World'"
        rows="10"
        required
      ></textarea>
    </div>

    <!-- Claude Code Prompts -->
    <div v-if="form.type === 'claude-code'" class="form-section">
      <h3>AI Prompts</h3>
      <div class="prompts-container">
        <div
          v-for="(prompt, index) in form.prompts"
          :key="index"
          class="prompt-item"
        >
          <div class="prompt-header">
            <span class="prompt-number">Phase {{ index + 1 }}</span>
            <button
              type="button"
              class="remove-prompt"
              @click="removePrompt(index)"
              :disabled="form.prompts.length === 1"
            >
              ×
            </button>
          </div>
          <textarea
            v-model="form.prompts[index]"
            :placeholder="`Enter prompt for phase ${index + 1}...`"
            rows="3"
            required
          ></textarea>
        </div>

        <button
          type="button"
          class="add-prompt-btn"
          @click="addPrompt"
          :disabled="form.prompts.length >= 5"
          v-if="form.prompts.length < 5"
        >
          + Add Prompt ({{ form.prompts.length }}/5)
        </button>
      </div>
    </div>

    <!-- Advanced Settings -->
    <div class="form-section">
      <h3>Advanced Settings</h3>
      <div class="advanced-settings">
        <div class="setting-item">
          <label>Timeout (seconds)</label>
          <input type="number" v-model.number="form.timeout" min="0" />
          <small>0 = no timeout</small>
        </div>
        <div class="setting-item">
          <label>Max Log Lines</label>
          <input type="number" v-model.number="form.maxLogLines" min="50" />
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="form-actions">
      <button type="button" @click="$emit('cancel')" class="btn-secondary">
        Cancel
      </button>
      <button type="submit" @click="createScript" class="btn-primary">
        Create Script
      </button>
    </div>
  </div>
</template>
```

#### 2.3 Form Logic Implementation
```typescript
// composables/useScriptCreation.ts
export function useScriptCreation() {
  const form = reactive({
    name: '',
    type: 'pure',
    projectPath: '',
    content: '',
    prompts: [''],
    interval: '',
    timeout: 0,
    maxLogLines: 100
  });

  const gitProjects = ref<GitProject[]>([]);

  const loadGitProjects = async () => {
    try {
      gitProjects.value = await GitService.getGitProjects();
    } catch (error) {
      console.error('Failed to load Git projects:', error);
    }
  };

  const addPrompt = () => {
    if (form.prompts.length < 5) {
      form.prompts.push('');
    }
  };

  const removePrompt = (index: number) => {
    if (form.prompts.length > 1) {
      form.prompts.splice(index, 1);
    }
  };

  const createScript = async () => {
    try {
      const payload = {
        name: form.name,
        type: form.type,
        project_path: form.projectPath,
        content: form.content,
        prompts: form.prompts.filter(p => p.trim()),
        interval: form.interval,
        timeout: form.timeout,
        max_log_lines: form.maxLogLines
      };

      await api.post('/api/scripts', payload);
      // Success handling
    } catch (error) {
      // Error handling
    }
  };

  return {
    form,
    gitProjects,
    loadGitProjects,
    addPrompt,
    removePrompt,
    createScript
  };
}
```

### 3. File Structure

```
service/
├── git_discovery.go     # New: Git project discovery
├── script_generator.go  # New: Script generation logic
└── templates/           # New: Script templates
    ├── pure_script.sh
    └── claude_code_script.sh

web/
├── handlers_scripts.go  # Enhanced with new endpoints
└── frontend/src/
    ├── services/
    │   └── gitService.ts        # New: Git projects API
    ├── components/
    │   └── CreateScriptForm.vue # Enhanced form component
    ├── composables/
    │   └── useScriptCreation.ts # New: Form logic
    └── views/
        └── Scripts.vue          # Updated to use new form
```

### 4. Implementation Steps

#### Phase 1: Backend Foundation
1. **Git Discovery Service**
   - Implement Git project scanning in home directory
   - Add caching for project list
   - Create API endpoint for project retrieval

2. **Script Generator**
   - Create script template system
   - Implement Claude Code script generation
   - Add validation for script templates

#### Phase 2: API Enhancements
1. **Enhanced Script Creation**
   - Modify existing script creation endpoint
   - Add support for different script types
   - Implement automatic file generation

2. **Git Projects API**
   - Add new endpoint for Git project discovery
   - Implement project metadata extraction
   - Add error handling for inaccessible projects

#### Phase 3: Frontend Redesign
1. **Form Components**
   - Create enhanced script creation form
   - Implement dynamic form sections
   - Add visual script type selection

2. **Project Selection UI**
   - Build Git project selection interface
   - Add project information display
   - Implement selection state management

#### Phase 4: Integration & Testing
1. **End-to-End Integration**
   - Connect frontend form to backend APIs
   - Test script generation flow
   - Verify created scripts work correctly

2. **User Experience Polish**
   - Add loading states and feedback
   - Implement form validation
   - Add helpful tooltips and guidance

### 5. Script Generation Templates

#### 5.1 Pure Script Template
```bash
#!/bin/bash

# Auto-generated Script: {{.Name}}
# Created: {{.CreatedAt}}

{{.Content}}
```

#### 5.2 Claude Code Script Template
```bash
#!/bin/bash

# Auto-generated Claude Code Script
# Script Name: {{.Name}}
# Project Path: {{.ProjectPath}}
export SKIP_CLAUDE_HOOKS=1

echo "$(date): Starting {{.Name}}..."
cd {{.ProjectPath}}

{{range $i, $prompt := .Prompts}}
# Phase {{add $i 1}}: {{$prompt}}
echo "$(date): Phase {{add $i 1}} - Executing..."
/home/logan/.claude/local/claude --dangerously-skip-permissions -p "{{$prompt}}" --output-format stream-json --verbose
PHASE{{add $i 1}}_EXIT=$?
echo "$(date): Phase {{add $i 1}} completed with exit code: $PHASE{{add $i 1}}_EXIT"
{{end}}

echo "$(date): {{.Name}} completed successfully"
```

### 6. UI/UX Improvements

#### 6.1 Visual Design
- Clean, modern form layout
- Clear visual hierarchy
- Intuitive navigation between form sections
- Responsive design for different screen sizes

#### 6.2 User Guidance
- Tooltips explaining script types
- Example prompts for Claude Code scripts
- Visual feedback for form validation
- Progress indicators during script creation

#### 6.3 Error Handling
- Clear error messages for validation failures
- Helpful suggestions for fixing issues
- Recovery options for failed operations

### 7. Configuration & Settings

#### 7.1 Git Discovery Settings
```json
{
  "git_discovery": {
    "home_directory": "/home/logan",
    "max_depth": 3,
    "exclude_patterns": [".git", "node_modules", "vendor"],
    "cache_duration": "5m"
  }
}
```

#### 7.2 Script Generation Settings
```json
{
  "script_generation": {
    "template_directory": "./templates",
    "output_directory": "./generated_scripts",
    "max_prompts": 5,
    "default_timeout": 0
  }
}
```

### 8. Testing Strategy

#### 8.1 Backend Tests
- Git project discovery functionality
- Script generation with various templates
- API endpoints for enhanced script creation

#### 8.2 Frontend Tests
- Form component behavior
- Script type switching
- Prompt management (add/remove)
- Integration with backend APIs

#### 8.3 E2E Tests
- Complete script creation workflow
- Generated script execution
- Error handling scenarios

### 9. Migration & Compatibility

#### 9.1 Backward Compatibility
- Existing script creation API remains functional
- Generated scripts use same execution framework
- No changes to existing script configurations

#### 9.2 Migration Path
- New UI replaces existing simple form
- Option to edit existing scripts with new interface
- Gradual rollout with feature flags

## Success Criteria

- ✅ Git project discovery from home directory
- ✅ Visual script type selection (Pure vs Claude Code)
- ✅ Directory selection UI for Git projects
- ✅ Dynamic prompt management (up to 5 prompts)
- ✅ Automatic script generation based on templates
- ✅ Enhanced user experience with clear guidance
- ✅ Integration with existing script management
- ✅ Comprehensive testing coverage
- ✅ Backward compatibility maintained

## Benefits

1. **Improved User Experience**
   - Intuitive script creation process
   - Visual guidance for different script types
   - Automated file generation

2. **Better Project Integration**
   - Easy discovery of Git projects
   - Direct integration with development workflows
   - Context-aware script creation

3. **Enhanced Productivity**
   - Faster script creation process
   - Template-based generation
   - Reduced manual configuration

4. **Professional Interface**
   - Modern, clean UI design
   - Comprehensive form validation
   - Clear visual feedback
