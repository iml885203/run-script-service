<template>
  <div class="create-script-form">
    <!-- Basic Info -->
    <div class="form-section">
      <h3>Basic Information</h3>
      <input
        v-model="form.name"
        placeholder="Script Name"
        required
        class="form-input"
      />
      <select v-model="form.interval" required class="form-select">
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
        class="form-textarea"
      ></textarea>
    </div>

    <!-- Claude Code Prompts -->
    <div v-if="form.type === 'claude-code'" class="form-section">
      <h3>AI Prompts</h3>
      <div class="prompts-container">
        <div
          v-for="(_, index) in form.prompts"
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
            class="form-textarea"
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
          <input type="number" v-model.number="form.timeout" min="0" class="form-input" />
          <small>0 = no timeout</small>
        </div>
        <div class="setting-item">
          <label>Max Log Lines</label>
          <input type="number" v-model.number="form.maxLogLines" min="50" class="form-input" />
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="form-actions">
      <button type="button" @click="$emit('cancel')" class="btn-secondary">
        Cancel
      </button>
      <button type="button" @click="createScript" class="btn-primary">
        Create Script
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, watch } from 'vue'
import { GitService, type GitProject } from '@/services/gitService'

// Emits
const emit = defineEmits<{
  create: [payload: ScriptCreatePayload]
  cancel: []
}>()

// Types
interface ScriptCreatePayload {
  name: string
  type: string
  project_path: string
  content: string
  prompts: string[]
  interval: string
  timeout: number
  max_log_lines: number
}

// Form data
const form = reactive({
  name: '',
  type: 'pure',
  projectPath: '',
  content: '',
  prompts: [''],
  interval: '',
  timeout: 0,
  maxLogLines: 100
})

// Git projects
const gitProjects = ref<GitProject[]>([])

// Load git projects when component mounts
onMounted(async () => {
  await loadGitProjects()
})

// Load git projects when type changes to claude-code
watch(() => form.type, async (newType) => {
  if (newType === 'claude-code' && gitProjects.value.length === 0) {
    await loadGitProjects()
  }
})

// Methods
const loadGitProjects = async () => {
  try {
    gitProjects.value = await GitService.getGitProjects()
  } catch (error) {
    console.error('Failed to load Git projects:', error)
  }
}

const addPrompt = () => {
  if (form.prompts.length < 5) {
    form.prompts.push('')
  }
}

const removePrompt = (index: number) => {
  if (form.prompts.length > 1) {
    form.prompts.splice(index, 1)
  }
}

const createScript = () => {
  const payload: ScriptCreatePayload = {
    name: form.name,
    type: form.type,
    project_path: form.projectPath,
    content: form.content,
    prompts: form.prompts.filter(p => p.trim()),
    interval: form.interval,
    timeout: form.timeout,
    max_log_lines: form.maxLogLines
  }

  emit('create', payload)
}
</script>

<style scoped>
.create-script-form {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
}

.form-section {
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.form-section:last-of-type {
  border-bottom: none;
}

.form-section h3 {
  margin: 0 0 1rem 0;
  color: #374151;
  font-size: 1.125rem;
  font-weight: 600;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  margin-bottom: 1rem;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.script-type-selector {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.type-option {
  cursor: pointer;
  display: block;
}

.type-option input[type="radio"] {
  display: none;
}

.option-card {
  padding: 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.type-option input[type="radio"]:checked + .option-card {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.option-card h4 {
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
}

.option-card p {
  margin: 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.project-selector {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
}

.project-item {
  padding: 1rem;
  border-bottom: 1px solid #e5e7eb;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.project-item:last-child {
  border-bottom: none;
}

.project-item:hover {
  background-color: #f9fafb;
}

.project-item.active {
  background-color: #eff6ff;
  border-color: #3b82f6;
}

.project-info h4 {
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
}

.project-path {
  margin: 0 0 0.25rem 0;
  font-size: 0.875rem;
  color: #6b7280;
  font-family: 'Courier New', monospace;
}

.last-commit {
  margin: 0;
  font-size: 0.75rem;
  color: #9ca3af;
}

.prompts-container {
  space-y: 1rem;
}

.prompt-item {
  margin-bottom: 1rem;
}

.prompt-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.prompt-number {
  font-weight: 600;
  color: #374151;
}

.remove-prompt {
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-prompt:disabled {
  background: #d1d5db;
  cursor: not-allowed;
}

.add-prompt-btn {
  background: #10b981;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
}

.add-prompt-btn:disabled {
  background: #d1d5db;
  cursor: not-allowed;
}

.advanced-settings {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.setting-item {
  display: flex;
  flex-direction: column;
}

.setting-item label {
  font-weight: 500;
  margin-bottom: 0.5rem;
  color: #374151;
}

.setting-item small {
  margin-top: 0.25rem;
  color: #6b7280;
  font-size: 0.75rem;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.btn-secondary {
  background: #6b7280;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 1rem;
}

.btn-primary {
  background: #3b82f6;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 1rem;
}

.btn-secondary:hover,
.btn-primary:hover {
  opacity: 0.9;
}
</style>
