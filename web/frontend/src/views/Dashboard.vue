<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h2>System Dashboard</h2>
      <!-- Real-time connection status indicator -->
      <div class="connection-status" :class="{ 
        'connected': wsIsConnected, 
        'disconnected': !wsIsConnected 
      }">
        <span class="status-dot"></span>
        {{ wsIsConnected ? 'Live' : 'Offline' }}
      </div>
    </div>

    <!-- WebSocket error display -->
    <div v-if="wsError" class="websocket-error">
      Real-time updates unavailable: {{ wsError }}
    </div>

    <div v-if="metricsLoading" class="loading">
      Loading system metrics...
    </div>

    <div v-else-if="metricsError" class="error">
      {{ metricsError }}
    </div>

    <div v-else-if="metrics" class="metrics-grid">
      <div class="metric-card" data-testid="system-status-card">
        <h3>System Status</h3>
        <div class="metric-value" :class="{ 'status-running': metrics.status === 'running' }" data-testid="status-value">
          {{ metrics.status }}
        </div>
      </div>

      <div class="metric-card" data-testid="uptime-card">
        <h3>Uptime</h3>
        <div class="metric-value" data-testid="uptime-value">{{ metrics.uptime }}</div>
      </div>

      <div class="metric-card" data-testid="running-scripts-card">
        <h3>Running Scripts</h3>
        <div class="metric-value" data-testid="running-scripts-value">{{ metrics.runningScripts }}</div>
      </div>

      <div class="metric-card" data-testid="total-scripts-card">
        <h3>Total Scripts</h3>
        <div class="metric-value" data-testid="total-scripts-value">{{ metrics.totalScripts }}</div>
      </div>

      <!-- System resource metrics -->
      <div v-if="metrics.cpu_percent !== undefined" class="metric-card" data-testid="cpu-card">
        <h3>CPU Usage</h3>
        <div class="metric-value cpu-metric">{{ metrics.cpu_percent }}%</div>
      </div>

      <div v-if="metrics.memory_percent !== undefined" class="metric-card" data-testid="memory-card">
        <h3>Memory Usage</h3>
        <div class="metric-value memory-metric">{{ metrics.memory_percent }}%</div>
      </div>

      <div v-if="metrics.disk_percent !== undefined" class="metric-card" data-testid="disk-card">
        <h3>Disk Usage</h3>
        <div class="metric-value disk-metric">{{ metrics.disk_percent }}%</div>
      </div>
    </div>

    <div class="scripts-overview">
      <div class="scripts-overview-header">
        <h3>Scripts Overview</h3>
        <router-link to="/scripts" class="scripts-link" data-testid="manage-scripts-link">Add scripts</router-link>
      </div>

      <div v-if="scriptsLoading" class="loading">
        Loading scripts...
      </div>

      <div v-else-if="scriptsError" class="error">
        {{ scriptsError }}
      </div>

      <div v-else-if="scripts.length > 0" class="scripts-list">
        <div v-for="script in scripts" :key="script.name" class="script-item">
          <div class="script-info">
            <h4>{{ script.name }}</h4>
            <p>{{ script.path }}</p>
            <div class="script-meta">
              <span class="interval">Interval: {{ script.interval }}s</span>
              <span class="status" :class="{ 'enabled': script.enabled, 'disabled': !script.enabled }">
                {{ script.enabled ? 'Enabled' : 'Disabled' }}
              </span>
              <!-- Real-time script status -->
              <span v-if="script.status" class="script-status" :class="script.status">
                {{ script.status }}
              </span>
            </div>
          </div>
          <div class="script-actions">
            <button @click="runScript(script.name)" :disabled="scriptsLoading" class="btn-primary">
              Run Now
            </button>
          </div>
        </div>
      </div>

      <div v-else class="no-scripts">
        No scripts configured. <router-link to="/scripts" data-testid="add-scripts-link">Add scripts</router-link> to get started.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useSystemMetrics } from '@/composables/useSystemMetrics'
import { useScripts } from '@/composables/useScripts'
import { useWebSocket } from '@/composables/useWebSocket'

const {
  metrics,
  loading: metricsLoading,
  error: metricsError,
  startAutoRefresh,
  cleanup: cleanupMetrics
} = useSystemMetrics()

const {
  scripts,
  loading: scriptsLoading,
  error: scriptsError,
  fetchScripts,
  runScript
} = useScripts()

// WebSocket for real-time updates
const {
  isConnected: wsIsConnected,
  error: wsError,
  connect: wsConnect,
  disconnect: wsDisconnect,
  onMessage
} = useWebSocket()

onMounted(async () => {
  // Start auto-refresh for metrics (every 30 seconds)
  startAutoRefresh(30000)

  // Fetch scripts initially
  await fetchScripts()

  // Connect WebSocket for real-time updates
  wsConnect()

  // Register WebSocket message handlers
  onMessage('system_metrics', (message: any) => {
    // Update system metrics in real-time
    if (metrics.value && message.data) {
      Object.assign(metrics.value, message.data)
    }
  })

  onMessage('script_status', (message: any) => {
    // Update script status in real-time
    if (message.data && scripts.value) {
      const script = scripts.value.find(s => s.name === message.data.script_name)
      if (script) {
        script.status = message.data.status
      }
    }
  })
})

onUnmounted(() => {
  cleanupMetrics()
  wsDisconnect()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

h2 {
  color: var(--color-text);
  margin: 0;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  border: 1px solid;
}

.connection-status.connected {
  color: var(--color-success);
  background: var(--color-success-soft);
  border-color: var(--color-success);
}

.connection-status.disconnected {
  color: var(--color-warning);
  background: var(--color-warning-soft);
  border-color: var(--color-warning);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.connection-status.connected .status-dot {
  background: var(--color-success);
  box-shadow: 0 0 6px var(--color-success);
}

.connection-status.disconnected .status-dot {
  background: var(--color-warning);
}

.websocket-error {
  background: var(--color-danger-soft);
  color: var(--color-danger);
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  border: 1px solid var(--color-danger);
}

.loading, .error {
  padding: 1rem;
  text-align: center;
  color: var(--color-text-muted);
}

.error {
  color: var(--color-danger);
  background: var(--color-danger-soft);
  border-radius: 0.5rem;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.metric-card {
  background: var(--color-background-soft);
  padding: 1.5rem;
  border-radius: 0.75rem;
  border: 1px solid var(--color-border);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.metric-card h3 {
  margin: 0 0 1rem 0;
  color: var(--color-text-muted);
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
}

.metric-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  text-transform: capitalize;
}

.metric-value.status-running {
  color: var(--color-success);
}

.scripts-overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.scripts-overview-header h3 {
  color: var(--color-text);
  margin: 0;
}

.scripts-link {
  color: var(--color-brand);
  text-decoration: none;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-brand);
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.scripts-link:hover {
  background: var(--color-brand);
  color: white;
}

.scripts-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.script-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--color-background-soft);
  padding: 1.5rem;
  border-radius: 0.75rem;
  border: 1px solid var(--color-border);
}

.script-info h4 {
  margin: 0 0 0.5rem 0;
  color: var(--color-text);
}

.script-info p {
  margin: 0 0 0.75rem 0;
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.script-meta {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.interval {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.status {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  text-transform: uppercase;
}

.status.enabled {
  color: var(--color-success);
  background: var(--color-success-soft);
}

.status.disabled {
  color: var(--color-warning);
  background: var(--color-warning-soft);
}

.script-status {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  text-transform: uppercase;
}

.script-status.running {
  color: var(--color-brand);
  background: var(--color-brand-soft);
}

.script-status.completed {
  color: var(--color-success);
  background: var(--color-success-soft);
}

.script-status.failed {
  color: var(--color-danger);
  background: var(--color-danger-soft);
}

.script-status.idle {
  color: var(--color-text-muted);
  background: var(--color-background-mute);
}

.btn-primary {
  background: var(--color-brand);
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-brand-dark);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.no-scripts {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-muted);
}

.no-scripts a {
  color: var(--color-brand);
  text-decoration: none;
  font-weight: 500;
}

.no-scripts a:hover {
  text-decoration: underline;
}

@media (max-width: 768px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }

  .script-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .script-actions {
    width: 100%;
  }

  .btn-primary {
    width: 100%;
  }
}
</style>
