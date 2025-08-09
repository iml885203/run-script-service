import { describe, it, expect, beforeEach, vi } from 'vitest'
import { authGuard } from '@/router/guards'
import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router'

// Mock useAuth composable
const mockIsAuthenticated = vi.fn()
const mockInitializeAuth = vi.fn()

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => ({
    isAuthenticated: { value: mockIsAuthenticated() },
    initializeAuth: mockInitializeAuth
  })
}))

describe('authGuard', () => {
  let mockNext: NavigationGuardNext
  let mockTo: RouteLocationNormalized
  let mockFrom: RouteLocationNormalized

  beforeEach(() => {
    vi.clearAllMocks()
    mockNext = vi.fn()
    mockTo = {
      name: 'dashboard',
      path: '/',
      params: {},
      query: {},
      hash: '',
      matched: [],
      meta: {},
      fullPath: '/',
      redirectedFrom: undefined
    }
    mockFrom = {
      name: 'login',
      path: '/login',
      params: {},
      query: {},
      hash: '',
      matched: [],
      meta: {},
      fullPath: '/login',
      redirectedFrom: undefined
    }
    mockInitializeAuth.mockResolvedValue(undefined)
  })

  describe('when navigating to login page', () => {
    beforeEach(() => {
      mockTo.name = 'login'
      mockTo.path = '/login'
    })

    it('should allow access when not authenticated', async () => {
      mockIsAuthenticated.mockReturnValue(false)

      await authGuard(mockTo, mockFrom, mockNext)

      expect(mockInitializeAuth).toHaveBeenCalledOnce()
      expect(mockNext).toHaveBeenCalledWith()
    })

    it('should redirect to dashboard when already authenticated', async () => {
      mockIsAuthenticated.mockReturnValue(true)

      await authGuard(mockTo, mockFrom, mockNext)

      expect(mockInitializeAuth).toHaveBeenCalledOnce()
      expect(mockNext).toHaveBeenCalledWith({ name: 'dashboard' })
    })
  })

  describe('when navigating to protected routes', () => {
    beforeEach(() => {
      mockTo.name = 'dashboard'
      mockTo.path = '/'
    })

    it('should allow access when authenticated', async () => {
      mockIsAuthenticated.mockReturnValue(true)

      await authGuard(mockTo, mockFrom, mockNext)

      expect(mockInitializeAuth).toHaveBeenCalledOnce()
      expect(mockNext).toHaveBeenCalledWith()
    })

    it('should redirect to login when not authenticated', async () => {
      mockIsAuthenticated.mockReturnValue(false)

      await authGuard(mockTo, mockFrom, mockNext)

      expect(mockInitializeAuth).toHaveBeenCalledOnce()
      expect(mockNext).toHaveBeenCalledWith({ name: 'login' })
    })
  })

  describe('when navigating to scripts page', () => {
    beforeEach(() => {
      mockTo.name = 'Scripts'
      mockTo.path = '/scripts'
    })

    it('should allow access when authenticated', async () => {
      mockIsAuthenticated.mockReturnValue(true)

      await authGuard(mockTo, mockFrom, mockNext)

      expect(mockNext).toHaveBeenCalledWith()
    })

    it('should redirect to login when not authenticated', async () => {
      mockIsAuthenticated.mockReturnValue(false)

      await authGuard(mockTo, mockFrom, mockNext)

      expect(mockNext).toHaveBeenCalledWith({ name: 'login' })
    })
  })

  it('should always initialize auth before making decisions', async () => {
    mockIsAuthenticated.mockReturnValue(true)

    await authGuard(mockTo, mockFrom, mockNext)

    expect(mockInitializeAuth).toHaveBeenCalledOnce()
    expect(mockNext).toHaveBeenCalledOnce()
  })
})
