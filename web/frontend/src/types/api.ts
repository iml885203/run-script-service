export interface ScriptConfig {
  name: string
  path: string
  interval: number
  enabled: boolean
  timeout?: number
}

export interface LogEntry {
  timestamp: string
  message: string
  level: 'info' | 'warning' | 'error'
  script?: string
}

export interface SystemMetrics {
  uptime: string
  status: string
  runningScripts: number
  totalScripts: number
}

export interface ServiceConfig {
  webPort: number
  interval: string
  logRetention: number
  autoRefresh: boolean
}

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
}
