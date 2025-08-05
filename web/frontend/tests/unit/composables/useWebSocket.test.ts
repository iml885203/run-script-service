import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useWebSocket } from '@/composables/useWebSocket'

// Mock Vue onUnmounted to avoid warnings in tests
vi.mock('vue', async () => {
  const actual = await vi.importActual('vue') as any
  return {
    ...actual,
    onUnmounted: vi.fn()
  }
})

// Simple mock WebSocket for testing
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
    // Mock successful send
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

  // Test helpers
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

  simulateError() {
    setTimeout(() => this.onerror?.(new Event('error')), 0)
  }
}

describe('useWebSocket', () => {
  let mockWebSocket: MockWebSocket

  beforeEach(() => {
    vi.clearAllMocks()
    // @ts-ignore
    global.WebSocket = vi.fn().mockImplementation((url: string) => {
      mockWebSocket = new MockWebSocket(url)
      return mockWebSocket
    })
    // Set WebSocket constants
    global.WebSocket.CONNECTING = 0
    global.WebSocket.OPEN = 1
    global.WebSocket.CLOSING = 2
    global.WebSocket.CLOSED = 3

    global.window = {
      ...global.window,
      location: { host: 'localhost:8080' }
    } as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should initialize with default values', () => {
    const { isConnected, error, lastMessage } = useWebSocket()

    expect(isConnected.value).toBe(false)
    expect(error.value).toBe(null)
    expect(lastMessage.value).toBe(null)
  })

  it('should connect to WebSocket server', async () => {
    const { connect, isConnected, error } = useWebSocket()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8080/ws')
    expect(isConnected.value).toBe(true)
    expect(error.value).toBe(null)
  })

  it('should connect with custom URL', () => {
    const customUrl = 'ws://example.com/websocket'
    const { connect } = useWebSocket()

    connect(customUrl)

    expect(global.WebSocket).toHaveBeenCalledWith(customUrl)
  })

  it('should use provided URL in constructor', () => {
    const customUrl = 'ws://test.com/ws'
    const { connect } = useWebSocket(customUrl)

    connect()

    expect(global.WebSocket).toHaveBeenCalledWith(customUrl)
  })

  it('should handle incoming messages', async () => {
    const { connect, lastMessage, onMessage } = useWebSocket()
    const messageHandler = vi.fn()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    onMessage('test_type', messageHandler)

    const testMessage = {
      type: 'test_type',
      data: { test: 'data' },
      timestamp: '2025-08-05T10:00:00Z'
    }

    mockWebSocket.simulateMessage(testMessage)
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(lastMessage.value).toEqual(testMessage)
    expect(messageHandler).toHaveBeenCalledWith(testMessage)
  })

  it('should handle global message handlers', async () => {
    const { connect, onMessage } = useWebSocket()
    const globalHandler = vi.fn()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    onMessage('*', globalHandler)

    const testMessage = {
      type: 'any_type',
      data: { test: 'data' },
      timestamp: '2025-08-05T10:00:00Z'
    }

    mockWebSocket.simulateMessage(testMessage)
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(globalHandler).toHaveBeenCalledWith(testMessage)
  })

  it('should send messages when connected', async () => {
    const { connect, send } = useWebSocket()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    // Create spy after connection is established
    const sendSpy = vi.spyOn(mockWebSocket, 'send')

    const message = { type: 'test', data: 'hello' }
    const result = send(message)

    expect(result).toBe(true)
    expect(sendSpy).toHaveBeenCalledWith(JSON.stringify(message))
  })

  it('should not send messages when disconnected', () => {
    const { send, error } = useWebSocket()

    const result = send({ type: 'test' })

    expect(result).toBe(false)
    expect(error.value).toBe('WebSocket is not connected')
  })

  it('should disconnect cleanly', async () => {
    const { connect, disconnect, isConnected } = useWebSocket()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    const closeSpy = vi.spyOn(mockWebSocket, 'close')

    disconnect()

    expect(closeSpy).toHaveBeenCalledWith(1000, 'Manual disconnect')
    await new Promise(resolve => setTimeout(resolve, 10))
    expect(isConnected.value).toBe(false)
  })

  it('should handle connection errors', async () => {
    const { connect, error } = useWebSocket()

    connect()
    mockWebSocket.simulateError()
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(error.value).toBe('WebSocket connection error')
  })

  it('should remove message handlers', async () => {
    const { connect, onMessage, offMessage } = useWebSocket()
    const handler1 = vi.fn()
    const handler2 = vi.fn()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    onMessage('test', handler1)
    onMessage('test', handler2)

    offMessage('test', handler1)

    const testMessage = {
      type: 'test',
      data: {},
      timestamp: '2025-08-05T10:00:00Z'
    }

    mockWebSocket.simulateMessage(testMessage)
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(handler1).not.toHaveBeenCalled()
    expect(handler2).toHaveBeenCalledWith(testMessage)
  })

  it('should clear all handlers', async () => {
    const { connect, onMessage, clearHandlers } = useWebSocket()
    const handler1 = vi.fn()
    const handler2 = vi.fn()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    onMessage('test1', handler1)
    onMessage('test2', handler2)

    clearHandlers()

    const testMessage = {
      type: 'test1',
      data: {},
      timestamp: '2025-08-05T10:00:00Z'
    }

    mockWebSocket.simulateMessage(testMessage)
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(handler1).not.toHaveBeenCalled()
    expect(handler2).not.toHaveBeenCalled()
  })

  it('should clear specific handler type', async () => {
    const { connect, onMessage, clearHandlers } = useWebSocket()
    const handler1 = vi.fn()
    const handler2 = vi.fn()

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    onMessage('test1', handler1)
    onMessage('test2', handler2)

    clearHandlers('test1')

    const testMessage1 = {
      type: 'test1',
      data: {},
      timestamp: '2025-08-05T10:00:00Z'
    }

    const testMessage2 = {
      type: 'test2',
      data: {},
      timestamp: '2025-08-05T10:00:00Z'
    }

    mockWebSocket.simulateMessage(testMessage1)
    mockWebSocket.simulateMessage(testMessage2)
    await new Promise(resolve => setTimeout(resolve, 10))

    expect(handler1).not.toHaveBeenCalled()
    expect(handler2).toHaveBeenCalledWith(testMessage2)
  })

  it('should handle malformed JSON messages gracefully', async () => {
    const { connect } = useWebSocket()
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    connect()
    mockWebSocket.simulateOpen()
    await new Promise(resolve => setTimeout(resolve, 10))

    // Simulate malformed JSON
    const messageEvent = new MessageEvent('message', {
      data: 'invalid json'
    })
    mockWebSocket.onmessage?.(messageEvent)

    expect(consoleSpy).toHaveBeenCalledWith('Failed to parse WebSocket message:', expect.any(Error))

    consoleSpy.mockRestore()
  })
})
