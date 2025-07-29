package mocks

import (
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
