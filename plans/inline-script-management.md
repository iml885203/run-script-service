# Inline Script Management Feature Plan

## Overview

實現內建腳本管理功能，允許用戶直接在 Web 界面中創建、編輯和管理腳本，而不需要手動管理檔案路徑。腳本將統一存放在專案目錄下的 `scripts/` 資料夾中。

## Current State vs Desired State

### 🔴 **Current State**
```javascript
// 用戶需要提供完整路徑
{
  name: "my-script",
  path: "/full/path/to/script.sh",  // 手動路徑管理
  interval: 3600,
  enabled: true
}
```

### 🟢 **Desired State**
```javascript
// 統一管理，內容編輯
{
  name: "my-script",
  filename: "backup-task.sh",       // 只需檔案名稱
  content: "#!/bin/bash\n...",      // 直接編輯內容
  interval: 3600,
  enabled: true
}
```

## Problem Analysis

### 現有問題
1. **路徑管理複雜** - 用戶需要手動管理腳本檔案路徑
2. **外部檔案依賴** - 需要在檔案系統中預先創建腳本
3. **編輯不便** - 無法在 Web 介面中直接編輯腳本內容
4. **部署複雜** - 部署時需要確保腳本檔案存在
5. **版本控制混亂** - 腳本檔案可能散落各處

### 目標改善
1. **統一腳本存放** - 所有腳本存放在 `./scripts/` 目錄
2. **Web 介面管理** - 直接在瀏覽器中創建和編輯腳本
3. **自動檔案管理** - 系統自動處理檔案創建和更新
4. **Git 友好** - 腳本目錄加入 `.gitignore`，避免版本控制衝突

## Architecture Design

### 目錄結構
```
run-script-service/
├── scripts/                    # 新增：統一腳本目錄
│   ├── backup-task.sh         # 用戶創建的腳本
│   ├── data-sync.sh           # 用戶創建的腳本
│   └── monitoring.sh          # 用戶創建的腳本
├── .gitignore                 # 更新：忽略 scripts/ 目錄
├── service_config.json        # 更新：引用 scripts/ 下的檔案
└── ...
```

### API 設計

#### 1. 腳本 CRUD 操作
```go
// 新的腳本管理 API
type ScriptFile struct {
    Name     string `json:"name"`         // 腳本名稱
    Filename string `json:"filename"`     // 檔案名稱 (如: backup.sh)
    Content  string `json:"content"`      // 腳本內容
    Path     string `json:"path"`         // 自動生成: ./scripts/backup.sh
    Size     int64  `json:"size"`         // 檔案大小
    Modified time.Time `json:"modified"`  // 修改時間
}

// API 端點
POST   /api/scripts/files         # 創建新腳本檔案
GET    /api/scripts/files         # 列出所有腳本檔案
GET    /api/scripts/files/:name   # 獲取腳本內容
PUT    /api/scripts/files/:name   # 更新腳本內容
DELETE /api/scripts/files/:name   # 刪除腳本檔案
```

#### 2. 整合現有 Script 配置
```go
// 擴展現有 ScriptConfig
type ScriptConfig struct {
    Name        string `json:"name"`
    Filename    string `json:"filename"`    // 新增：檔案名稱
    Path        string `json:"path"`        // 自動生成
    Content     string `json:"content"`     // 新增：腳本內容（僅 API 回應使用）
    Interval    int    `json:"interval"`
    Enabled     bool   `json:"enabled"`
    MaxLogLines int    `json:"max_log_lines"`
    Timeout     int    `json:"timeout"`
}
```

### Frontend 設計

#### 1. 新的腳本創建表單
```vue
<template>
  <div class="create-script-form">
    <!-- 基本資訊 -->
    <div class="form-section">
      <h3>Script Information</h3>

      <div class="form-group">
        <label for="name">Script Name:</label>
        <input
          v-model="form.name"
          type="text"
          id="name"
          placeholder="Backup Task"
          required
        />
      </div>

      <div class="form-group">
        <label for="filename">File Name:</label>
        <input
          v-model="form.filename"
          type="text"
          id="filename"
          placeholder="backup-task.sh"
          pattern="^[a-zA-Z0-9._-]+\.sh$"
          required
        />
        <small>Must end with .sh extension</small>
      </div>
    </div>

    <!-- 腳本內容編輯器 -->
    <div class="form-section">
      <h3>Script Content</h3>

      <div class="editor-controls">
        <button type="button" @click="insertTemplate" class="btn-secondary">
          Insert Template
        </button>
        <span class="editor-info">
          Lines: {{ lineCount }} | Characters: {{ charCount }}
        </span>
      </div>

      <textarea
        v-model="form.content"
        class="script-editor"
        placeholder="#!/bin/bash&#10;&#10;# Enter your script content here&#10;echo 'Hello World'"
        rows="20"
        spellcheck="false"
        required
      ></textarea>
    </div>

    <!-- 配置設定 -->
    <div class="form-section">
      <h3>Execution Settings</h3>

      <div class="form-row">
        <div class="form-group">
          <label for="interval">Interval (seconds):</label>
          <input
            v-model.number="form.interval"
            type="number"
            id="interval"
            min="1"
            required
          />
        </div>

        <div class="form-group">
          <label for="timeout">Timeout (seconds):</label>
          <input
            v-model.number="form.timeout"
            type="number"
            id="timeout"
            min="0"
          />
          <small>0 = no timeout</small>
        </div>
      </div>

      <div class="form-group">
        <label>
          <input v-model="form.enabled" type="checkbox" />
          Enable script execution
        </label>
      </div>
    </div>

    <!-- 操作按鈕 -->
    <div class="form-actions">
      <button type="button" @click="$emit('cancel')" class="btn-secondary">
        Cancel
      </button>
      <button type="button" @click="saveScript" class="btn-primary">
        {{ isEditing ? 'Update Script' : 'Create Script' }}
      </button>
    </div>
  </div>
</template>
```

#### 2. 腳本列表增強
```vue
<template>
  <div class="scripts-list">
    <div v-for="script in scripts" :key="script.name" class="script-card">
      <div class="script-info">
        <h3>{{ script.name }}</h3>
        <p class="script-filename">{{ script.filename }}</p>
        <div class="script-details">
          <span class="file-size">{{ formatFileSize(script.size) }}</span>
          <span class="modified-time">{{ formatDate(script.modified) }}</span>
          <span class="interval">Interval: {{ script.interval }}s</span>
          <span :class="{ 'enabled': script.enabled, 'disabled': !script.enabled }">
            {{ script.enabled ? 'Enabled' : 'Disabled' }}
          </span>
        </div>
      </div>

      <div class="script-actions">
        <button @click="runScript(script.name)" class="btn btn-secondary">
          Run Now
        </button>
        <button @click="editScript(script)" class="btn btn-secondary">
          Edit Script
        </button>
        <button @click="viewContent(script)" class="btn btn-secondary">
          View Content
        </button>
        <button @click="toggleScript(script.name)" class="btn btn-secondary">
          {{ script.enabled ? 'Disable' : 'Enable' }}
        </button>
        <button @click="deleteScript(script.name)" class="btn btn-danger">
          Delete
        </button>
      </div>
    </div>
  </div>
</template>
```

#### 3. 腳本編輯器組件
```vue
<!-- ScriptEditor.vue -->
<template>
  <div class="script-editor-modal" v-if="visible">
    <div class="modal-backdrop" @click="$emit('close')"></div>

    <div class="editor-container">
      <div class="editor-header">
        <h2>{{ isEditing ? 'Edit Script' : 'New Script' }}: {{ scriptName }}</h2>
        <button @click="$emit('close')" class="close-btn">&times;</button>
      </div>

      <div class="editor-toolbar">
        <button @click="insertTemplate" class="toolbar-btn">
          📝 Template
        </button>
        <button @click="formatCode" class="toolbar-btn">
          🎨 Format
        </button>
        <button @click="validateSyntax" class="toolbar-btn">
          ✓ Validate
        </button>

        <div class="editor-stats">
          Lines: {{ lineCount }} | Size: {{ formatFileSize(content.length) }}
        </div>
      </div>

      <textarea
        v-model="content"
        class="code-editor"
        :placeholder="placeholder"
        spellcheck="false"
        @input="updateStats"
      ></textarea>

      <div class="editor-footer">
        <div class="validation-status" :class="validationStatus.type">
          {{ validationStatus.message }}
        </div>

        <div class="editor-actions">
          <button @click="$emit('cancel')" class="btn-secondary">
            Cancel
          </button>
          <button @click="saveScript" class="btn-primary" :disabled="!isValid">
            {{ isEditing ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
```

## Backend Implementation

### 1. 腳本檔案管理服務
```go
// service/script_file_manager.go
package service

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type ScriptFileManager struct {
    scriptsDir string
    mutex      sync.RWMutex
}

func NewScriptFileManager(baseDir string) *ScriptFileManager {
    scriptsDir := filepath.Join(baseDir, "scripts")

    // 確保 scripts 目錄存在
    if err := os.MkdirAll(scriptsDir, 0755); err != nil {
        log.Printf("Warning: Failed to create scripts directory: %v", err)
    }

    return &ScriptFileManager{
        scriptsDir: scriptsDir,
    }
}

// CreateScript 創建新腳本檔案
func (sfm *ScriptFileManager) CreateScript(filename, content string) error {
    sfm.mutex.Lock()
    defer sfm.mutex.Unlock()

    // 驗證檔案名稱
    if !strings.HasSuffix(filename, ".sh") {
        return fmt.Errorf("script filename must end with .sh extension")
    }

    if !isValidFilename(filename) {
        return fmt.Errorf("invalid filename format")
    }

    filePath := filepath.Join(sfm.scriptsDir, filename)

    // 檢查檔案是否已存在
    if _, err := os.Stat(filePath); err == nil {
        return fmt.Errorf("script file already exists: %s", filename)
    }

    // 創建檔案
    if err := os.WriteFile(filePath, []byte(content), 0755); err != nil {
        return fmt.Errorf("failed to create script file: %v", err)
    }

    return nil
}

// GetScript 獲取腳本內容
func (sfm *ScriptFileManager) GetScript(filename string) (*ScriptFile, error) {
    sfm.mutex.RLock()
    defer sfm.mutex.RUnlock()

    filePath := filepath.Join(sfm.scriptsDir, filename)

    // 讀取檔案資訊
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("script file not found: %s", filename)
    }

    // 讀取檔案內容
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read script file: %v", err)
    }

    return &ScriptFile{
        Filename: filename,
        Content:  string(content),
        Path:     filePath,
        Size:     fileInfo.Size(),
        Modified: fileInfo.ModTime(),
    }, nil
}

// UpdateScript 更新腳本內容
func (sfm *ScriptFileManager) UpdateScript(filename, content string) error {
    sfm.mutex.Lock()
    defer sfm.mutex.Unlock()

    filePath := filepath.Join(sfm.scriptsDir, filename)

    // 檢查檔案是否存在
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return fmt.Errorf("script file not found: %s", filename)
    }

    // 更新檔案內容
    if err := os.WriteFile(filePath, []byte(content), 0755); err != nil {
        return fmt.Errorf("failed to update script file: %v", err)
    }

    return nil
}

// ListScripts 列出所有腳本檔案
func (sfm *ScriptFileManager) ListScripts() ([]*ScriptFile, error) {
    sfm.mutex.RLock()
    defer sfm.mutex.RUnlock()

    var scripts []*ScriptFile

    err := filepath.WalkDir(sfm.scriptsDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if d.IsDir() || !strings.HasSuffix(d.Name(), ".sh") {
            return nil
        }

        fileInfo, err := d.Info()
        if err != nil {
            return err
        }

        scripts = append(scripts, &ScriptFile{
            Filename: d.Name(),
            Path:     path,
            Size:     fileInfo.Size(),
            Modified: fileInfo.ModTime(),
            // 不在列表中載入內容，按需載入
        })

        return nil
    })

    return scripts, err
}

// DeleteScript 刪除腳本檔案
func (sfm *ScriptFileManager) DeleteScript(filename string) error {
    sfm.mutex.Lock()
    defer sfm.mutex.Unlock()

    filePath := filepath.Join(sfm.scriptsDir, filename)

    if err := os.Remove(filePath); err != nil {
        return fmt.Errorf("failed to delete script file: %v", err)
    }

    return nil
}

// GetScriptPath 獲取腳本的完整路徑
func (sfm *ScriptFileManager) GetScriptPath(filename string) string {
    return filepath.Join(sfm.scriptsDir, filename)
}

// 驗證檔案名稱格式
func isValidFilename(filename string) bool {
    // 允許字母、數字、點、底線、連字符
    validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-"

    for _, char := range filename {
        if !strings.ContainsRune(validChars, char) {
            return false
        }
    }

    return len(filename) > 0 && len(filename) <= 255
}
```

### 2. Web API 端點
```go
// web/handlers_scripts.go - 新增腳本檔案管理端點

// handleGetScriptFiles 列出所有腳本檔案
func (ws *WebServer) handleGetScriptFiles(c *gin.Context) {
    if ws.scriptFileManager == nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Error:   "Script file manager not initialized",
        })
        return
    }

    scripts, err := ws.scriptFileManager.ListScripts()
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Error:   fmt.Sprintf("Failed to list scripts: %v", err),
        })
        return
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Data:    scripts,
    })
}

// handleGetScriptFile 獲取單個腳本檔案內容
func (ws *WebServer) handleGetScriptFile(c *gin.Context) {
    filename := c.Param("filename")

    if ws.scriptFileManager == nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Error:   "Script file manager not initialized",
        })
        return
    }

    script, err := ws.scriptFileManager.GetScript(filename)
    if err != nil {
        c.JSON(http.StatusNotFound, APIResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Data:    script,
    })
}

// handleCreateScriptFile 創建新腳本檔案
func (ws *WebServer) handleCreateScriptFile(c *gin.Context) {
    var request struct {
        Name     string `json:"name" binding:"required"`
        Filename string `json:"filename" binding:"required"`
        Content  string `json:"content" binding:"required"`
        Interval int    `json:"interval" binding:"required"`
        Enabled  bool   `json:"enabled"`
        Timeout  int    `json:"timeout"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    // 創建腳本檔案
    if err := ws.scriptFileManager.CreateScript(request.Filename, request.Content); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    // 創建腳本配置
    scriptConfig := service.ScriptConfig{
        Name:        request.Name,
        Filename:    request.Filename,
        Path:        ws.scriptFileManager.GetScriptPath(request.Filename),
        Interval:    request.Interval,
        Enabled:     request.Enabled,
        MaxLogLines: 100, // 預設值
        Timeout:     request.Timeout,
    }

    // 添加到腳本管理器
    if err := ws.scriptManager.AddScript(scriptConfig); err != nil {
        // 如果添加配置失敗，清理已創建的檔案
        ws.scriptFileManager.DeleteScript(request.Filename)

        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Error:   fmt.Sprintf("Failed to add script configuration: %v", err),
        })
        return
    }

    c.JSON(http.StatusCreated, APIResponse{
        Success: true,
        Message: "Script created successfully",
    })
}

// handleUpdateScriptFile 更新腳本檔案
func (ws *WebServer) handleUpdateScriptFile(c *gin.Context) {
    filename := c.Param("filename")

    var request struct {
        Content string `json:"content" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    if err := ws.scriptFileManager.UpdateScript(filename, request.Content); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Script updated successfully",
    })
}

// handleDeleteScriptFile 刪除腳本檔案
func (ws *WebServer) handleDeleteScriptFile(c *gin.Context) {
    filename := c.Param("filename")

    // 從腳本管理器中移除（通過檔案名查找）
    scripts, err := ws.scriptManager.GetScripts()
    if err == nil {
        for _, script := range scripts {
            if script.Filename == filename {
                ws.scriptManager.RemoveScript(script.Name)
                break
            }
        }
    }

    // 刪除檔案
    if err := ws.scriptFileManager.DeleteScript(filename); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Script deleted successfully",
    })
}
```

### 3. 路由註冊
```go
// web/server.go - 在 setupRoutes 中添加
func (ws *WebServer) setupRoutes() {
    // ... 現有路由 ...

    // 腳本檔案管理端點
    protected.GET("/script-files", ws.handleGetScriptFiles)
    protected.GET("/script-files/:filename", ws.handleGetScriptFile)
    protected.POST("/script-files", ws.handleCreateScriptFile)
    protected.PUT("/script-files/:filename", ws.handleUpdateScriptFile)
    protected.DELETE("/script-files/:filename", ws.handleDeleteScriptFile)
}
```

## Frontend Implementation

### 1. 腳本檔案服務
```typescript
// services/scriptFileService.ts
import type { ScriptFile, CreateScriptRequest } from '@/types/script'

export class ScriptFileService {
  private static readonly BASE_URL = '/api/script-files'

  static async listScriptFiles(): Promise<ScriptFile[]> {
    const response = await fetch(this.BASE_URL)
    const result = await response.json()

    if (!result.success) {
      throw new Error(result.error || 'Failed to fetch script files')
    }

    return result.data
  }

  static async getScriptFile(filename: string): Promise<ScriptFile> {
    const response = await fetch(`${this.BASE_URL}/${encodeURIComponent(filename)}`)
    const result = await response.json()

    if (!result.success) {
      throw new Error(result.error || 'Failed to fetch script file')
    }

    return result.data
  }

  static async createScriptFile(request: CreateScriptRequest): Promise<void> {
    const response = await fetch(this.BASE_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    })

    const result = await response.json()

    if (!result.success) {
      throw new Error(result.error || 'Failed to create script file')
    }
  }

  static async updateScriptFile(filename: string, content: string): Promise<void> {
    const response = await fetch(`${this.BASE_URL}/${encodeURIComponent(filename)}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ content }),
    })

    const result = await response.json()

    if (!result.success) {
      throw new Error(result.error || 'Failed to update script file')
    }
  }

  static async deleteScriptFile(filename: string): Promise<void> {
    const response = await fetch(`${this.BASE_URL}/${encodeURIComponent(filename)}`, {
      method: 'DELETE',
    })

    const result = await response.json()

    if (!result.success) {
      throw new Error(result.error || 'Failed to delete script file')
    }
  }
}
```

### 2. TypeScript 類型定義
```typescript
// types/script.ts
export interface ScriptFile {
  filename: string
  content?: string
  path: string
  size: number
  modified: string
}

export interface CreateScriptRequest {
  name: string
  filename: string
  content: string
  interval: number
  enabled: boolean
  timeout: number
}

export interface ScriptFormData {
  name: string
  filename: string
  content: string
  interval: number
  enabled: boolean
  timeout: number
}
```

### 3. Composable 邏輯
```typescript
// composables/useScriptFiles.ts
import { ref, computed } from 'vue'
import { ScriptFileService } from '@/services/scriptFileService'
import type { ScriptFile, CreateScriptRequest } from '@/types/script'

export function useScriptFiles() {
  const scriptFiles = ref<ScriptFile[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchScriptFiles = async () => {
    loading.value = true
    error.value = null

    try {
      scriptFiles.value = await ScriptFileService.listScriptFiles()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch script files'
    } finally {
      loading.value = false
    }
  }

  const getScriptFile = async (filename: string) => {
    try {
      return await ScriptFileService.getScriptFile(filename)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch script file'
      throw err
    }
  }

  const createScriptFile = async (request: CreateScriptRequest) => {
    loading.value = true
    error.value = null

    try {
      await ScriptFileService.createScriptFile(request)
      await fetchScriptFiles() // 重新載入列表
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create script file'
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateScriptFile = async (filename: string, content: string) => {
    loading.value = true
    error.value = null

    try {
      await ScriptFileService.updateScriptFile(filename, content)
      await fetchScriptFiles() // 重新載入列表
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update script file'
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteScriptFile = async (filename: string) => {
    loading.value = true
    error.value = null

    try {
      await ScriptFileService.deleteScriptFile(filename)
      await fetchScriptFiles() // 重新載入列表
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete script file'
      throw err
    } finally {
      loading.value = false
    }
  }

  // 檔案大小格式化
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'

    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 腳本模板
  const getScriptTemplate = (type: 'basic' | 'advanced' = 'basic'): string => {
    const templates = {
      basic: `#!/bin/bash

# Script Name: New Script
# Description: Enter script description here
# Author: $(whoami)
# Created: $(date)

set -e  # Exit on error

echo "Script started at $(date)"

# Add your script logic here
echo "Hello World"

echo "Script completed at $(date)"
`,
      advanced: `#!/bin/bash

# Script Name: Advanced Script
# Description: Advanced script template with error handling
# Author: $(whoami)
# Created: $(date)

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="${SCRIPT_DIR}/script.log"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# Error handling
trap 'log "ERROR: Script failed at line $LINENO"' ERR

# Main script logic
main() {
    log "Script started"

    # Add your script logic here
    log "Executing main tasks..."

    log "Script completed successfully"
}

# Run main function
main "$@"
`
    }

    return templates[type]
  }

  return {
    scriptFiles,
    loading,
    error,
    fetchScriptFiles,
    getScriptFile,
    createScriptFile,
    updateScriptFile,
    deleteScriptFile,
    formatFileSize,
    getScriptTemplate
  }
}
```

## Configuration Updates

### 1. .gitignore 更新
```bash
# 在 .gitignore 中添加
# User-created scripts (managed via web interface)
scripts/
```

### 2. 目錄初始化
```go
// main.go 中初始化
func main() {
    // ... 現有初始化邏輯 ...

    // 初始化腳本檔案管理器
    scriptFileManager := service.NewScriptFileManager(".")

    // 設置到 Web 服務器
    webServer.SetScriptFileManager(scriptFileManager)

    // ... 剩餘邏輯 ...
}
```

### 3. 服務配置擴展
```json
{
  "scripts": [
    {
      "name": "backup-task",
      "filename": "backup-task.sh",
      "path": "./scripts/backup-task.sh",
      "interval": 3600,
      "enabled": true,
      "max_log_lines": 100,
      "timeout": 0
    }
  ],
  "web_port": 8080,
  "script_management": {
    "scripts_directory": "./scripts",
    "auto_create_directory": true,
    "default_permissions": "0755",
    "max_file_size": 1048576
  }
}
```

## Migration Strategy

### Phase 1: Backend Infrastructure (2天)
1. **腳本檔案管理服務**
   - [ ] 實現 `ScriptFileManager`
   - [ ] 添加檔案 CRUD 操作
   - [ ] 實現檔案驗證和安全檢查
   - [ ] 添加併發安全機制

2. **Web API 端點**
   - [ ] 實現腳本檔案管理 API
   - [ ] 整合現有腳本配置系統
   - [ ] 添加錯誤處理和驗證
   - [ ] 實現路由註冊

3. **服務初始化**
   - [ ] 修改 main.go 初始化邏輯
   - [ ] 創建 scripts 目錄
   - [ ] 更新 .gitignore
   - [ ] 測試基礎功能

### Phase 2: Frontend Implementation (2天)
1. **服務層開發**
   - [ ] 實現 `ScriptFileService`
   - [ ] 創建 TypeScript 類型定義
   - [ ] 實現 `useScriptFiles` composable
   - [ ] 添加錯誤處理機制

2. **UI 組件開發**
   - [ ] 創建腳本編輯器組件
   - [ ] 更新腳本創建表單
   - [ ] 實現腳本列表顯示
   - [ ] 添加操作按鈕和確認對話框

3. **用戶體驗優化**
   - [ ] 添加腳本模板功能
   - [ ] 實現語法高亮（可選）
   - [ ] 添加檔案大小和修改時間顯示
   - [ ] 實現拖放檔案上傳（可選）

### Phase 3: Integration & Testing (1天)
1. **系統整合**
   - [ ] 整合前後端功能
   - [ ] 測試腳本創建和執行流程
   - [ ] 驗證檔案系統操作
   - [ ] 測試錯誤處理場景

2. **向後相容性**
   - [ ] 支援現有外部腳本檔案
   - [ ] 提供遷移工具（可選）
   - [ ] 驗證配置檔案格式
   - [ ] 測試升級場景

### Phase 4: Polish & Documentation (0.5天)
1. **用戶體驗優化**
   - [ ] 添加載入狀態和進度指示
   - [ ] 實現操作確認和回饋訊息
   - [ ] 優化響應式設計
   - [ ] 添加鍵盤快捷鍵支援

2. **文檔和範例**
   - [ ] 更新使用者指南
   - [ ] 提供腳本範例
   - [ ] 添加 API 文檔
   - [ ] 創建遷移指南

## Testing Strategy

### 1. 單元測試
```go
// service/script_file_manager_test.go
func TestScriptFileManager_CreateScript(t *testing.T) {
    tmpDir := t.TempDir()
    manager := NewScriptFileManager(tmpDir)

    content := "#!/bin/bash\necho 'test'"

    err := manager.CreateScript("test.sh", content)
    assert.NoError(t, err)

    // 驗證檔案是否創建
    filePath := filepath.Join(tmpDir, "scripts", "test.sh")
    assert.FileExists(t, filePath)

    // 驗證檔案內容
    fileContent, err := os.ReadFile(filePath)
    assert.NoError(t, err)
    assert.Equal(t, content, string(fileContent))
}

func TestScriptFileManager_ValidateFilename(t *testing.T) {
    tests := []struct {
        filename string
        valid    bool
    }{
        {"valid-script.sh", true},
        {"valid_script.sh", true},
        {"valid.123.sh", true},
        {"invalid/script.sh", false},
        {"invalid<script.sh", false},
        {"script.txt", false},
        {"", false},
    }

    for _, test := range tests {
        t.Run(test.filename, func(t *testing.T) {
            result := isValidFilename(test.filename) && strings.HasSuffix(test.filename, ".sh")
            assert.Equal(t, test.valid, result)
        })
    }
}
```

### 2. 整合測試
```typescript
// tests/integration/scriptFiles.test.ts
import { test, expect } from '@playwright/test'

test.describe('Script File Management', () => {
  test('should create new script file', async ({ page }) => {
    await page.goto('/scripts')

    // 點擊創建腳本按鈕
    await page.click('[data-testid="create-script-btn"]')

    // 填寫表單
    await page.fill('[data-testid="script-name"]', 'Test Script')
    await page.fill('[data-testid="script-filename"]', 'test-script.sh')
    await page.fill('[data-testid="script-content"]', '#!/bin/bash\necho "Hello World"')
    await page.fill('[data-testid="script-interval"]', '3600')

    // 提交表單
    await page.click('[data-testid="create-script-submit"]')

    // 驗證腳本出現在列表中
    await expect(page.locator('[data-testid="script-list"]')).toContainText('Test Script')
    await expect(page.locator('[data-testid="script-list"]')).toContainText('test-script.sh')
  })

  test('should edit existing script', async ({ page }) => {
    // 先創建腳本（使用 API 或前面的測試）
    // ...

    // 點擊編輯按鈕
    await page.click('[data-testid="edit-script-btn"]')

    // 修改內容
    await page.fill('[data-testid="script-content"]', '#!/bin/bash\necho "Modified Script"')

    // 保存更改
    await page.click('[data-testid="save-script-btn"]')

    // 驗證更改已保存
    await expect(page.locator('[data-testid="success-message"]')).toBeVisible()
  })
})
```

### 3. API 測試
```go
// web/handlers_scripts_test.go
func TestCreateScriptFile(t *testing.T) {
    // 設置測試環境
    tmpDir := t.TempDir()
    server := setupTestServer(tmpDir)

    request := map[string]interface{}{
        "name":     "test-script",
        "filename": "test.sh",
        "content":  "#!/bin/bash\necho 'test'",
        "interval": 3600,
        "enabled":  true,
        "timeout":  0,
    }

    // 發送請求
    resp, err := makeAuthenticatedRequest("POST", "/api/script-files", request, server)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // 驗證檔案創建
    scriptPath := filepath.Join(tmpDir, "scripts", "test.sh")
    assert.FileExists(t, scriptPath)

    // 驗證配置添加
    scripts, err := server.scriptManager.GetScripts()
    assert.NoError(t, err)
    assert.Len(t, scripts, 1)
    assert.Equal(t, "test-script", scripts[0].Name)
}
```

## Security Considerations

### 1. 檔案系統安全
- **路徑驗證**: 防止目錄遍歷攻擊
- **檔案名稱限制**: 只允許安全的字符
- **檔案大小限制**: 防止磁盤空間攻擊
- **權限控制**: 適當的檔案權限設定

### 2. 腳本內容安全
- **內容檢查**: 檢測潛在危險命令（可選）
- **沙箱執行**: 考慮腳本執行環境隔離
- **日誌審計**: 記錄所有檔案操作

### 3. API 安全
- **認證檢查**: 所有 API 需要認證
- **輸入驗證**: 嚴格的輸入格式檢查
- **速率限制**: 防止 API 濫用

## Expected Benefits

### 1. **用戶體驗改善**
- ✅ **統一管理**: 所有腳本集中在一個位置
- ✅ **直接編輯**: 無需離開 Web 介面
- ✅ **即時預覽**: 立即看到腳本內容
- ✅ **模板支援**: 快速開始腳本開發

### 2. **開發效率提升**
- ✅ **快速創建**: 幾分鐘內創建新腳本
- ✅ **版本控制友好**: 腳本目錄與代碼分離
- ✅ **部署簡化**: 無需手動管理腳本檔案
- ✅ **備份容易**: 統一目錄易於備份

### 3. **維護性改善**
- ✅ **代碼組織**: 清晰的目錄結構
- ✅ **錯誤減少**: 自動路徑管理
- ✅ **監控友好**: 統一的檔案操作日誌
- ✅ **擴展性**: 易於添加更多檔案管理功能

### 4. **安全性提升**
- ✅ **權限控制**: 通過 Web 介面控制存取
- ✅ **審計追蹤**: 所有操作都有日誌記錄
- ✅ **隔離**: 腳本與系統檔案分離
- ✅ **驗證**: 嚴格的輸入驗證和格式檢查

## Future Enhancements

### 1. **進階編輯功能**
- 語法高亮和自動完成
- 代碼折疊和行號顯示
- 搜索和替換功能
- 多標籤頁編輯

### 2. **協作功能**
- 腳本版本歷史
- 多用戶編輯鎖定
- 變更追蹤和比較
- 註解和討論功能

### 3. **進階管理**
- 腳本分類和標籤
- 搜索和過濾功能
- 批量操作支援
- 導入/導出功能

### 4. **監控集成**
- 腳本執行統計
- 效能分析
- 錯誤追蹤
- 使用情況報告

這個功能將大幅改善腳本管理的用戶體驗，並為未來的擴展奠定基礎。
