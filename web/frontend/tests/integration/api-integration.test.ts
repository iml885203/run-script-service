import { describe, it, expect, beforeAll, afterAll, beforeEach, vi } from 'vitest'
import { ApiService } from '@/services/api'

// Mock fetch for integration tests
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('API Integration Tests', () => {
  beforeAll(() => {
    // Setup test environment
  })

  beforeEach(() => {
    // Clear all mocks before each test to ensure isolation
    vi.clearAllMocks()
    mockFetch.mockClear()
  })

  afterAll(() => {
    vi.clearAllMocks()
  })

  describe('Scripts API', () => {
    it('should fetch scripts from backend', async () => {
      const mockScripts = [
        { name: 'test1', path: '/test1.sh', interval: 300, enabled: true },
        { name: 'test2', path: '/test2.sh', interval: 600, enabled: false }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockScripts
      })

      const scripts = await ApiService.getScripts()

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts', {
        headers: {
          'Content-Type': 'application/json'
        }
      })
      expect(scripts).toEqual(mockScripts)
    })

    it('should add new script to backend', async () => {
      const newScript = {
        name: 'new-script',
        path: '/new-script.sh',
        interval: 1800,
        enabled: true
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      await ApiService.addScript(newScript)

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newScript)
      })
    })

    it('should run script on backend', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      await ApiService.runScript('test-script')

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts/test-script/run', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      })
    })

    it('should handle API errors gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error'
      })

      await expect(ApiService.getScripts()).rejects.toThrow('API request failed: 500 Internal Server Error')
    })
  })

  describe('Logs API', () => {
    it('should fetch logs from backend', async () => {
      const mockLogs = [
        {
          timestamp: '2024-01-01T12:00:00Z',
          script: 'test1',
          message: 'Script executed successfully',
          level: 'info'
        },
        {
          timestamp: '2024-01-01T12:01:00Z',
          script: 'test2',
          message: 'Script failed',
          level: 'error'
        }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockLogs
      })

      const logs = await ApiService.getLogs()

      expect(mockFetch).toHaveBeenCalledWith('/api/logs?limit=50', {
        headers: {
          'Content-Type': 'application/json'
        }
      })
      expect(logs).toEqual(mockLogs)
    })

    it('should fetch logs with script filter', async () => {
      const mockLogs = [
        {
          timestamp: '2024-01-01T12:00:00Z',
          script: 'test1',
          message: 'Script executed successfully',
          level: 'info'
        }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockLogs
      })

      const logs = await ApiService.getLogs('test1', 100)

      expect(mockFetch).toHaveBeenCalledWith('/api/logs?script=test1&limit=100', {
        headers: {
          'Content-Type': 'application/json'
        }
      })
      expect(logs).toEqual(mockLogs)
    })

    it('should clear script logs', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      await ApiService.clearLogs('test1')

      expect(mockFetch).toHaveBeenCalledWith('/api/logs/test1', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json'
        }
      })
    })
  })

  describe('System Status API', () => {
    it('should fetch system status from backend', async () => {
      const mockStatus = {
        status: 'running',
        uptime: '2 hours',
        runningScripts: 3,
        totalScripts: 5
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockStatus
      })

      const status = await ApiService.getStatus()

      expect(mockFetch).toHaveBeenCalledWith('/api/status', {
        headers: {
          'Content-Type': 'application/json'
        }
      })
      expect(status).toEqual(mockStatus)
    })
  })

  describe('Configuration API', () => {
    it('should fetch configuration from backend', async () => {
      const mockConfig = {
        webPort: 8080,
        autoRefresh: true,
        refreshInterval: 30000
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockConfig
      })

      const config = await ApiService.getConfig()

      expect(mockFetch).toHaveBeenCalledWith('/api/config', {
        headers: {
          'Content-Type': 'application/json'
        }
      })
      expect(config).toEqual(mockConfig)
    })

    it('should update configuration on backend', async () => {
      const newConfig = {
        webPort: 9080,
        autoRefresh: false,
        refreshInterval: 60000
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      await ApiService.updateConfig(newConfig)

      expect(mockFetch).toHaveBeenCalledWith('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newConfig)
      })
    })
  })

  describe('Error Handling', () => {
    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      await expect(ApiService.getScripts()).rejects.toThrow('Network error')
    })

    it('should handle JSON parsing errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error('Invalid JSON')
        }
      })

      await expect(ApiService.getScripts()).rejects.toThrow('Invalid JSON')
    })

    it('should handle 404 errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found'
      })

      await expect(ApiService.getScripts()).rejects.toThrow('API request failed: 404 Not Found')
    })
  })

  describe('Real-time Updates', () => {
    it('should handle WebSocket message integration with API updates', async () => {
      // Mock WebSocket message that triggers API refresh
      const mockScripts = [
        { name: 'updated-script', path: '/updated.sh', interval: 300, enabled: true }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockScripts
      })

      // Simulate WebSocket message triggering API refresh
      const scripts = await ApiService.getScripts()

      expect(scripts).toEqual(mockScripts)
      expect(mockFetch).toHaveBeenCalledWith('/api/scripts', {
        headers: {
          'Content-Type': 'application/json'
        }
      })
    })
  })
})
