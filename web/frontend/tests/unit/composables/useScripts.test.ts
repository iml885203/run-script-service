import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useScripts } from '@/composables/useScripts'
import { ApiService } from '@/services/api'
import type { ScriptConfig } from '@/types/api'

// Mock ApiService
vi.mock('@/services/api', () => ({
  ApiService: {
    getScripts: vi.fn(),
    addScript: vi.fn(),
    updateScript: vi.fn(),
    deleteScript: vi.fn(),
    runScript: vi.fn()
  }
}))

const mockApiService = vi.mocked(ApiService)

describe('useScripts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should initialize with empty scripts array', () => {
    const { scripts, loading, error } = useScripts()

    expect(scripts.value).toEqual([])
    expect(loading.value).toBe(false)
    expect(error.value).toBe(null)
  })

  it('should fetch scripts successfully', async () => {
    const mockScripts: ScriptConfig[] = [
      {
        name: 'test-script-1',
        path: '/path/to/script1.sh',
        interval: 300,
        enabled: true
      },
      {
        name: 'test-script-2',
        path: '/path/to/script2.sh',
        interval: 600,
        enabled: false
      }
    ]

    mockApiService.getScripts.mockResolvedValueOnce(mockScripts)

    const { scripts, loading, error, fetchScripts } = useScripts()

    await fetchScripts()

    expect(loading.value).toBe(false)
    expect(error.value).toBe(null)
    expect(scripts.value).toEqual(mockScripts)
    expect(mockApiService.getScripts).toHaveBeenCalledOnce()
  })

  it('should handle fetch errors', async () => {
    const errorMessage = 'Network error'
    mockApiService.getScripts.mockRejectedValueOnce(new Error(errorMessage))

    const { scripts, loading, error, fetchScripts } = useScripts()

    await fetchScripts()

    expect(loading.value).toBe(false)
    expect(error.value).toBe(errorMessage)
    expect(scripts.value).toEqual([])
  })

  it('should compute enabled and disabled scripts correctly', async () => {
    const mockScripts: ScriptConfig[] = [
      {
        name: 'enabled-script',
        path: '/path/to/enabled.sh',
        interval: 300,
        enabled: true
      },
      {
        name: 'disabled-script',
        path: '/path/to/disabled.sh',
        interval: 600,
        enabled: false
      }
    ]

    mockApiService.getScripts.mockResolvedValueOnce(mockScripts)

    const { enabledScripts, disabledScripts, fetchScripts } = useScripts()

    await fetchScripts()

    expect(enabledScripts.value).toHaveLength(1)
    expect(enabledScripts.value[0].name).toBe('enabled-script')
    expect(disabledScripts.value).toHaveLength(1)
    expect(disabledScripts.value[0].name).toBe('disabled-script')
  })

  it('should add script and refresh list', async () => {
    const newScript = {
      name: 'new-script',
      path: '/path/to/new-script.sh',
      interval: 900,
      enabled: true
    }

    mockApiService.addScript.mockResolvedValueOnce(undefined)
    mockApiService.getScripts.mockResolvedValueOnce([])

    const { addScript } = useScripts()

    await addScript(newScript)

    expect(mockApiService.addScript).toHaveBeenCalledWith(newScript)
    expect(mockApiService.getScripts).toHaveBeenCalledOnce()
  })

  it('should toggle script enabled state', async () => {
    const mockScripts: ScriptConfig[] = [
      {
        name: 'test-script',
        path: '/path/to/script.sh',
        interval: 300,
        enabled: true
      }
    ]

    mockApiService.getScripts
      .mockResolvedValueOnce(mockScripts)
      .mockResolvedValueOnce([])
    mockApiService.updateScript.mockResolvedValueOnce(undefined)

    const { toggleScript, fetchScripts } = useScripts()

    await fetchScripts()
    await toggleScript('test-script')

    expect(mockApiService.updateScript).toHaveBeenCalledWith('test-script', { enabled: false })
  })

  it('should handle toggle script for non-existent script', async () => {
    mockApiService.getScripts.mockResolvedValueOnce([])

    const { toggleScript, fetchScripts } = useScripts()

    await fetchScripts()

    await expect(toggleScript('non-existent')).rejects.toThrow('Script non-existent not found')
  })
})
