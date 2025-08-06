import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiService } from '@/services/api'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('ApiService - Real Backend Data', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should parse real backend scripts response correctly', async () => {
    // This is the actual response format from the backend
    const mockResponse = {
      success: true,
      data: [
        {
          enabled: true,
          interval: 300,
          max_log_lines: 100,
          name: "test1",
          path: "/home/logan/run-script-service-develop/test1.sh",
          running: true,
          timeout: 0
        },
        {
          enabled: true,
          interval: 600,
          max_log_lines: 100,
          name: "test2",
          path: "/home/logan/run-script-service-develop/test2.sh",
          running: true,
          timeout: 30
        }
      ]
    }

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => mockResponse
    })

    const scripts = await ApiService.getScripts()

    expect(scripts).toHaveLength(2)
    expect(scripts[0]).toEqual({
      enabled: true,
      interval: 300,
      max_log_lines: 100,
      name: "test1",
      path: "/home/logan/run-script-service-develop/test1.sh",
      running: true,
      timeout: 0
    })
    expect(scripts[1]).toEqual({
      enabled: true,
      interval: 600,
      max_log_lines: 100,
      name: "test2",
      path: "/home/logan/run-script-service-develop/test2.sh",
      running: true,
      timeout: 30
    })
  })

  it('should handle successful response with success=true', async () => {
    const mockResponse = {
      success: true,
      data: []
    }

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => mockResponse
    })

    const scripts = await ApiService.getScripts()
    expect(scripts).toEqual([])
  })

  it('should handle API response with success=false', async () => {
    const mockResponse = {
      success: false,
      error: 'Failed to fetch scripts'
    }

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => mockResponse
    })

    await expect(ApiService.getScripts()).rejects.toThrow('Failed to fetch scripts')
  })

  it('should handle network errors', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network error'))

    await expect(ApiService.getScripts()).rejects.toThrow('Network error')
  })

  it('should handle HTTP error responses', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error'
    })

    await expect(ApiService.getScripts()).rejects.toThrow('API request failed: 500 Internal Server Error')
  })
})
