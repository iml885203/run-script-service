package service

import (
	"context"
	"fmt"
	"time"
)

// Service manages the execution of scripts at regular intervals
type Service struct {
	config     Config
	scriptPath string
	logPath    string
	configPath string
	maxLines   int
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
	executor   *Executor
}

// NewService creates a new service instance
func NewService(scriptPath, logPath, configPath string, maxLines int) *Service {
	s := &Service{
		config:     Config{Interval: 3600}, // Default 1 hour
		scriptPath: scriptPath,
		logPath:    logPath,
		configPath: configPath,
		maxLines:   maxLines,
		running:    false,
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.executor = NewExecutor(scriptPath, logPath, maxLines)

	// Load existing config if available
	LoadConfig(configPath, &s.config)

	return s
}

// SetInterval sets the execution interval and saves the configuration
func (s *Service) SetInterval(interval int) error {
	s.config.Interval = interval
	if err := SaveConfig(s.configPath, &s.config); err != nil {
		return err
	}
	fmt.Printf("Interval set to %d seconds\n", interval)
	return nil
}

// Start begins the service execution loop
func (s *Service) Start(ctx context.Context) {
	s.running = true
	fmt.Printf("Service started with %d second interval\n", s.config.Interval)

	ticker := time.NewTicker(time.Duration(s.config.Interval) * time.Second)
	defer ticker.Stop()

	// Execute immediately on start
	s.executeScript()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Service stopping...")
			return
		case <-s.ctx.Done():
			fmt.Println("Service stopping...")
			return
		case <-ticker.C:
			if s.running {
				s.executeScript()
			}
		}
	}
}

// Stop stops the service
func (s *Service) Stop() {
	s.running = false
	s.cancel()
}

// ShowConfig displays the current configuration
func (s *Service) ShowConfig() {
	fmt.Printf("Current configuration:\n")
	fmt.Printf("  Interval: %d seconds (%s)\n", s.config.Interval, formatDuration(s.config.Interval))
	fmt.Printf("  Script: %s\n", s.scriptPath)
	fmt.Printf("  Log: %s\n", s.logPath)
	fmt.Printf("  Config: %s\n", s.configPath)
}

// executeScript executes the script and logs the result
func (s *Service) executeScript() {
	result := s.executor.ExecuteScript()
	fmt.Printf("Script executed at %s, exit code: %d\n",
		result.Timestamp.Format("2006-01-02 15:04:05"), result.ExitCode)
}

// formatDuration formats seconds into a human-readable duration string
func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	} else {
		return fmt.Sprintf("%dh", seconds/3600)
	}
}
