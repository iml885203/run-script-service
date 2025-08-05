<template>
  <div class="settings">
    <h2>System Settings</h2>

    <div v-if="loading" class="loading">
      Loading configuration...
    </div>

    <div v-else-if="error" class="error">
      {{ error }}
    </div>

    <div v-else class="settings-form">
      <form @submit.prevent="handleUpdateConfig">
        <div class="form-section">
          <h3>Web Interface</h3>

          <div class="form-group">
            <label for="webPort">Web Server Port:</label>
            <input
              v-model.number="configForm.webPort"
              type="number"
              id="webPort"
              min="1024"
              max="65535"
              required
            />
            <small class="help-text">Port for the web interface (requires restart)</small>
          </div>
        </div>

        <div class="form-section">
          <h3>Script Execution</h3>

          <div class="form-group">
            <label for="interval">Default Execution Interval:</label>
            <input
              v-model="configForm.interval"
              type="text"
              id="interval"
              placeholder="1h"
              required
            />
            <small class="help-text">Default interval for new scripts (e.g., 30s, 5m, 1h)</small>
          </div>
        </div>

        <div class="form-section">
          <h3>Logging</h3>

          <div class="form-group">
            <label for="logRetention">Log Retention (lines):</label>
            <input
              v-model.number="configForm.logRetention"
              type="number"
              id="logRetention"
              min="10"
              max="10000"
              required
            />
            <small class="help-text">Number of log lines to keep in memory</small>
          </div>
        </div>

        <div class="form-section">
          <h3>User Interface</h3>

          <div class="form-group">
            <label>
              <input
                v-model="configForm.autoRefresh"
                type="checkbox"
              />
              Enable auto-refresh
            </label>
            <small class="help-text">Automatically refresh data every 30 seconds</small>
          </div>
        </div>

        <div class="form-actions">
          <button type="button" @click="resetForm" class="btn btn-secondary">
            Reset
          </button>
          <button type="submit" :disabled="loading || !hasChanges" class="btn btn-primary">
            Save Settings
          </button>
        </div>
      </form>
    </div>

    <div class="system-info">
      <h3>System Information</h3>
      <div class="info-grid">
        <div class="info-item">
          <label>Version:</label>
          <span>1.0.0</span>
        </div>
        <div class="info-item">
          <label>Platform:</label>
          <span>{{ navigator.platform }}</span>
        </div>
        <div class="info-item">
          <label>User Agent:</label>
          <span class="small-text">{{ navigator.userAgent }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ApiService } from '@/services/api'
import type { ServiceConfig } from '@/types/api'

const loading = ref(false)
const error = ref<string | null>(null)
const originalConfig = ref<ServiceConfig | null>(null)
const configForm = ref<ServiceConfig>({
  webPort: 8080,
  interval: '1h',
  logRetention: 100,
  autoRefresh: true
})

const hasChanges = computed(() => {
  if (!originalConfig.value) return false
  return JSON.stringify(configForm.value) !== JSON.stringify(originalConfig.value)
})

// Access navigator globally
const navigator = window.navigator

const fetchConfig = async () => {
  loading.value = true
  error.value = null

  try {
    const config = await ApiService.getConfig()
    originalConfig.value = { ...config }
    configForm.value = { ...config }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch configuration'
    console.error('Failed to fetch config:', err)
  } finally {
    loading.value = false
  }
}

const handleUpdateConfig = async () => {
  loading.value = true
  error.value = null

  try {
    await ApiService.updateConfig(configForm.value)
    originalConfig.value = { ...configForm.value }

    // Show success message (could be enhanced with a toast system)
    console.log('Configuration updated successfully')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update configuration'
    console.error('Failed to update config:', err)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  if (originalConfig.value) {
    configForm.value = { ...originalConfig.value }
  }
}

// Watch for changes and warn before leaving if there are unsaved changes
watch(hasChanges, (newValue) => {
  if (newValue) {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      e.returnValue = ''
    }
    window.addEventListener('beforeunload', handleBeforeUnload)

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }
})

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.settings-form {
  max-width: 600px;
}

.form-section {
  margin-bottom: 3rem;
  padding: 1.5rem;
  background: var(--color-background-soft);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}

.form-section h3 {
  margin: 0 0 1.5rem 0;
  color: var(--color-text);
  font-size: 1.125rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--color-text);
}

.form-group input[type="checkbox"] {
  width: auto;
  margin-right: 0.5rem;
}

.help-text {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--color-text-muted);
  line-height: 1.4;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid var(--color-border);
}

.system-info {
  margin-top: 3rem;
  padding: 1.5rem;
  background: var(--color-background-soft);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}

.system-info h3 {
  margin: 0 0 1.5rem 0;
  color: var(--color-text);
  font-size: 1.125rem;
}

.info-grid {
  display: grid;
  gap: 1rem;
}

.info-item {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.info-item label {
  font-weight: 600;
  color: var(--color-text);
  min-width: 100px;
  margin: 0;
}

.info-item span {
  color: var(--color-text-muted);
  flex: 1;
}

.small-text {
  font-size: 0.75rem;
  word-break: break-all;
}

@media (max-width: 768px) {
  .form-actions {
    flex-direction: column;
  }

  .info-item {
    flex-direction: column;
    gap: 0.25rem;
  }

  .info-item label {
    min-width: auto;
  }
}
</style>
