import { ref, computed, type Ref } from 'vue'
import { ApiService } from '@/services/api'
import type { ScriptConfig } from '@/types/api'

export function useScripts() {
  const scripts: Ref<ScriptConfig[]> = ref([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const enabledScripts = computed(() =>
    scripts.value.filter(script => script.enabled)
  )

  const disabledScripts = computed(() =>
    scripts.value.filter(script => !script.enabled)
  )

  const fetchScripts = async (): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      scripts.value = await ApiService.getScripts()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch scripts'
      console.error('Failed to fetch scripts:', err)
    } finally {
      loading.value = false
    }
  }

  const addScript = async (scriptData: Omit<ScriptConfig, 'name'>): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      await ApiService.addScript(scriptData)
      await fetchScripts() // Refresh the list
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to add script'
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateScript = async (name: string, updates: Partial<ScriptConfig>): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      await ApiService.updateScript(name, updates)
      await fetchScripts() // Refresh the list
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update script'
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteScript = async (name: string): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      await ApiService.deleteScript(name)
      await fetchScripts() // Refresh the list
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete script'
      throw err
    } finally {
      loading.value = false
    }
  }

  const runScript = async (name: string): Promise<void> => {
    loading.value = true
    error.value = null

    try {
      await ApiService.runScript(name)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to run script'
      throw err
    } finally {
      loading.value = false
    }
  }

  const toggleScript = async (name: string): Promise<void> => {
    const script = scripts.value.find(s => s.name === name)
    if (!script) {
      throw new Error(`Script ${name} not found`)
    }

    await updateScript(name, { enabled: !script.enabled })
  }

  return {
    scripts: scripts as Readonly<Ref<ScriptConfig[]>>,
    loading: loading as Readonly<Ref<boolean>>,
    error: error as Readonly<Ref<string | null>>,
    enabledScripts,
    disabledScripts,
    fetchScripts,
    addScript,
    updateScript,
    deleteScript,
    runScript,
    toggleScript
  }
}
