package service

import (
	"fmt"
	"os"
	"path/filepath"
)

// GitProject represents a discovered Git project
type GitProject struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	LastCommit  string `json:"last_commit,omitempty"`
}

// GitDiscoveryService handles Git project discovery in directories
type GitDiscoveryService struct {
}

// NewGitDiscoveryService creates a new Git discovery service
func NewGitDiscoveryService() *GitDiscoveryService {
	return &GitDiscoveryService{}
}

// DiscoverGitProjects scans a directory for Git projects
func (gds *GitDiscoveryService) DiscoverGitProjects(rootDir string) ([]GitProject, error) {
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", rootDir)
	}

	var projects []GitProject

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible directories
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			// The parent directory is a Git project
			projectPath := filepath.Dir(path)
			projectName := filepath.Base(projectPath)

			project := GitProject{
				Name: projectName,
				Path: projectPath,
			}

			projects = append(projects, project)
			return filepath.SkipDir // Skip walking inside .git directory
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return projects, nil
}
