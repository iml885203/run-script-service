import { describe, it, expect, beforeEach, vi } from 'vitest'
import { authService } from '@/services/authService'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('authService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Setup window.location.origin mock
    Object.defineProperty(window, 'location', {
      value: {
        origin: 'http://localhost:3000'
      },
      writable: true
    })
  })

  describe('login', () => {
    it('should return true on successful login', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      const result = await authService.login('test-secret-key')

      expect(result).toBe(true)
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:3000/api/auth/login',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify({ secretKey: 'test-secret-key' }),
        }
      )
    })

    it('should return false on failed login', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: async () => ({ success: false, error: 'Invalid secret key' })
      })

      const result = await authService.login('wrong-key')

      expect(result).toBe(false)
    })

    it('should throw error on network failure', async () => {
      const networkError = new Error('Network error')
      mockFetch.mockRejectedValueOnce(networkError)

      await expect(authService.login('test-key')).rejects.toThrow('Network error')
    })
  })

  describe('logout', () => {
    it('should call logout endpoint', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await authService.logout()

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:3000/api/auth/logout',
        {
          method: 'POST',
          credentials: 'include',
        }
      )
    })

    it('should not throw on network failure', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      await expect(authService.logout()).resolves.toBeUndefined()
    })
  })

  describe('checkAuthStatus', () => {
    it('should return authentication status on success', async () => {
      const mockStatus = { authenticated: true, user: 'test-user' }
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockStatus
      })

      const result = await authService.checkAuthStatus()

      expect(result).toEqual(mockStatus)
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:3000/api/auth/status',
        {
          method: 'GET',
          credentials: 'include',
        }
      )
    })

    it('should return unauthenticated on failed request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false
      })

      const result = await authService.checkAuthStatus()

      expect(result).toEqual({ authenticated: false })
    })

    it('should return unauthenticated on network error', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      const result = await authService.checkAuthStatus()

      expect(result).toEqual({ authenticated: false })
    })
  })
})
