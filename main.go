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

	svc := service.NewService(scriptPath, logPath, configPath, maxLines)

	if len(os.Args) < 2 {
		runService(svc)
		return
	}

	command := os.Args[1]

	switch command {
	case "run":
		runService(svc)
	case "set-interval":
		if len(os.Args) != 3 {
			fmt.Println("Usage: ./run-script-service set-interval <interval>")
			fmt.Println("Examples: 30s, 5m, 1h, 3600")
			os.Exit(1)
		}
		interval, err := parseInterval(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid interval: %v\n", err)
			os.Exit(1)
		}
		if err := svc.SetInterval(interval); err != nil {
			fmt.Printf("Error setting interval: %v\n", err)
			os.Exit(1)
		}
	case "show-config":
		svc.ShowConfig()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: run, set-interval, show-config")
		os.Exit(1)
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
	if len(intervalStr) == 0 {
		return 0, fmt.Errorf("empty interval")
	}

	suffix := intervalStr[len(intervalStr)-1:]
	valueStr := intervalStr[:len(intervalStr)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// Try parsing as plain number (seconds)
		return strconv.Atoi(intervalStr)
	}

	switch suffix {
	case "s":
		return value, nil
	case "m":
		return value * 60, nil
	case "h":
		return value * 3600, nil
	default:
		// No suffix, treat as seconds
		return strconv.Atoi(intervalStr)
	}
}
