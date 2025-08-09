import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/views/Login.vue'

// Mock useAuth composable
const mockLogin = vi.fn()
const mockIsAuthenticated = vi.fn()

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => ({
    login: mockLogin,
    isAuthenticated: { value: mockIsAuthenticated() }
  })
}))

// Mock router
const mockPush = vi.fn()
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: mockPush
    })
  }
})

describe('Login.vue', () => {
  let wrapper: any

  beforeEach(() => {
    vi.clearAllMocks()
    mockIsAuthenticated.mockReturnValue(false)

    const router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', name: 'dashboard', component: { template: '<div>Dashboard</div>' } },
        { path: '/login', name: 'login', component: Login }
      ]
    })

    wrapper = mount(Login, {
      global: {
        plugins: [router]
      }
    })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  it('should render login form', () => {
    expect(wrapper.find('h1').text()).toContain('Authentication Required')
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('should disable submit button when secret key is empty', async () => {
    const submitButton = wrapper.find('button[type="submit"]')
    expect(submitButton.attributes('disabled')).toBeDefined()
  })

  it('should enable submit button when secret key is entered', async () => {
    const input = wrapper.find('input[type="password"]')
    await input.setValue('test-secret-key')

    const submitButton = wrapper.find('button[type="submit"]')
    expect(submitButton.attributes('disabled')).toBeUndefined()
  })

  it('should call login on form submission with valid input', async () => {
    mockLogin.mockResolvedValueOnce(true)

    const input = wrapper.find('input[type="password"]')
    await input.setValue('test-secret-key')

    const form = wrapper.find('form')
    await form.trigger('submit')

    expect(mockLogin).toHaveBeenCalledWith('test-secret-key')
  })

  it('should redirect to dashboard on successful login', async () => {
    mockLogin.mockResolvedValueOnce(true)

    const input = wrapper.find('input[type="password"]')
    await input.setValue('test-secret-key')

    const form = wrapper.find('form')
    await form.trigger('submit')

    // Wait for async operations
    await wrapper.vm.$nextTick()

    expect(mockPush).toHaveBeenCalledWith('/')
  })

  it('should show error message on failed login', async () => {
    mockLogin.mockResolvedValueOnce(false)

    const input = wrapper.find('input[type="password"]')
    await input.setValue('wrong-secret-key')

    const form = wrapper.find('form')
    await form.trigger('submit')

    // Wait for async operations
    await wrapper.vm.$nextTick()

    expect(wrapper.find('.error-message').text()).toContain('Invalid secret key')
  })

  it('should show loading state during authentication', async () => {
    let resolveLogin: (value: boolean) => void
    const loginPromise = new Promise<boolean>((resolve) => {
      resolveLogin = resolve
    })
    mockLogin.mockReturnValueOnce(loginPromise)

    const input = wrapper.find('input[type="password"]')
    await input.setValue('test-secret-key')

    const form = wrapper.find('form')
    await form.trigger('submit')
    await wrapper.vm.$nextTick() // Let the submit handler start

    // Check loading state
    expect(wrapper.find('button[type="submit"]').text()).toContain('Authenticating...')
    expect(wrapper.find('input[type="password"]').attributes('disabled')).toBeDefined()

    // Resolve login
    resolveLogin!(true)
    await loginPromise // Wait for the promise to resolve
    await new Promise(resolve => setTimeout(resolve, 0)) // Wait for microtasks
    await wrapper.vm.$nextTick()

    // Check loading state is cleared
    expect(wrapper.find('button[type="submit"]').text()).toContain('Sign In')
  })

  it('should clear secret key from memory after successful login', async () => {
    mockLogin.mockResolvedValueOnce(true)

    const input = wrapper.find('input[type="password"]')
    await input.setValue('test-secret-key')

    const form = wrapper.find('form')
    await form.trigger('submit')

    // Wait for async operations
    await wrapper.vm.$nextTick()

    expect(input.element.value).toBe('')
  })

  it('should handle login errors gracefully', async () => {
    mockLogin.mockRejectedValueOnce(new Error('Network error'))

    const input = wrapper.find('input[type="password"]')
    await input.setValue('test-secret-key')

    const form = wrapper.find('form')
    await form.trigger('submit')

    // Wait for async operations
    await wrapper.vm.$nextTick()

    expect(wrapper.find('.error-message').text()).toContain('Authentication failed')
  })

  it('should validate empty input', async () => {
    const form = wrapper.find('form')
    await form.trigger('submit')

    expect(wrapper.find('.error-message').text()).toContain('Please enter a secret key')
    expect(mockLogin).not.toHaveBeenCalled()
  })
})
