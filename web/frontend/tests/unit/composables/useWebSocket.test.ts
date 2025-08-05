import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useWebSocket } from '@/composables/useWebSocket'

// Mock WebSocket
class MockWebSocket {
  url: string
  readyState: number = WebSocket.CONNECTING
  onopen: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null

  constructor(url: string) {
    this.url = url
    // Simulate async connection
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 10)
  }

  send(data: string) {
    if (this.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket is not open')
    }
    // Echo the message back for testing
    setTimeout(() => {
      if (this.onmessage) {
        this.onmessage(new MessageEvent('message', { data }))
      }
    }, 10)
  }

  close() {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close'))
    }
  }

  // Static constants
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3
}

// Replace global WebSocket with mock
Object.defineProperty(global, 'WebSocket', {
  writable: true,
  value: MockWebSocket
})

describe('useWebSocket', () => {
  beforeEach(() => {
    vi.clearAllTimers()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('should establish WebSocket connection', async () => {
    const { connect, isConnected } = useWebSocket('ws://localhost:8080/ws')

    connect()

    // Wait for connection to establish
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(true)
  })

  it('should handle connection errors', async () => {
    // Mock WebSocket that throws error
    const ErrorWebSocket = class extends MockWebSocket {
      constructor(url: string) {
        super(url)
        setTimeout(() => {
          this.readyState = WebSocket.CLOSED
          if (this.onerror) {
            this.onerror(new Event('error'))
          }
        }, 10)
      }
    }

    global.WebSocket = ErrorWebSocket as any

    const { connect, isConnected, error } = useWebSocket('ws://invalid-url')

    connect()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(false)
    expect(error.value).toBeTruthy()
  })

  it('should receive messages', async () => {
    const { connect, onMessage } = useWebSocket('ws://localhost:8080/ws')

    const receivedMessages: any[] = []
    onMessage((data) => {
      receivedMessages.push(data)
    })

    connect()
    await vi.advanceTimersByTimeAsync(50)

    // Simulate receiving a message
    const testMessage = { type: 'system_metrics', data: { cpu: 50 } }
    const ws = (global as any).lastWebSocket || new MockWebSocket('test')

    if (ws.onmessage) {
      ws.onmessage(new MessageEvent('message', {
        data: JSON.stringify(testMessage)
      }))
    }

    await vi.advanceTimersByTimeAsync(50)

    expect(receivedMessages).toHaveLength(1)
    expect(receivedMessages[0]).toEqual(testMessage)
  })

  it('should send messages', async () => {
    const { connect, send, isConnected } = useWebSocket('ws://localhost:8080/ws')

    connect()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(true)

    const testMessage = { type: 'ping', data: 'hello' }
    const result = send(testMessage)

    expect(result).toBe(true)
  })

  it('should handle connection close', async () => {
    const { connect, disconnect, isConnected } = useWebSocket('ws://localhost:8080/ws')

    connect()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(true)

    disconnect()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(false)
  })

  it('should auto-reconnect on connection loss', async () => {
    const { connect, isConnected } = useWebSocket('ws://localhost:8080/ws', {
      autoReconnect: true,
      reconnectDelay: 1000
    })

    connect()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(true)

    // Simulate connection loss
    const ws = (global as any).lastWebSocket || new MockWebSocket('test')
    if (ws.onclose) {
      ws.onclose(new CloseEvent('close'))
    }

    expect(isConnected.value).toBe(false)

    // Wait for reconnection attempt
    await vi.advanceTimersByTimeAsync(1100)

    expect(isConnected.value).toBe(true)
  })

  it('should handle message filtering', async () => {
    const { connect, onMessage } = useWebSocket('ws://localhost:8080/ws')

    const systemMessages: any[] = []
    const logMessages: any[] = []

    onMessage((data) => {
      if (data.type === 'system_metrics') {
        systemMessages.push(data)
      } else if (data.type === 'log_update') {
        logMessages.push(data)
      }
    })

    connect()
    await vi.advanceTimersByTimeAsync(50)

    const ws = (global as any).lastWebSocket || new MockWebSocket('test')

    // Send different message types
    const messages = [
      { type: 'system_metrics', data: { cpu: 50 } },
      { type: 'log_update', data: { message: 'test log' } },
      { type: 'system_metrics', data: { cpu: 60 } }
    ]

    messages.forEach(msg => {
      if (ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify(msg)
        }))
      }
    })

    await vi.advanceTimersByTimeAsync(50)

    expect(systemMessages).toHaveLength(2)
    expect(logMessages).toHaveLength(1)
    expect(systemMessages[0].data.cpu).toBe(50)
    expect(systemMessages[1].data.cpu).toBe(60)
    expect(logMessages[0].data.message).toBe('test log')
  })

  it('should cleanup on unmount', async () => {
    const { connect, cleanup, isConnected } = useWebSocket('ws://localhost:8080/ws')

    connect()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(true)

    cleanup()
    await vi.advanceTimersByTimeAsync(50)

    expect(isConnected.value).toBe(false)
  })
})
