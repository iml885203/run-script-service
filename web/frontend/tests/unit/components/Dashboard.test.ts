import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import { ref } from 'vue'

// Mock composables with proper reactive refs
const mockSystemMetrics = vi.fn()
const mockScripts = vi.fn()

vi.mock('@/composables/useSystemMetrics', () => ({
  useSystemMetrics: () => mockSystemMetrics()
}))

vi.mock('@/composables/useScripts', () => ({
  useScripts: () => mockScripts()
}))

describe('Dashboard Component', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Dashboard</div>' } },
        { path: '/scripts', component: { template: '<div>Scripts</div>' } }
      ]
    })

    // Default mock implementations
    mockSystemMetrics.mockReturnValue({
      metrics: ref({ status: 'running', uptime: '1 hour', runningScripts: 2, totalScripts: 3 }),
      loading: ref(false),
      error: ref(null),
      startAutoRefresh: vi.fn(),
      cleanup: vi.fn()
    })

    mockScripts.mockReturnValue({
      scripts: ref([
        { name: 'test-script', path: '/test.sh', interval: 300, enabled: true }
      ]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: vi.fn(),
      runScript: vi.fn()
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should display system metrics correctly', () => {
    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.metric-value.status-running').text()).toBe('running')
    expect(wrapper.text()).toContain('1 hour')
    expect(wrapper.text()).toContain('2')
    expect(wrapper.text()).toContain('3')
  })

  it('should display scripts overview', () => {
    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.script-item').exists()).toBe(true)
    expect(wrapper.text()).toContain('test-script')
    expect(wrapper.text()).toContain('/test.sh')
    expect(wrapper.text()).toContain('Interval: 300s')
    expect(wrapper.find('.status.enabled').text()).toBe('Enabled')
  })

  it('should have run now button for each script', () => {
    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    const runButton = wrapper.find('.btn-primary')
    expect(runButton.exists()).toBe(true)
    expect(runButton.text()).toBe('Run Now')
  })

  it('should show loading state for metrics', () => {
    mockSystemMetrics.mockReturnValue({
      metrics: ref(null),
      loading: ref(true),
      error: ref(null),
      startAutoRefresh: vi.fn(),
      cleanup: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.loading').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading system metrics')
  })

  it('should show error state for metrics', () => {
    mockSystemMetrics.mockReturnValue({
      metrics: ref(null),
      loading: ref(false),
      error: ref('Failed to load metrics'),
      startAutoRefresh: vi.fn(),
      cleanup: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.error').exists()).toBe(true)
    expect(wrapper.text()).toContain('Failed to load metrics')
  })

  it('should show no scripts message when no scripts configured', () => {
    mockScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: vi.fn(),
      runScript: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.no-scripts').exists()).toBe(true)
    expect(wrapper.text()).toContain('No scripts configured')
  })

  it('should call runScript when run button is clicked', async () => {
    const mockRunScript = vi.fn()
    mockScripts.mockReturnValue({
      scripts: ref([
        { name: 'test-script', path: '/test.sh', interval: 300, enabled: true }
      ]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: vi.fn(),
      runScript: mockRunScript
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    const runButton = wrapper.find('.btn-primary')
    await runButton.trigger('click')

    expect(mockRunScript).toHaveBeenCalledWith('test-script')
  })

  it('should show scripts loading state', () => {
    mockScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(true),
      error: ref(null),
      fetchScripts: vi.fn(),
      runScript: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    const scriptsSection = wrapper.find('.scripts-overview')
    expect(scriptsSection.find('.loading').exists()).toBe(true)
    expect(scriptsSection.text()).toContain('Loading scripts')
  })

  it('should show scripts error state', () => {
    mockScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(false),
      error: ref('Failed to load scripts'),
      fetchScripts: vi.fn(),
      runScript: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    const scriptsSection = wrapper.find('.scripts-overview')
    expect(scriptsSection.find('.error').exists()).toBe(true)
    expect(scriptsSection.text()).toContain('Failed to load scripts')
  })
})
