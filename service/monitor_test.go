package service

import (
	"context"
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

func TestSystemMonitor_PeriodicBroadcasting(t *testing.T) {
	monitor := NewSystemMonitor()

	// Mock event publisher
	var broadcastedMessages []map[string]interface{}
	mockPublisher := func(msgType string, data map[string]interface{}) error {
		if msgType == "system_metrics" {
			broadcastedMessages = append(broadcastedMessages, data)
		}
		return nil
	}

	// Start periodic broadcasting with short interval for testing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitor.StartPeriodicBroadcasting(ctx, 50*time.Millisecond, mockPublisher)

	// Wait for at least 2 broadcasts
	time.Sleep(150 * time.Millisecond)
	cancel()

	// Give some time for the goroutine to finish
	time.Sleep(10 * time.Millisecond)

	if len(broadcastedMessages) < 2 {
		t.Errorf("Expected at least 2 broadcasted messages, got: %d", len(broadcastedMessages))
	}

	// Verify message structure
	for i, msg := range broadcastedMessages {
		if _, ok := msg["cpu_percent"]; !ok {
			t.Errorf("Message %d missing cpu_percent field", i)
		}
		if _, ok := msg["memory_percent"]; !ok {
			t.Errorf("Message %d missing memory_percent field", i)
		}
		if _, ok := msg["disk_percent"]; !ok {
			t.Errorf("Message %d missing disk_percent field", i)
		}
		if _, ok := msg["active_scripts"]; !ok {
			t.Errorf("Message %d missing active_scripts field", i)
		}
		if _, ok := msg["total_executions"]; !ok {
			t.Errorf("Message %d missing total_executions field", i)
		}
	}
}
