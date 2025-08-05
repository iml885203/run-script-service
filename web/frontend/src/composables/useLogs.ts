import { ref, computed, type Ref } from 'vue'
import { ApiService } from '@/services/api'
import type { LogEntry } from '@/types/api'

export function useLogs() {
  const logs: Ref<LogEntry[]> = ref([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const selectedScript = ref<string | null>(null)
  const limit = ref(50)

  const filteredLogs = computed(() => {
    if (!selectedScript.value) return logs.value
    return logs.value.filter(log => log.script === selectedScript.value)
  })

  const errorLogs = computed(() =>
    filteredLogs.value.filter(log => log.level === 'error')
  )

  const warningLogs = computed(() =>
    filteredLogs.value.filter(log => log.level === 'warning')
  )

  const infoLogs = computed(() =>
    filteredLogs.value.filter(log => log.level === 'info')
  )

  const fetchLogs = async (scriptName?: string, logLimit?: number): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      const result = await ApiService.getLogs(
        scriptName || selectedScript.value || undefined,
        logLimit || limit.value
      )
      // Ensure we always have an array
      logs.value = Array.isArray(result) ? result : []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch logs'
      console.error('Failed to fetch logs:', err)
      logs.value = [] // Ensure we have an empty array on error
    } finally {
      loading.value = false
    }
  }

  const clearLogs = async (scriptName?: string): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      await ApiService.clearLogs(scriptName || selectedScript.value || undefined)
      await fetchLogs() // Refresh the logs
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to clear logs'
      throw err
    } finally {
      loading.value = false
    }
  }

  const setScriptFilter = (scriptName: string | null): void => {
    selectedScript.value = scriptName
    fetchLogs() // Auto-refresh when filter changes
  }

  const setLimit = (newLimit: number): void => {
    limit.value = newLimit
    fetchLogs() // Auto-refresh when limit changes
  }

  const refreshLogs = (): Promise<void> => {
    return fetchLogs()
  }

  return {
    logs: logs as Readonly<Ref<LogEntry[]>>,
    loading: loading as Readonly<Ref<boolean>>,
    error: error as Readonly<Ref<string | null>>,
    selectedScript: selectedScript as Readonly<Ref<string | null>>,
    limit: limit as Readonly<Ref<number>>,
    filteredLogs,
    errorLogs,
    warningLogs,
    infoLogs,
    fetchLogs,
    clearLogs,
    setScriptFilter,
    setLimit,
    refreshLogs
  }
}
