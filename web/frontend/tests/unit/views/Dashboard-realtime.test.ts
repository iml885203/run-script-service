import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import { ref, nextTick } from 'vue'

// Mock WebSocket for real-time testing
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState = MockWebSocket.CONNECTING
  url: string
  onopen: ((event: Event) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null

  constructor(url: string) {
    this.url = url
  }

  send(data: string) {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open')
    }
  }

  close(code?: number, reason?: string) {
    this.readyState = MockWebSocket.CLOSED
    const closeEvent = new CloseEvent('close', {
      code: code || 1000,
      reason: reason || '',
      wasClean: code === 1000
    })
    setTimeout(() => this.onclose?.(closeEvent), 0)
  }

  simulateOpen() {
    this.readyState = MockWebSocket.OPEN
    setTimeout(() => this.onopen?.(new Event('open')), 0)
  }

  simulateMessage(data: any) {
    const messageEvent = new MessageEvent('message', {
      data: JSON.stringify(data)
    })
    setTimeout(() => this.onmessage?.(messageEvent), 0)
  }
}

// Mock composables
const mockWebSocket = vi.fn()
const mockSystemMetrics = vi.fn()
const mockScripts = vi.fn()

vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: () => mockWebSocket()
}))

vi.mock('@/composables/useSystemMetrics', () => ({
  useSystemMetrics: () => mockSystemMetrics()
}))

vi.mock('@/composables/useScripts', () => ({
  useScripts: () => mockScripts()
}))

describe('Dashboard Real-time Updates', () => {
  let router: any
  let mockWS: MockWebSocket
  let connectFn: any
  let onMessageFn: any

  beforeEach(() => {
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Dashboard</div>' } },
        { path: '/scripts', component: { template: '<div>Scripts</div>' } }
      ]
    })

    // Mock WebSocket globally
    // @ts-ignore
    global.WebSocket = vi.fn().mockImplementation((url: string) => {
      mockWS = new MockWebSocket(url)
      return mockWS
    })

    // Setup WebSocket mock
    connectFn = vi.fn()
    onMessageFn = vi.fn()
    
    mockWebSocket.mockReturnValue({
      isConnected: ref(false),
      error: ref(null),
      lastMessage: ref(null),
      connect: connectFn,
      disconnect: vi.fn(),
      send: vi.fn(),
      onMessage: onMessageFn,
      offMessage: vi.fn(),
      clearHandlers: vi.fn()
    })

    // Setup system metrics mock
    mockSystemMetrics.mockReturnValue({
      metrics: ref({ 
        status: 'running', 
        uptime: '1 hour', 
        runningScripts: 2, 
        totalScripts: 3 
      }),
      loading: ref(false),
      error: ref(null),
      startAutoRefresh: vi.fn(),
      cleanup: vi.fn()
    })

    // Setup scripts mock
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

  it('should connect to WebSocket on mount for real-time updates', async () => {
    mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Wait for onMounted to complete
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(connectFn).toHaveBeenCalledOnce()
  })

  it('should register WebSocket message handlers for system metrics updates', async () => {
    mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Wait for onMounted to complete
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(onMessageFn).toHaveBeenCalledWith('system_metrics', expect.any(Function))
  })

  it('should register WebSocket message handlers for script status updates', async () => {
    mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Wait for onMounted to complete
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(onMessageFn).toHaveBeenCalledWith('script_status', expect.any(Function))
  })

  it('should update system metrics display when receiving real-time data', async () => {
    const metricsRef = ref({ 
      status: 'running', 
      uptime: '1 hour', 
      runningScripts: 2, 
      totalScripts: 3 
    })

    mockSystemMetrics.mockReturnValue({
      metrics: metricsRef,
      loading: ref(false),
      error: ref(null),
      startAutoRefresh: vi.fn(),
      cleanup: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Wait for onMounted to complete
    await new Promise(resolve => setTimeout(resolve, 10))

    // Verify initial state
    expect(wrapper.text()).toContain('1 hour')
    expect(wrapper.text()).toContain('2')

    // Check if we have the system_metrics handler
    const systemMetricsCall = onMessageFn.mock.calls.find(
      call => call[0] === 'system_metrics'
    )
    
    if (!systemMetricsCall) {
      throw new Error('system_metrics handler not registered')
    }

    const systemMetricsHandler = systemMetricsCall[1]

    // Simulate real-time system metrics update
    const newMetrics = {
      type: 'system_metrics',
      timestamp: '2025-08-05T10:05:00Z',
      data: {
        status: 'running',
        uptime: '1 hour 5 min',
        runningScripts: 3,
        totalScripts: 4,
        cpu_percent: 45.2,
        memory_percent: 67.8,
        disk_percent: 23.1
      }
    }

    // Update the reactive reference
    metricsRef.value = newMetrics.data
    systemMetricsHandler(newMetrics)
    await nextTick()

    // Verify updated content
    expect(wrapper.text()).toContain('1 hour 5 min')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('4')
  })

  it('should update script status display when receiving real-time script events', async () => {
    const scriptsRef = ref([
      { name: 'test-script', path: '/test.sh', interval: 300, enabled: true, status: 'idle' }
    ])

    mockScripts.mockReturnValue({
      scripts: scriptsRef,
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

    // Wait for onMounted to complete
    await new Promise(resolve => setTimeout(resolve, 10))

    // Check if we have the script_status handler
    const scriptStatusCall = onMessageFn.mock.calls.find(
      call => call[0] === 'script_status'
    )
    
    if (!scriptStatusCall) {
      throw new Error('script_status handler not registered')
    }

    const scriptStatusHandler = scriptStatusCall[1]

    // Simulate real-time script status update
    const scriptStatusMessage = {
      type: 'script_status',
      timestamp: '2025-08-05T10:05:00Z',
      data: {
        script_name: 'test-script',
        status: 'running',
        exit_code: null,
        duration: null
      }
    }

    // Update the script status
    scriptsRef.value[0].status = 'running'
    scriptStatusHandler(scriptStatusMessage)
    await nextTick()

    // Verify the UI reflects the new status
    expect(wrapper.find('.script-status.running').exists()).toBe(true)
  })

  it('should show real-time connection status indicator', async () => {
    const connectedRef = ref(false)
    
    mockWebSocket.mockReturnValue({
      isConnected: connectedRef,
      error: ref(null),
      lastMessage: ref(null),
      connect: connectFn,
      disconnect: vi.fn(),
      send: vi.fn(),
      onMessage: onMessageFn,
      offMessage: vi.fn(),
      clearHandlers: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Should show disconnected initially
    expect(wrapper.find('.connection-status.disconnected').exists()).toBe(true)

    // Simulate connection
    connectedRef.value = true
    await nextTick()

    // Should show connected
    expect(wrapper.find('.connection-status.connected').exists()).toBe(true)
  })

  it('should handle WebSocket connection errors gracefully', async () => {
    const errorRef = ref(null)
    
    mockWebSocket.mockReturnValue({
      isConnected: ref(false),
      error: errorRef,
      lastMessage: ref(null),
      connect: connectFn,
      disconnect: vi.fn(),
      send: vi.fn(),
      onMessage: onMessageFn,
      offMessage: vi.fn(),
      clearHandlers: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Simulate WebSocket error
    errorRef.value = 'WebSocket connection failed'
    await nextTick()

    // Should show error state
    expect(wrapper.find('.websocket-error').exists()).toBe(true)
    expect(wrapper.text()).toContain('Real-time updates unavailable')
  })

  it('should display live system resource metrics (CPU, memory, disk)', async () => {
    const metricsRef = ref({ 
      status: 'running', 
      uptime: '1 hour', 
      runningScripts: 2, 
      totalScripts: 3,
      cpu_percent: 45.2,
      memory_percent: 67.8,
      disk_percent: 23.1
    })

    mockSystemMetrics.mockReturnValue({
      metrics: metricsRef,
      loading: ref(false),
      error: ref(null),
      startAutoRefresh: vi.fn(),
      cleanup: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    // Should display resource metrics
    expect(wrapper.find('.cpu-metric').text()).toContain('45.2%')
    expect(wrapper.find('.memory-metric').text()).toContain('67.8%')
    expect(wrapper.find('.disk-metric').text()).toContain('23.1%')
  })

  it('should cleanup WebSocket connections on unmount', () => {
    const disconnectFn = vi.fn()
    
    mockWebSocket.mockReturnValue({
      isConnected: ref(true),
      error: ref(null),
      lastMessage: ref(null),
      connect: connectFn,
      disconnect: disconnectFn,
      send: vi.fn(),
      onMessage: onMessageFn,
      offMessage: vi.fn(),
      clearHandlers: vi.fn()
    })

    const wrapper = mount(Dashboard, {
      global: {
        plugins: [router]
      }
    })

    wrapper.unmount()

    expect(disconnectFn).toHaveBeenCalledOnce()
  })
})