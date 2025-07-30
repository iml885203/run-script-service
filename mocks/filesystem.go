// Package mocks provides mock implementations for testing purposes.
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

// NewMockFileSystem creates a new mock file system for testing.
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
	}
}

// WriteFile implements the FileSystem interface for mock testing.
func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(filename, data, perm)
	}
	m.Files[filename] = data
	return nil
}

// ReadFile implements the FileSystem interface for mock testing.
func (m *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(filename)
	}
	if data, exists := m.Files[filename]; exists {
		return data, nil
	}
	return nil, os.ErrNotExist
}

// OpenFile implements the FileSystem interface for mock testing.
func (m *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// For simplicity, this mock doesn't implement OpenFile
	return nil, os.ErrNotExist
}

// Stat implements the FileSystem interface for mock testing.
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

// TimeProvider interface abstracts time operations for testing
type TimeProvider interface {
	Now() time.Time
	Sleep(d time.Duration)
	After(d time.Duration) <-chan time.Time
}

// MockTime provides a mock implementation for testing
type MockTime struct {
	fixedTime *time.Time
}

// NewMockTime creates a new mock time provider for testing.
func NewMockTime() *MockTime {
	return &MockTime{}
}

// SetFixedTime sets a fixed time for testing purposes.
func (m *MockTime) SetFixedTime(t time.Time) {
	m.fixedTime = &t
}

// Now implements the TimeProvider interface for mock testing.
func (m *MockTime) Now() time.Time {
	if m.fixedTime != nil {
		return *m.fixedTime
	}
	return time.Now()
}

// Sleep implements the TimeProvider interface for mock testing.
func (m *MockTime) Sleep(d time.Duration) {
	// In mock, sleep is instant (don't actually sleep)
}

// After implements the TimeProvider interface for mock testing.
func (m *MockTime) After(d time.Duration) <-chan time.Time {
	// In mock, return a channel that receives immediately
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	return ch
}
