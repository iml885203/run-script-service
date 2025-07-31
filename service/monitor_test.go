package service

import (
	"testing"
	"time"
)

func TestNewSystemMonitor(t *testing.T) {
	monitor := NewSystemMonitor()
	if monitor == nil {
		t.Fatal("Expected NewSystemMonitor to return a non-nil monitor")
	}
}

func TestSystemMetrics_Structure(t *testing.T) {
	monitor := NewSystemMonitor()
	metrics, err := monitor.GetSystemMetrics()
	if err != nil {
		t.Fatalf("Expected GetSystemMetrics to not return error, got: %v", err)
	}

	if metrics.CPUPercent < 0 || metrics.CPUPercent > 100 {
		t.Errorf("Expected CPUPercent to be between 0-100, got: %f", metrics.CPUPercent)
	}

	if metrics.MemoryPercent < 0 || metrics.MemoryPercent > 100 {
		t.Errorf("Expected MemoryPercent to be between 0-100, got: %f", metrics.MemoryPercent)
	}

	if metrics.DiskPercent < 0 || metrics.DiskPercent > 100 {
		t.Errorf("Expected DiskPercent to be between 0-100, got: %f", metrics.DiskPercent)
	}

	if metrics.ActiveScripts < 0 {
		t.Errorf("Expected ActiveScripts to be non-negative, got: %d", metrics.ActiveScripts)
	}

	if metrics.TotalExecutions < 0 {
		t.Errorf("Expected TotalExecutions to be non-negative, got: %d", metrics.TotalExecutions)
	}

	if metrics.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

func TestSystemMetrics_JSONSerialization(t *testing.T) {
	monitor := NewSystemMonitor()
	metrics, _ := monitor.GetSystemMetrics()

	// Test that metrics can be marshaled to JSON
	data := metrics.ToJSON()
	if len(data) == 0 {
		t.Error("Expected ToJSON to return non-empty data")
	}
}

func TestSystemMonitor_WithScriptCounts(t *testing.T) {
	monitor := NewSystemMonitor()

	// Test setting active scripts
	monitor.SetActiveScripts(5)
	monitor.SetTotalExecutions(150)

	metrics, err := monitor.GetSystemMetrics()
	if err != nil {
		t.Fatalf("Expected GetSystemMetrics to not return error, got: %v", err)
	}

	if metrics.ActiveScripts != 5 {
		t.Errorf("Expected ActiveScripts to be 5, got: %d", metrics.ActiveScripts)
	}

	if metrics.TotalExecutions != 150 {
		t.Errorf("Expected TotalExecutions to be 150, got: %d", metrics.TotalExecutions)
	}
}

func TestSystemMonitor_ConcurrentAccess(t *testing.T) {
	monitor := NewSystemMonitor()

	// Test concurrent access to metrics
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			metrics, err := monitor.GetSystemMetrics()
			if err != nil {
				t.Errorf("Concurrent access failed: %v", err)
				return
			}

			if metrics == nil {
				t.Error("Expected metrics to not be nil")
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent access test")
		}
	}
}

func TestSystemMonitor_PeriodicUpdates(t *testing.T) {
	monitor := NewSystemMonitor()

	// Get initial metrics
	metrics1, err := monitor.GetSystemMetrics()
	if err != nil {
		t.Fatalf("Expected GetSystemMetrics to not return error, got: %v", err)
	}

	// Wait a small amount and get metrics again
	time.Sleep(10 * time.Millisecond)

	metrics2, err := monitor.GetSystemMetrics()
	if err != nil {
		t.Fatalf("Expected GetSystemMetrics to not return error, got: %v", err)
	}

	// Timestamps should be different (metrics should be fresh)
	if !metrics2.Timestamp.After(metrics1.Timestamp) {
		t.Error("Expected second metrics to have later timestamp")
	}
}
