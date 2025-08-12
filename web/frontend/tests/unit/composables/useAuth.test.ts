import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useAuth } from '@/composables/useAuth'
import { authService } from '@/services/authService'

// Mock authService
vi.mock('@/services/authService', () => ({
  authService: {
    login: vi.fn(),
    logout: vi.fn(),
    checkAuthStatus: vi.fn()
  }
}))

describe('useAuth', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    // Reset the shared authentication state by calling logout
    const { logout } = await import('@/composables/useAuth').then(m => m.useAuth())
    const mockAuthService = vi.mocked(authService)
    mockAuthService.logout.mockResolvedValueOnce()
    await logout()
  })

  it('should initialize with unauthenticated state', () => {
    const { isAuthenticated, user, isLoading } = useAuth()

    expect(isAuthenticated.value).toBe(false)
    expect(user.value).toBeUndefined()
    expect(isLoading.value).toBe(false)
  })

  describe('login', () => {
    it('should call authService.login and update state on success', async () => {
      const mockAuthService = vi.mocked(authService)
      mockAuthService.login.mockResolvedValueOnce(true)
      mockAuthService.checkAuthStatus.mockResolvedValueOnce({
        authenticated: true,
        user: 'test-user'
      })

      const { login, isAuthenticated, isLoading } = useAuth()

      const promise = login('test-secret')
      expect(isLoading.value).toBe(true)

      const result = await promise
      expect(result).toBe(true)
      expect(isLoading.value).toBe(false)
      expect(isAuthenticated.value).toBe(true)
      expect(mockAuthService.login).toHaveBeenCalledWith('test-secret')
    })

    it('should call authService.login and not update state on failure', async () => {
      const mockAuthService = vi.mocked(authService)
      mockAuthService.login.mockResolvedValueOnce(false)
      // Don't call checkAuthStatus on failed login, so state remains false

      const { login, isAuthenticated, isLoading } = useAuth()

      const promise = login('wrong-secret')
      expect(isLoading.value).toBe(true)

      const result = await promise
      expect(result).toBe(false)
      expect(isLoading.value).toBe(false)
      expect(isAuthenticated.value).toBe(false)
    })
  })

  describe('logout', () => {
    it('should call authService.logout and reset auth state', async () => {
      const mockAuthService = vi.mocked(authService)
      vi.clearAllMocks() // Clear mocks called in beforeEach
      mockAuthService.logout.mockResolvedValueOnce()

      const { logout, isAuthenticated, isLoading } = useAuth()

      const promise = logout()
      expect(isLoading.value).toBe(true)

      await promise
      expect(isLoading.value).toBe(false)
      expect(isAuthenticated.value).toBe(false)
      expect(mockAuthService.logout).toHaveBeenCalledOnce()
    })
  })

  describe('checkAuthStatus', () => {
    it('should update auth state based on service response', async () => {
      const mockAuthService = vi.mocked(authService)
      mockAuthService.checkAuthStatus.mockResolvedValueOnce({
        authenticated: true,
        user: 'test-user'
      })

      const { checkAuthStatus, isAuthenticated, user } = useAuth()

      await checkAuthStatus()

      expect(isAuthenticated.value).toBe(true)
      expect(user.value).toBe('test-user')
    })

    it('should handle errors gracefully', async () => {
      const mockAuthService = vi.mocked(authService)
      mockAuthService.checkAuthStatus.mockRejectedValueOnce(new Error('Network error'))

      const { checkAuthStatus, isAuthenticated } = useAuth()

      await checkAuthStatus()

      expect(isAuthenticated.value).toBe(false)
    })
  })

  describe('initializeAuth', () => {
    it('should check auth status if not already authenticated', async () => {
      const mockAuthService = vi.mocked(authService)
      mockAuthService.checkAuthStatus.mockResolvedValueOnce({
        authenticated: false
      })

      const { initializeAuth } = useAuth()

      await initializeAuth()

      expect(mockAuthService.checkAuthStatus).toHaveBeenCalledOnce()
    })

    it('should not check auth status if already authenticated', async () => {
      const mockAuthService = vi.mocked(authService)
      mockAuthService.checkAuthStatus.mockResolvedValueOnce({
        authenticated: true
      })

      const { initializeAuth, checkAuthStatus } = useAuth()

      // First call to authenticate
      await checkAuthStatus()
      vi.clearAllMocks()

      // Second call should not trigger another check
      await initializeAuth()

      expect(mockAuthService.checkAuthStatus).not.toHaveBeenCalled()
    })
  })
})
