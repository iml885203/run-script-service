import type { ScriptConfig, LogEntry, SystemMetrics, ServiceConfig, ApiResponse } from '@/types/api'

export class ApiService {
  private static readonly BASE_URL = '/api'

  private static async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.BASE_URL}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    })

    if (!response.ok) {
      throw new Error(`API request failed: ${response.status} ${response.statusText}`)
    }

    return response.json()
  }

  static async getScripts(): Promise<ScriptConfig[]> {
    return this.request<ScriptConfig[]>('/scripts')
  }

  static async addScript(script: Omit<ScriptConfig, 'name'>): Promise<void> {
    await this.request('/scripts', {
      method: 'POST',
      body: JSON.stringify(script),
    })
  }

  static async updateScript(name: string, script: Partial<ScriptConfig>): Promise<void> {
    await this.request(`/scripts/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(script),
    })
  }

  static async deleteScript(name: string): Promise<void> {
    await this.request(`/scripts/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    })
  }

  static async runScript(name: string): Promise<void> {
    await this.request(`/scripts/${encodeURIComponent(name)}/run`, {
      method: 'POST',
    })
  }

  static async getLogs(scriptName?: string, limit: number = 50): Promise<LogEntry[]> {
    const params = new URLSearchParams()
    if (scriptName) params.set('script', scriptName)
    params.set('limit', limit.toString())

    return this.request<LogEntry[]>(`/logs?${params.toString()}`)
  }

  static async clearLogs(scriptName?: string): Promise<void> {
    const endpoint = scriptName ? `/logs/${encodeURIComponent(scriptName)}` : '/logs'
    await this.request(endpoint, { method: 'DELETE' })
  }

  static async getStatus(): Promise<SystemMetrics> {
    return this.request<SystemMetrics>('/status')
  }

  static async getConfig(): Promise<ServiceConfig> {
    return this.request<ServiceConfig>('/config')
  }

  static async updateConfig(config: Partial<ServiceConfig>): Promise<void> {
    await this.request('/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  }
}
