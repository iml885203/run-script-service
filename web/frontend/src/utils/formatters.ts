export function formatPercentage(value: number, decimals = 1): string {
  if (typeof value !== 'number' || isNaN(value) || value === null || value === undefined) {
    return 'N/A'
  }
  return `${value.toFixed(decimals)}%`
}

export function formatMemoryUsage(value: number): string {
  return formatPercentage(value, 1)
}

export function formatDiskUsage(value: number): string {
  return formatPercentage(value, 1)
}

export function formatCpuUsage(value: number): string {
  return formatPercentage(value, 1)
}
