// Package main provides the run-script-service daemon executable.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"run-script-service/service"
)

// CommandResult represents the result of command processing
type CommandResult struct {
	shouldRunService bool
}

// handleCommand processes command line arguments and returns appropriate action
func handleCommand(args []string, scriptPath, logPath, configPath string, maxLines int) (CommandResult, error) {
	svc := service.NewService(scriptPath, logPath, configPath, maxLines)

	if len(args) < 2 {
		return CommandResult{shouldRunService: true}, nil
	}

	command := args[1]

	switch command {
	case "run":
		return CommandResult{shouldRunService: true}, nil
	case "set-interval":
		if len(args) != 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service set-interval <interval>\nexamples: 30s, 5m, 1h, 3600")
		}
		interval, err := parseInterval(args[2])
		if err != nil {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("invalid interval: %v", err)
		}
		if err := svc.SetInterval(interval); err != nil {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("error setting interval: %v", err)
		}
		return CommandResult{shouldRunService: false}, nil
	case "show-config":
		svc.ShowConfig()
		return CommandResult{shouldRunService: false}, nil
	default:
		return CommandResult{shouldRunService: false},
			fmt.Errorf("unknown command: %s\navailable commands: run, set-interval, show-config", command)
	}
}

func main() {
	// Get paths relative to executable
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	scriptPath := filepath.Join(dir, "run.sh")
	logPath := filepath.Join(dir, "run.log")
	configPath := filepath.Join(dir, "service_config.json")
	maxLines := 100

	result, err := handleCommand(os.Args, scriptPath, logPath, configPath, maxLines)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if result.shouldRunService {
		svc := service.NewService(scriptPath, logPath, configPath, maxLines)
		runService(svc)
	}
}

func runService(svc *service.Service) {
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service in a goroutine
	done := make(chan bool)
	go func() {
		svc.Start(ctx)
		done <- true
	}()

	// Wait for signal or service completion
	select {
	case <-sigChan:
		fmt.Println("Received shutdown signal")
		svc.Stop()
		cancel()
		<-done // Wait for service to finish
	case <-done:
		// Service finished naturally
	}
}

func parseInterval(intervalStr string) (int, error) {
	if intervalStr == "" {
		return 0, fmt.Errorf("empty interval")
	}

	suffix := intervalStr[len(intervalStr)-1:]
	valueStr := intervalStr[:len(intervalStr)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// Try parsing as plain number (seconds)
		result, parseErr := strconv.Atoi(intervalStr)
		if parseErr != nil {
			return 0, parseErr
		}
		if result < 0 {
			return 0, fmt.Errorf("negative interval not allowed")
		}
		return result, nil
	}

	switch suffix {
	case "s":
		if value < 0 {
			return 0, fmt.Errorf("negative interval not allowed")
		}
		return value, nil
	case "m":
		if value < 0 {
			return 0, fmt.Errorf("negative interval not allowed")
		}
		return value * 60, nil
	case "h":
		if value < 0 {
			return 0, fmt.Errorf("negative interval not allowed")
		}
		return value * 3600, nil
	default:
		// No suffix, treat as seconds
		result, err := strconv.Atoi(intervalStr)
		if err != nil {
			return 0, err
		}
		if result < 0 {
			return 0, fmt.Errorf("negative interval not allowed")
		}
		return result, nil
	}
}
