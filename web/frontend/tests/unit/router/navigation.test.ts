import { describe, it, expect, beforeEach } from 'vitest'
import { createRouter, createWebHistory } from 'vue-router'
import { mount } from '@vue/test-utils'
import App from '@/App.vue'

// Create test router with actual routes
const createTestRouter = () => {
  return createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/', name: 'Dashboard', component: { template: '<div>Dashboard View</div>' } },
      { path: '/scripts', name: 'Scripts', component: { template: '<div>Scripts View</div>' } },
      { path: '/logs', name: 'Logs', component: { template: '<div>Logs View</div>' } },
      { path: '/settings', name: 'Settings', component: { template: '<div>Settings View</div>' } }
    ]
  })
}

describe('Router Navigation', () => {
  let router: any

  beforeEach(() => {
    router = createTestRouter()
  })

  it('should navigate to dashboard by default', async () => {
    await router.push('/')
    expect(router.currentRoute.value.name).toBe('Dashboard')
    expect(router.currentRoute.value.path).toBe('/')
  })

  it('should navigate to scripts page', async () => {
    await router.push('/scripts')
    expect(router.currentRoute.value.name).toBe('Scripts')
    expect(router.currentRoute.value.path).toBe('/scripts')
  })

  it('should navigate to logs page', async () => {
    await router.push('/logs')
    expect(router.currentRoute.value.name).toBe('Logs')
    expect(router.currentRoute.value.path).toBe('/logs')
  })

  it('should navigate to settings page', async () => {
    await router.push('/settings')
    expect(router.currentRoute.value.name).toBe('Settings')
    expect(router.currentRoute.value.path).toBe('/settings')
  })

  it('should handle invalid routes gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    try {
      await router.push('/invalid-route')
      // Router should handle this gracefully
    } catch (error) {
      // Error handling for invalid routes
    }

    consoleSpy.mockRestore()
  })

  it('should maintain route history', async () => {
    await router.push('/')
    await router.push('/scripts')
    await router.push('/logs')

    expect(router.currentRoute.value.path).toBe('/logs')

    await router.back()
    expect(router.currentRoute.value.path).toBe('/scripts')

    await router.back()
    expect(router.currentRoute.value.path).toBe('/')
  })

  it('should support programmatic navigation', async () => {
    await router.push({ name: 'Scripts' })
    expect(router.currentRoute.value.name).toBe('Scripts')

    await router.push({ name: 'Logs' })
    expect(router.currentRoute.value.name).toBe('Logs')

    await router.push({ name: 'Settings' })
    expect(router.currentRoute.value.name).toBe('Settings')
  })

  it('should handle route parameters correctly', async () => {
    // Add parameterized route for testing
    router.addRoute({
      path: '/scripts/:id',
      name: 'ScriptDetail',
      component: { template: '<div>Script Detail</div>' }
    })

    await router.push('/scripts/test-script')
    expect(router.currentRoute.value.params.id).toBe('test-script')
  })

  it('should handle query parameters', async () => {
    await router.push({ path: '/logs', query: { script: 'test1', limit: '50' } })

    expect(router.currentRoute.value.query.script).toBe('test1')
    expect(router.currentRoute.value.query.limit).toBe('50')
  })
})
