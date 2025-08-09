import { describe, it, expect } from 'vitest'
import { formatPercentage, formatMemoryUsage, formatDiskUsage, formatCpuUsage } from '@/utils/formatters'

describe('formatPercentage', () => {
  it('should format percentage with specified decimals', () => {
    expect(formatPercentage(45.6789, 1)).toBe('45.7%')
    expect(formatPercentage(45.6789, 2)).toBe('45.68%')
    expect(formatPercentage(45.6789, 0)).toBe('46%')
  })

  it('should handle invalid values', () => {
    expect(formatPercentage(NaN)).toBe('N/A')
    expect(formatPercentage(undefined as any)).toBe('N/A')
    expect(formatPercentage(null as any)).toBe('N/A')
  })

  it('should handle edge cases', () => {
    expect(formatPercentage(0, 1)).toBe('0.0%')
    expect(formatPercentage(100, 1)).toBe('100.0%')
    expect(formatPercentage(99.95, 1)).toBe('100.0%')
  })
})

describe('formatCpuUsage', () => {
  it('should format CPU usage with 1 decimal place', () => {
    expect(formatCpuUsage(45.6789)).toBe('45.7%')
    expect(formatCpuUsage(25.0)).toBe('25.0%')
  })

  it('should handle invalid CPU values', () => {
    expect(formatCpuUsage(NaN)).toBe('N/A')
    expect(formatCpuUsage(undefined as any)).toBe('N/A')
  })
})

describe('formatMemoryUsage', () => {
  it('should format memory usage with 1 decimal place', () => {
    expect(formatMemoryUsage(45.234234234)).toBe('45.2%')
    expect(formatMemoryUsage(78.87654321)).toBe('78.9%')
  })

  it('should handle invalid memory values', () => {
    expect(formatMemoryUsage(NaN)).toBe('N/A')
    expect(formatMemoryUsage(undefined as any)).toBe('N/A')
  })
})

describe('formatDiskUsage', () => {
  it('should format disk usage with 1 decimal place', () => {
    expect(formatDiskUsage(78.87654321)).toBe('78.9%')
    expect(formatDiskUsage(99.99)).toBe('100.0%')
  })

  it('should handle invalid disk values', () => {
    expect(formatDiskUsage(NaN)).toBe('N/A')
    expect(formatDiskUsage(undefined as any)).toBe('N/A')
  })
})
