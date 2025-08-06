export interface ScriptConfig {
  name: string
  path: string
  interval: number
  enabled: boolean
  timeout?: number
  status?: 'running' | 'completed' | 'failed' | 'idle'
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
  cpu_percent?: number
  memory_percent?: number
  disk_percent?: number
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
