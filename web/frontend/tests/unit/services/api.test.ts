import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiService } from '@/services/api'
import type { ScriptConfig, SystemMetrics } from '@/types/api'

// Mock global fetch
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('ApiService', () => {
  beforeEach(() => {
    mockFetch.mockClear()
  })

  describe('getScripts', () => {
    it('should fetch scripts correctly', async () => {
      const mockScripts: ScriptConfig[] = [
        {
          name: 'test-script',
          path: '/path/to/script.sh',
          interval: 300,
          enabled: true,
          timeout: 60
        }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockScripts
      })

      const result = await ApiService.getScripts()

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts', {
        headers: { 'Content-Type': 'application/json' }
      })
      expect(result).toEqual(mockScripts)
    })

    it('should handle API errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error'
      })

      await expect(ApiService.getScripts()).rejects.toThrow(
        'API request failed: 500 Internal Server Error'
      )
    })
  })

  describe('addScript', () => {
    it('should add script with correct data', async () => {
      const scriptData = {
        path: '/path/to/new-script.sh',
        interval: 600,
        enabled: true
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({})
      })

      await ApiService.addScript(scriptData)

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(scriptData)
      })
    })
  })

  describe('getStatus', () => {
    it('should fetch system metrics', async () => {
      const mockMetrics: SystemMetrics = {
        uptime: '2 hours 30 minutes',
        status: 'running',
        runningScripts: 3,
        totalScripts: 5
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetrics
      })

      const result = await ApiService.getStatus()

      expect(mockFetch).toHaveBeenCalledWith('/api/status', {
        headers: { 'Content-Type': 'application/json' }
      })
      expect(result).toEqual(mockMetrics)
    })
  })

  describe('runScript', () => {
    it('should run script with correct name', async () => {
      const scriptName = 'test-script'

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({})
      })

      await ApiService.runScript(scriptName)

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts/test-script/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      })
    })

    it('should encode script names with special characters', async () => {
      const scriptName = 'test script with spaces'

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({})
      })

      await ApiService.runScript(scriptName)

      expect(mockFetch).toHaveBeenCalledWith('/api/scripts/test%20script%20with%20spaces/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      })
    })
  })
})
