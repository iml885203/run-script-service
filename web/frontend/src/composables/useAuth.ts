import { ref, computed } from 'vue'
import { authService, type AuthStatus } from '@/services/authService'

const authStatus = ref<AuthStatus>({ authenticated: false })
const isLoading = ref(false)

export function useAuth() {
  const isAuthenticated = computed(() => authStatus.value.authenticated)
  const user = computed(() => authStatus.value.user)

  const login = async (secretKey: string): Promise<boolean> => {
    isLoading.value = true
    try {
      const success = await authService.login(secretKey)
      if (success) {
        await checkAuthStatus()
      }
      return success
    } finally {
      isLoading.value = false
    }
  }

  const logout = async (): Promise<void> => {
    isLoading.value = true
    try {
      await authService.logout()
      authStatus.value = { authenticated: false }
    } finally {
      isLoading.value = false
    }
  }

  const checkAuthStatus = async (): Promise<void> => {
    try {
      const status = await authService.checkAuthStatus()
      authStatus.value = status
    } catch (error) {
      console.error('Failed to check auth status:', error)
      authStatus.value = { authenticated: false }
    }
  }

  // Initialize auth status on first use
  const initializeAuth = async () => {
    if (!authStatus.value.authenticated) {
      await checkAuthStatus()
    }
  }

  return {
    isAuthenticated,
    user,
    isLoading: computed(() => isLoading.value),
    login,
    logout,
    checkAuthStatus,
    initializeAuth,
  }
}
