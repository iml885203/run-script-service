export interface LoginRequest {
  secretKey: string
}

export interface AuthStatus {
  authenticated: boolean
  user?: string
}

class AuthService {
  private baseURL: string

  constructor() {
    this.baseURL = window.location.origin
  }

  async login(secretKey: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseURL}/api/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include', // Include cookies in request
        body: JSON.stringify({ secretKey }),
      })

      if (response.ok) {
        return true
      } else {
        const errorData = await response.json().catch(() => ({}))
        console.error('Login failed:', errorData)
        return false
      }
    } catch (error) {
      console.error('Login request failed:', error)
      throw error
    }
  }

  async logout(): Promise<void> {
    try {
      await fetch(`${this.baseURL}/api/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      })
    } catch (error) {
      console.error('Logout request failed:', error)
      // Don't throw - logout should work even if request fails
    }
  }

  async checkAuthStatus(): Promise<AuthStatus> {
    try {
      const response = await fetch(`${this.baseURL}/api/auth/status`, {
        method: 'GET',
        credentials: 'include',
      })

      if (response.ok) {
        const result = await response.json()
        // Extract data field from backend response
        return result.data || { authenticated: false }
      } else {
        return { authenticated: false }
      }
    } catch (error) {
      console.error('Auth status check failed:', error)
      return { authenticated: false }
    }
  }
}

export const authService = new AuthService()
