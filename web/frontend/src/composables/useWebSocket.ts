import { ref, onUnmounted, type Ref } from 'vue'

interface WebSocketMessage {
  type: string
  data: any
  timestamp: string
}

type MessageHandler = (message: WebSocketMessage) => void

export function useWebSocket(url?: string) {
  const isConnected = ref(false)
  const error = ref<string | null>(null)
  const lastMessage: Ref<WebSocketMessage | null> = ref(null)

  let ws: WebSocket | null = null
  let reconnectAttempts = 0
  const maxReconnectAttempts = 5
  let reconnectTimeoutId: number | null = null
  const messageHandlers = new Map<string, MessageHandler[]>()

  const connect = (wsUrl?: string): void => {
    const targetUrl = wsUrl || url || `ws://${window.location.host}/ws`

    try {
      ws = new WebSocket(targetUrl)

      ws.onopen = () => {
        isConnected.value = true
        error.value = null
        reconnectAttempts = 0
        console.log('WebSocket connected')
      }

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          lastMessage.value = message

          // Call registered handlers for this message type
          const handlers = messageHandlers.get(message.type) || []
          handlers.forEach(handler => handler(message))

          // Call global handlers
          const globalHandlers = messageHandlers.get('*') || []
          globalHandlers.forEach(handler => handler(message))
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }

      ws.onclose = (event) => {
        isConnected.value = false
        ws = null

        if (!event.wasClean && reconnectAttempts < maxReconnectAttempts) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
          reconnectTimeoutId = window.setTimeout(() => {
            reconnectAttempts++
            connect(wsUrl)
          }, delay)
        }
      }

      ws.onerror = (event) => {
        error.value = 'WebSocket connection error'
        console.error('WebSocket error:', event)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create WebSocket connection'
    }
  }

  const disconnect = (): void => {
    if (reconnectTimeoutId) {
      clearTimeout(reconnectTimeoutId)
      reconnectTimeoutId = null
    }

    if (ws) {
      ws.close(1000, 'Manual disconnect')
      ws = null
    }

    isConnected.value = false
  }

  const send = (message: any): boolean => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      error.value = 'WebSocket is not connected'
      return false
    }

    try {
      ws.send(JSON.stringify(message))
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to send message'
      return false
    }
  }

  const onMessage = (type: string, handler: MessageHandler): void => {
    if (!messageHandlers.has(type)) {
      messageHandlers.set(type, [])
    }
    messageHandlers.get(type)!.push(handler)
  }

  const offMessage = (type: string, handler: MessageHandler): void => {
    const handlers = messageHandlers.get(type)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  const clearHandlers = (type?: string): void => {
    if (type) {
      messageHandlers.delete(type)
    } else {
      messageHandlers.clear()
    }
  }

  // Cleanup on component unmount
  onUnmounted(() => {
    disconnect()
    clearHandlers()
  })

  return {
    isConnected: isConnected as Readonly<Ref<boolean>>,
    error: error as Readonly<Ref<string | null>>,
    lastMessage: lastMessage as Readonly<Ref<WebSocketMessage | null>>,
    connect,
    disconnect,
    send,
    onMessage,
    offMessage,
    clearHandlers
  }
}
