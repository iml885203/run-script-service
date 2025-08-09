package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitProjectDiscovery(t *testing.T) {
	// Create temporary directory structure with Git projects
	tempDir := t.TempDir()

	// Create test Git project 1
	project1 := filepath.Join(tempDir, "project1")
	os.MkdirAll(filepath.Join(project1, ".git"), 0755)

	// Create test Git project 2
	project2 := filepath.Join(tempDir, "project2")
	os.MkdirAll(filepath.Join(project2, ".git"), 0755)

	// Create non-Git directory
	nonGitDir := filepath.Join(tempDir, "not-git")
	os.MkdirAll(nonGitDir, 0755)

	// Test discovery
	discovery := NewGitDiscoveryService()
	projects, err := discovery.DiscoverGitProjects(tempDir)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(projects))
	}

	// Verify project 1
	var found1, found2 bool
	for _, project := range projects {
		if project.Name == "project1" {
			found1 = true
			if project.Path != project1 {
				t.Errorf("Expected path %s, got %s", project1, project.Path)
			}
		}
		if project.Name == "project2" {
			found2 = true
			if project.Path != project2 {
				t.Errorf("Expected path %s, got %s", project2, project.Path)
			}
		}
	}

	if !found1 {
		t.Error("Project1 not found in results")
	}
	if !found2 {
		t.Error("Project2 not found in results")
	}
}

func TestGitProject_Fields(t *testing.T) {
	project := GitProject{
		Name:        "test-project",
		Path:        "/path/to/project",
		Description: "Test description",
		LastCommit:  "abc123",
	}

	if project.Name != "test-project" {
		t.Errorf("Expected name 'test-project', got %s", project.Name)
	}
	if project.Path != "/path/to/project" {
		t.Errorf("Expected path '/path/to/project', got %s", project.Path)
	}
}

func TestGitDiscoveryService_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	discovery := NewGitDiscoveryService()
	projects, err := discovery.DiscoverGitProjects(tempDir)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 0 {
		t.Fatalf("Expected 0 projects in empty directory, got %d", len(projects))
	}
}

func TestGitDiscoveryService_NonExistentDirectory(t *testing.T) {
	discovery := NewGitDiscoveryService()
	_, err := discovery.DiscoverGitProjects("/non/existent/path")

	if err == nil {
		t.Fatal("Expected error for non-existent directory, got nil")
	}
}
