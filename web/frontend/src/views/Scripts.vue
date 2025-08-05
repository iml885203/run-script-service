<template>
  <div class="scripts">
    <h2>Script Management</h2>

    <div v-if="loading" class="loading">
      Loading scripts...
    </div>

    <div v-else-if="error" class="error">
      {{ error }}
    </div>

    <div v-else>
      <div class="scripts-header">
        <button @click="showAddForm = true" class="btn btn-primary" data-testid="add-script-button">
          Add New Script
        </button>
      </div>

      <div v-if="scripts.length === 0" class="no-scripts">
        <p data-testid="no-scripts-message">No scripts configured yet.</p>
      </div>

      <div v-else class="scripts-list">
        <div v-for="script in scripts" :key="script.name" class="script-card" data-testid="script-card">
          <div class="script-info">
            <h3 data-testid="script-name">{{ script.name }}</h3>
            <p class="script-path" data-testid="script-path">{{ script.path }}</p>
            <div class="script-details">
              <span class="detail" data-testid="script-interval">
                <strong>Interval:</strong> {{ script.interval }}s
              </span>
              <span class="detail" data-testid="script-status">
                <strong>Status:</strong>
                <span :class="{ 'enabled': script.enabled, 'disabled': !script.enabled }">
                  {{ script.enabled ? 'Enabled' : 'Disabled' }}
                </span>
              </span>
              <span v-if="script.timeout" class="detail">
                <strong>Timeout:</strong> {{ script.timeout }}s
              </span>
            </div>
          </div>

          <div class="script-actions">
            <button @click="runScript(script.name)" :disabled="loading" class="btn btn-secondary btn-sm">
              Run Now
            </button>
            <button @click="toggleScript(script.name)" :disabled="loading" class="btn btn-secondary btn-sm">
              {{ script.enabled ? 'Disable' : 'Enable' }}
            </button>
            <button @click="editScript(script)" class="btn btn-secondary btn-sm">
              Edit
            </button>
            <button @click="deleteScript(script.name)" class="btn btn-danger btn-sm">
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Script Modal (simplified) -->
    <div v-if="showAddForm" class="modal-overlay" @click="showAddForm = false">
      <div class="modal" @click.stop>
        <h3>Add New Script</h3>
        <form @submit.prevent="handleAddScript">
          <div class="form-group">
            <label for="path">Script Path:</label>
            <input
              v-model="newScript.path"
              type="text"
              id="path"
              placeholder="/path/to/script.sh"
              required
            />
          </div>
          <div class="form-group">
            <label for="interval">Interval (seconds):</label>
            <input
              v-model.number="newScript.interval"
              type="number"
              id="interval"
              min="1"
              required
            />
          </div>
          <div class="form-group">
            <label for="timeout">Timeout (seconds, optional):</label>
            <input
              v-model.number="newScript.timeout"
              type="number"
              id="timeout"
              min="1"
            />
          </div>
          <div class="form-group">
            <label>
              <input v-model="newScript.enabled" type="checkbox" />
              Enable script
            </label>
          </div>
          <div class="modal-actions">
            <button type="button" @click="showAddForm = false" class="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" :disabled="loading" class="btn btn-primary">
              Add Script
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useScripts } from '@/composables/useScripts'
import type { ScriptConfig } from '@/types/api'

const {
  scripts,
  loading,
  error,
  fetchScripts,
  addScript,
  runScript,
  toggleScript,
  deleteScript
} = useScripts()

const showAddForm = ref(false)
const newScript = ref({
  path: '',
  interval: 3600,
  enabled: true,
  timeout: undefined as number | undefined
})

const editScript = (script: ScriptConfig) => {
  // TODO: Implement edit functionality
  console.log('Edit script:', script)
}

const handleAddScript = async () => {
  try {
    await addScript(newScript.value)
    showAddForm.value = false
    newScript.value = {
      path: '',
      interval: 3600,
      enabled: true,
      timeout: undefined
    }
  } catch (err) {
    console.error('Failed to add script:', err)
  }
}

onMounted(() => {
  fetchScripts()
})
</script>

<style scoped>
.scripts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.scripts-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.script-card {
  background: var(--color-background-soft);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.script-info h3 {
  margin: 0 0 0.5rem 0;
  color: var(--color-text);
}

.script-path {
  color: var(--color-text-muted);
  font-family: monospace;
  margin: 0 0 1rem 0;
}

.script-details {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.detail {
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.enabled {
  color: var(--color-success);
  font-weight: 600;
}

.disabled {
  color: var(--color-warning);
  font-weight: 600;
}

.script-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.no-scripts {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-muted);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--color-background);
  border-radius: var(--radius-lg);
  padding: 2rem;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal h3 {
  margin: 0 0 1.5rem 0;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-group input[type="checkbox"] {
  width: auto;
  margin-right: 0.5rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

@media (max-width: 768px) {
  .script-card {
    flex-direction: column;
    gap: 1rem;
  }

  .script-actions {
    width: 100%;
  }

  .modal {
    margin: 1rem;
    width: calc(100% - 2rem);
  }
}
</style>
