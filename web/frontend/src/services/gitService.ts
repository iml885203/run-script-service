import { ApiService } from './api'

export interface GitProject {
  name: string
  path: string
  description?: string
  lastCommit?: string
}

export class GitService {
  static async getGitProjects(directory?: string): Promise<GitProject[]> {
    const response = await ApiService.getGitProjects(directory)
    return response.projects as GitProject[]
  }
}
