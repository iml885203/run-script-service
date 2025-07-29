package mocks

import (
	"os"
	"time"
)

// FileSystem interface abstracts file system operations for testing
type FileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
	ReadFile(filename string) ([]byte, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

// MockFileSystem provides a mock implementation for testing
type MockFileSystem struct {
	Files         map[string][]byte
	WriteFileFunc func(filename string, data []byte, perm os.FileMode) error
	ReadFileFunc  func(filename string) ([]byte, error)
	StatFunc      func(name string) (os.FileInfo, error)
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
	}
}

func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(filename, data, perm)
	}
	m.Files[filename] = data
	return nil
}

func (m *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(filename)
	}
	if data, exists := m.Files[filename]; exists {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// For simplicity, this mock doesn't implement OpenFile
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(name)
	}
	if _, exists := m.Files[name]; exists {
		return &mockFileInfo{name: name}, nil
	}
	return nil, os.ErrNotExist
}

type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }
