package mocks

import (
	"os"
	"testing"
	"time"
)

func TestMockTime_Now(t *testing.T) {
	mockTime := NewMockTime()
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test default behavior - should return current time
	now1 := mockTime.Now()
	now2 := mockTime.Now()
	if now2.Before(now1) {
		t.Error("Expected time to advance or stay same, but went backwards")
	}

	// Test with fixed time
	mockTime.SetFixedTime(fixedTime)
	result := mockTime.Now()
	if !result.Equal(fixedTime) {
		t.Errorf("Expected %v, got %v", fixedTime, result)
	}

	// Multiple calls should return same fixed time
	result2 := mockTime.Now()
	if !result2.Equal(fixedTime) {
		t.Errorf("Expected %v, got %v", fixedTime, result2)
	}
}

func TestMockTime_Sleep(t *testing.T) {
	mockTime := NewMockTime()

	// Test that Sleep doesn't actually sleep (should be instant)
	start := time.Now()
	mockTime.Sleep(1 * time.Second)
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("Sleep took too long: %v", elapsed)
	}
}

func TestMockTime_After(t *testing.T) {
	mockTime := NewMockTime()

	// Test After returns a channel that receives immediately in mock
	ch := mockTime.After(1 * time.Second)

	select {
	case <-ch:
		// Expected - channel should receive immediately
	case <-time.After(100 * time.Millisecond):
		t.Error("Mock After channel should receive immediately")
	}
}

func TestNewMockFileSystem(t *testing.T) {
	fs := NewMockFileSystem()

	if fs == nil {
		t.Error("NewMockFileSystem should return a non-nil instance")
	}

	if fs.Files == nil {
		t.Error("NewMockFileSystem should initialize Files map")
	}

	if len(fs.Files) != 0 {
		t.Error("NewMockFileSystem should create empty Files map")
	}
}

func TestMockFileSystem_WriteFile(t *testing.T) {
	fs := NewMockFileSystem()
	testFile := "test.txt"
	testData := []byte("test content")
	testPerm := os.FileMode(0644)

	// Test successful write
	err := fs.WriteFile(testFile, testData, testPerm)
	if err != nil {
		t.Errorf("WriteFile should succeed, got error: %v", err)
	}

	// Check that data was stored
	if data, exists := fs.Files[testFile]; !exists {
		t.Error("File should exist after WriteFile")
	} else if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}
}

func TestMockFileSystem_WriteFile_WithCustomFunc(t *testing.T) {
	fs := NewMockFileSystem()
	expectedError := os.ErrPermission

	fs.WriteFileFunc = func(filename string, data []byte, perm os.FileMode) error {
		return expectedError
	}

	err := fs.WriteFile("test.txt", []byte("data"), 0644)
	if err != expectedError {
		t.Errorf("Expected custom error %v, got %v", expectedError, err)
	}
}

func TestMockFileSystem_ReadFile(t *testing.T) {
	fs := NewMockFileSystem()
	testFile := "test.txt"
	testData := []byte("test content")

	// Test reading non-existent file
	_, err := fs.ReadFile(testFile)
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}

	// Add file and test reading
	fs.Files[testFile] = testData
	data, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile should succeed, got error: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}
}

func TestMockFileSystem_ReadFile_WithCustomFunc(t *testing.T) {
	fs := NewMockFileSystem()
	expectedData := []byte("custom data")
	expectedError := os.ErrPermission

	// Test custom function returns data
	fs.ReadFileFunc = func(filename string) ([]byte, error) {
		return expectedData, nil
	}

	data, err := fs.ReadFile("test.txt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(data) != string(expectedData) {
		t.Errorf("Expected %s, got %s", string(expectedData), string(data))
	}

	// Test custom function returns error
	fs.ReadFileFunc = func(filename string) ([]byte, error) {
		return nil, expectedError
	}

	_, err = fs.ReadFile("test.txt")
	if err != expectedError {
		t.Errorf("Expected custom error %v, got %v", expectedError, err)
	}
}

func TestMockFileSystem_OpenFile(t *testing.T) {
	fs := NewMockFileSystem()

	// OpenFile is not implemented in mock, should always return error
	file, err := fs.OpenFile("test.txt", os.O_RDONLY, 0644)
	if file != nil {
		t.Error("OpenFile should return nil file in mock")
	}
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestMockFileSystem_Stat(t *testing.T) {
	fs := NewMockFileSystem()
	testFile := "test.txt"

	// Test stat on non-existent file
	_, err := fs.Stat(testFile)
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}

	// Add file and test stat
	fs.Files[testFile] = []byte("content")
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Errorf("Stat should succeed for existing file, got error: %v", err)
	}
	if info == nil {
		t.Error("Stat should return non-nil FileInfo")
	}
	if info.Name() != testFile {
		t.Errorf("Expected name %s, got %s", testFile, info.Name())
	}
}

func TestMockFileSystem_Stat_WithCustomFunc(t *testing.T) {
	fs := NewMockFileSystem()
	expectedError := os.ErrPermission

	fs.StatFunc = func(name string) (os.FileInfo, error) {
		return nil, expectedError
	}

	_, err := fs.Stat("test.txt")
	if err != expectedError {
		t.Errorf("Expected custom error %v, got %v", expectedError, err)
	}
}

func TestMockFileInfo(t *testing.T) {
	info := &mockFileInfo{name: "test.txt"}

	// Test all FileInfo methods
	if info.Name() != "test.txt" {
		t.Errorf("Expected name 'test.txt', got %s", info.Name())
	}

	if info.Size() != 0 {
		t.Errorf("Expected size 0, got %d", info.Size())
	}

	if info.Mode() != 0644 {
		t.Errorf("Expected mode 0644, got %v", info.Mode())
	}

	if !info.ModTime().IsZero() {
		t.Errorf("Expected zero time, got %v", info.ModTime())
	}

	if info.IsDir() != false {
		t.Error("Expected IsDir() to return false")
	}

	if info.Sys() != nil {
		t.Error("Expected Sys() to return nil")
	}
}
