import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

export async function authGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
) {
  const { isAuthenticated, initializeAuth } = useAuth()

  // Initialize auth status if not already done
  await initializeAuth()

  // Allow access to login page regardless of auth status
  if (to.name === 'login') {
    if (isAuthenticated.value) {
      // Already authenticated, redirect to dashboard
      next({ name: 'dashboard' })
    } else {
      next()
    }
    return
  }

  // For all other routes, require authentication
  if (!isAuthenticated.value) {
    next({ name: 'login' })
  } else {
    next()
  }
}
