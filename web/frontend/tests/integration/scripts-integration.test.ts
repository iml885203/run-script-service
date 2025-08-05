import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import Scripts from '@/views/Scripts.vue'

// Mock fetch to simulate real backend API
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('Scripts Integration Test', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Home</div>' } },
        { path: '/scripts', component: Scripts }
      ]
    })
    vi.clearAllMocks()
  })

  it('should load and display scripts from real API response', async () => {
    // Mock the real backend API response
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

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // Wait for the API call to complete
    await new Promise(resolve => setTimeout(resolve, 100))

    // Check if scripts are displayed
    const scriptCards = wrapper.findAll('[data-testid="script-card"]')
    console.log('Found script cards:', scriptCards.length)
    console.log('Component HTML:', wrapper.html())

    expect(scriptCards).toHaveLength(2)

    // Check script details
    if (scriptCards.length > 0) {
      const firstCard = scriptCards[0]
      expect(firstCard.find('[data-testid="script-name"]').text()).toBe('test1')
      expect(firstCard.find('[data-testid="script-path"]').text()).toBe('/home/logan/run-script-service-develop/test1.sh')
      expect(firstCard.find('[data-testid="script-interval"]').text()).toBe('Interval: 300s')
    }
  })

  it('should show loading state initially', () => {
    // Don't resolve the fetch promise immediately
    mockFetch.mockImplementation(() => new Promise(() => {}))

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // Should show loading state
    expect(wrapper.find('.loading').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading scripts...')
  })

  it('should show error state on API failure', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network error'))

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // Wait for the error to propagate
    await new Promise(resolve => setTimeout(resolve, 100))

    expect(wrapper.find('.error').exists()).toBe(true)
    expect(wrapper.text()).toContain('Network error')
  })
})
