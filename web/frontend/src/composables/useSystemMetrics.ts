import { ref, type Ref } from 'vue'
import { ApiService } from '@/services/api'
import type { SystemMetrics } from '@/types/api'

export function useSystemMetrics() {
  const metrics: Ref<SystemMetrics | null> = ref(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  let intervalId: number | null = null

  const fetchMetrics = async (): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      metrics.value = await ApiService.getStatus()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch system metrics'
      console.error('Failed to fetch system metrics:', err)
    } finally {
      loading.value = false
    }
  }

  const startAutoRefresh = (intervalMs: number = 30000): void => {
    stopAutoRefresh() // Clear any existing interval

    // Fetch immediately
    fetchMetrics()

    // Set up periodic refresh
    intervalId = window.setInterval(() => {
      fetchMetrics()
    }, intervalMs)
  }

  const stopAutoRefresh = (): void => {
    if (intervalId !== null) {
      clearInterval(intervalId)
      intervalId = null
    }
  }

  const refreshMetrics = (): Promise<void> => {
    return fetchMetrics()
  }

  // Cleanup on unmount
  const cleanup = (): void => {
    stopAutoRefresh()
  }

  return {
    metrics: metrics as Readonly<Ref<SystemMetrics | null>>,
    loading: loading as Readonly<Ref<boolean>>,
    error: error as Readonly<Ref<string | null>>,
    fetchMetrics,
    startAutoRefresh,
    stopAutoRefresh,
    refreshMetrics,
    cleanup
  }
}
