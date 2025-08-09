import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { GitService, type GitProject } from '@/services/gitService'
import { ApiService } from '@/services/api'

// Mock the api module
vi.mock('@/services/api', () => ({
  ApiService: {
    getGitProjects: vi.fn()
  }
}))

describe('GitService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('getGitProjects', () => {
    it('should fetch and return git projects from API', async () => {
      // Arrange
      const mockProjects: GitProject[] = [
        {
          name: 'test-project',
          path: '/home/user/test-project',
          description: 'A test project',
          lastCommit: 'abc123 Initial commit'
        },
        {
          name: 'another-project',
          path: '/home/user/another-project'
        }
      ]

      const mockResponse = {
        projects: mockProjects
      }

      vi.mocked(ApiService.getGitProjects).mockResolvedValue(mockResponse)

      // Act
      const result = await GitService.getGitProjects()

      // Assert
      expect(ApiService.getGitProjects).toHaveBeenCalledWith(undefined)
      expect(result).toEqual(mockProjects)
    })

    it('should pass directory parameter when provided', async () => {
      // Arrange
      const mockProjects: GitProject[] = []
      const mockResponse = { projects: mockProjects }
      vi.mocked(ApiService.getGitProjects).mockResolvedValue(mockResponse)

      const testDir = '/custom/directory'

      // Act
      await GitService.getGitProjects(testDir)

      // Assert
      expect(ApiService.getGitProjects).toHaveBeenCalledWith(testDir)
    })

    it('should handle API errors gracefully', async () => {
      // Arrange
      const errorMessage = 'Network error'
      vi.mocked(ApiService.getGitProjects).mockRejectedValue(new Error(errorMessage))

      // Act & Assert
      await expect(GitService.getGitProjects()).rejects.toThrow(errorMessage)
    })

    it('should handle empty projects response', async () => {
      // Arrange
      const mockResponse = { projects: [] }
      vi.mocked(ApiService.getGitProjects).mockResolvedValue(mockResponse)

      // Act
      const result = await GitService.getGitProjects()

      // Assert
      expect(result).toEqual([])
    })
  })
})
