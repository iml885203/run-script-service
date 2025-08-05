<template>
  <div class="logs">
    <h2>System Logs</h2>

    <div class="logs-controls">
      <div class="filters">
        <label for="script-filter">Filter by script:</label>
        <select id="script-filter" v-model="selectedScript" @change="setScriptFilter(selectedScript)">
          <option value="">All scripts</option>
          <option v-for="script in availableScripts" :key="script" :value="script">
            {{ script }}
          </option>
        </select>

        <label for="limit">Limit:</label>
        <select id="limit" v-model="currentLimit" @change="setLimit(currentLimit)">
          <option :value="25">25 entries</option>
          <option :value="50">50 entries</option>
          <option :value="100">100 entries</option>
          <option :value="200">200 entries</option>
        </select>
      </div>

      <div class="actions">
        <button @click="refreshLogs" :disabled="loading" class="btn btn-secondary">
          Refresh
        </button>
        <button @click="clearLogs()" :disabled="loading" class="btn btn-danger">
          Clear Logs
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      Loading logs...
    </div>

    <div v-else-if="error" class="error">
      {{ error }}
    </div>

    <div v-else-if="filteredLogs.length === 0" class="no-logs">
      No logs found.
    </div>

    <div v-else class="logs-container">
      <div class="logs-summary">
        <span class="summary-item">
          <strong>Total:</strong> {{ filteredLogs.length }}
        </span>
        <span class="summary-item error-count" v-if="errorLogs.length > 0">
          <strong>Errors:</strong> {{ errorLogs.length }}
        </span>
        <span class="summary-item warning-count" v-if="warningLogs.length > 0">
          <strong>Warnings:</strong> {{ warningLogs.length }}
        </span>
        <span class="summary-item info-count">
          <strong>Info:</strong> {{ infoLogs.length }}
        </span>
      </div>

      <div class="logs-list">
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          :class="['log-entry', `log-${log.level}`]"
        >
          <div class="log-header">
            <span class="log-timestamp">{{ formatTimestamp(log.timestamp) }}</span>
            <span :class="['log-level', `level-${log.level}`]">{{ log.level.toUpperCase() }}</span>
            <span v-if="log.script" class="log-script">{{ log.script }}</span>
          </div>
          <div class="log-message">{{ log.message }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useLogs } from '@/composables/useLogs'
import { useScripts } from '@/composables/useScripts'

const {
  logs,
  loading,
  error,
  selectedScript,
  limit,
  filteredLogs,
  errorLogs,
  warningLogs,
  infoLogs,
  fetchLogs,
  clearLogs,
  setScriptFilter,
  setLimit,
  refreshLogs
} = useLogs()

const { scripts, fetchScripts } = useScripts()

const currentLimit = ref(50)

const availableScripts = computed(() => {
  const scriptNames = new Set<string>()
  logs.value.forEach(log => {
    if (log.script) {
      scriptNames.add(log.script)
    }
  })
  return Array.from(scriptNames).sort()
})

const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp)
    return date.toLocaleString()
  } catch {
    return timestamp
  }
}

onMounted(async () => {
  await Promise.all([
    fetchLogs(),
    fetchScripts()
  ])
})
</script>

<style scoped>
.logs-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding: 1rem;
  background: var(--color-background-soft);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}

.filters {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.filters label {
  font-size: 0.875rem;
  font-weight: 500;
  margin: 0;
}

.filters select {
  width: auto;
  min-width: 120px;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.logs-summary {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1rem;
  padding: 1rem;
  background: var(--color-background-soft);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}

.summary-item {
  font-size: 0.875rem;
}

.error-count {
  color: var(--color-danger);
}

.warning-count {
  color: var(--color-warning);
}

.info-count {
  color: var(--color-text-muted);
}

.logs-list {
  max-height: 600px;
  overflow-y: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.log-entry {
  padding: 1rem;
  border-bottom: 1px solid var(--color-border);
  font-family: monospace;
  font-size: 0.875rem;
}

.log-entry:last-child {
  border-bottom: none;
}

.log-entry.log-error {
  background-color: var(--color-danger-soft);
  border-left: 4px solid var(--color-danger);
}

.log-entry.log-warning {
  background-color: var(--color-warning-soft);
  border-left: 4px solid var(--color-warning);
}

.log-entry.log-info {
  background-color: var(--color-background);
  border-left: 4px solid var(--color-text-muted);
}

.log-header {
  display: flex;
  gap: 1rem;
  align-items: center;
  margin-bottom: 0.5rem;
}

.log-timestamp {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.log-level {
  font-weight: 600;
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
}

.level-error {
  background: var(--color-danger);
  color: white;
}

.level-warning {
  background: var(--color-warning);
  color: white;
}

.level-info {
  background: var(--color-text-muted);
  color: white;
}

.log-script {
  background: var(--color-brand-soft);
  color: var(--color-brand);
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.log-message {
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text);
  line-height: 1.4;
}

.no-logs {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-muted);
}

@media (max-width: 768px) {
  .logs-controls {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }

  .filters {
    flex-direction: column;
    gap: 0.5rem;
    align-items: stretch;
  }

  .logs-summary {
    flex-direction: column;
    gap: 0.5rem;
  }

  .log-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
}
</style>
