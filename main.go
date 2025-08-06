// Package main provides the run-script-service daemon executable.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"run-script-service/service"
	"run-script-service/web"
)

// CommandResult represents the result of command processing
type CommandResult struct {
	shouldRunService bool
	webMode          bool
}

// handleCommand processes command line arguments and returns appropriate action
func handleCommand(args []string, scriptPath, logPath, configPath string, maxLines int) (CommandResult, error) {
	svc := service.NewService(scriptPath, logPath, configPath, maxLines)

	if len(args) < 2 {
		return CommandResult{shouldRunService: true, webMode: true}, nil
	}

	command := args[1]

	switch command {
	case "run":
		// Always enable web mode by default
		result := CommandResult{shouldRunService: true, webMode: true}
		return result, nil
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
	case "add-script":
		return handleAddScript(args[2:], configPath)
	case "list-scripts":
		return handleListScripts(configPath)
	case "enable-script":
		if len(args) != 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service enable-script <script-name>")
		}
		return handleEnableScript(args[2], configPath)
	case "disable-script":
		if len(args) != 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service disable-script <script-name>")
		}
		return handleDisableScript(args[2], configPath)
	case "remove-script":
		if len(args) != 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service remove-script <script-name>")
		}
		return handleRemoveScript(args[2], configPath)
	case "run-script":
		if len(args) != 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service run-script <script-name>")
		}
		return handleRunScript(args[2], configPath)
	case "logs":
		return handleLogs(args[2:], configPath)
	case "clear-logs":
		return handleClearLogs(args[2:], configPath)
	case "set-web-port":
		if len(args) != 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service set-web-port <port>")
		}
		return handleSetWebPort(args[2], configPath)
	case "daemon":
		if len(args) < 3 {
			return CommandResult{shouldRunService: false},
				fmt.Errorf("usage: ./run-script-service daemon <start|stop|status|restart|logs>")
		}
		return handleDaemonCommand(args[2], configPath)
	default:
		availableCommands := "run, set-interval, show-config, add-script, " +
			"list-scripts, enable-script, disable-script, remove-script, run-script, logs, clear-logs, set-web-port, daemon"
		return CommandResult{shouldRunService: false},
			fmt.Errorf("unknown command: %s\navailable commands: %s", command, availableCommands)
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
		if result.webMode {
			runMultiScriptServiceWithWeb(configPath)
		} else {
			runMultiScriptService(configPath)
		}
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

func runMultiScriptService(configPath string) {
	// Load service configuration
	var config service.ServiceConfig
	err := service.LoadServiceConfig(configPath, &config)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Set default web port if not configured
	if config.WebPort == 0 {
		config.WebPort = 8080
	}

	// Create script manager
	manager := service.NewScriptManager(&config)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all enabled scripts
	err = manager.StartAllEnabled(ctx)
	if err != nil {
		fmt.Printf("Failed to start scripts: %v\n", err)
		cancel()
		os.Exit(1)
	}

	fmt.Println("Multi-script service started")
	fmt.Printf("Running scripts: %v\n", manager.GetRunningScripts())

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("Received shutdown signal")

	// Stop all scripts
	manager.StopAll()
	cancel()

	fmt.Println("Service stopped")
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

// parseScriptFlags parses command line flags for script management
func parseScriptFlags(args []string) (map[string]string, error) {
	flags := make(map[string]string)

	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		parts := strings.SplitN(arg[2:], "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid flag format: %s (expected --key=value)", arg)
		}

		flags[parts[0]] = parts[1]
	}

	// Check required flags for add-script
	required := []string{"name", "path", "interval"}
	for _, req := range required {
		if _, ok := flags[req]; !ok {
			return nil, fmt.Errorf("missing required flag: --%s", req)
		}
	}

	return flags, nil
}

// handleAddScript adds a new script to the configuration
func handleAddScript(args []string, configPath string) (CommandResult, error) {
	flags, err := parseScriptFlags(args)
	if err != nil {
		return CommandResult{shouldRunService: false}, err
	}

	// Parse interval
	interval, err := parseInterval(flags["interval"])
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("invalid interval: %v", err)
	}

	// Parse optional flags
	maxLogLines := 100
	if val, ok := flags["max-log-lines"]; ok {
		if parsed, parseErr := strconv.Atoi(val); parseErr == nil && parsed > 0 {
			maxLogLines = parsed
		}
	}

	timeout := 0
	if val, ok := flags["timeout"]; ok {
		if parsed, parseErr := strconv.Atoi(val); parseErr == nil && parsed >= 0 {
			timeout = parsed
		}
	}

	// Load existing configuration
	var config service.ServiceConfig
	err = service.LoadServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to load config: %v", err)
	}

	// Check if script name already exists
	for _, script := range config.Scripts {
		if script.Name == flags["name"] {
			return CommandResult{shouldRunService: false}, fmt.Errorf("script with name '%s' already exists", flags["name"])
		}
	}

	// Add new script
	newScript := service.ScriptConfig{
		Name:        flags["name"],
		Path:        flags["path"],
		Interval:    interval,
		Enabled:     true,
		MaxLogLines: maxLogLines,
		Timeout:     timeout,
	}

	if validateErr := newScript.Validate(); validateErr != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("invalid script configuration: %v", validateErr)
	}

	config.Scripts = append(config.Scripts, newScript)

	// Save configuration
	err = service.SaveServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Script '%s' added successfully\n", flags["name"])
	return CommandResult{shouldRunService: false}, nil
}

// handleListScripts lists all configured scripts
func handleListScripts(configPath string) (CommandResult, error) {
	var config service.ServiceConfig
	err := service.LoadServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to load config: %v", err)
	}

	if len(config.Scripts) == 0 {
		fmt.Println("No scripts configured")
		return CommandResult{shouldRunService: false}, nil
	}

	fmt.Printf("%-15s %-50s %-10s %-8s %-10s %-7s\n", "NAME", "PATH", "INTERVAL", "ENABLED", "MAX_LOGS", "TIMEOUT")
	fmt.Println(strings.Repeat("-", 100))

	for _, script := range config.Scripts {
		enabled := "false"
		if script.Enabled {
			enabled = "true"
		}

		timeout := "none"
		if script.Timeout > 0 {
			timeout = fmt.Sprintf("%ds", script.Timeout)
		}

		fmt.Printf("%-15s %-50s %-10ds %-8s %-10d %-7s\n",
			script.Name, script.Path, script.Interval, enabled, script.MaxLogLines, timeout)
	}

	return CommandResult{shouldRunService: false}, nil
}

// handleScriptToggle enables or disables a script
func handleScriptToggle(scriptName, configPath string, enable bool) (CommandResult, error) {
	var config service.ServiceConfig
	err := service.LoadServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to load config: %v", err)
	}

	found := false
	for i, script := range config.Scripts {
		if script.Name == scriptName {
			config.Scripts[i].Enabled = enable
			found = true
			break
		}
	}

	if !found {
		return CommandResult{shouldRunService: false}, fmt.Errorf("script '%s' not found", scriptName)
	}

	err = service.SaveServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to save config: %v", err)
	}

	action := "disabled"
	if enable {
		action = "enabled"
	}
	fmt.Printf("Script '%s' %s\n", scriptName, action)
	return CommandResult{shouldRunService: false}, nil
}

// handleEnableScript enables a script
func handleEnableScript(scriptName, configPath string) (CommandResult, error) {
	return handleScriptToggle(scriptName, configPath, true)
}

// handleDisableScript disables a script
func handleDisableScript(scriptName, configPath string) (CommandResult, error) {
	return handleScriptToggle(scriptName, configPath, false)
}

// handleRemoveScript removes a script from configuration
func handleRemoveScript(scriptName, configPath string) (CommandResult, error) {
	var config service.ServiceConfig
	err := service.LoadServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to load config: %v", err)
	}

	found := false
	newScripts := make([]service.ScriptConfig, 0, len(config.Scripts))
	for _, script := range config.Scripts {
		if script.Name != scriptName {
			newScripts = append(newScripts, script)
		} else {
			found = true
		}
	}

	if !found {
		return CommandResult{shouldRunService: false}, fmt.Errorf("script '%s' not found", scriptName)
	}

	config.Scripts = newScripts

	err = service.SaveServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Script '%s' removed\n", scriptName)
	return CommandResult{shouldRunService: false}, nil
}

// handleRunScript executes a script once
func handleRunScript(scriptName, configPath string) (CommandResult, error) {
	var config service.ServiceConfig
	err := service.LoadServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to load config: %v", err)
	}

	var scriptConfig *service.ScriptConfig
	for i, script := range config.Scripts {
		if script.Name == scriptName {
			scriptConfig = &config.Scripts[i]
			break
		}
	}

	if scriptConfig == nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("script '%s' not found", scriptName)
	}

	// Create a temporary script runner and execute once
	logPath := fmt.Sprintf("%s.log", scriptName)
	runner := service.NewScriptRunner(*scriptConfig, logPath)

	ctx := context.Background()
	err = runner.RunOnce(ctx)
	if err != nil {
		fmt.Printf("Script '%s' execution failed: %v\n", scriptName, err)
	} else {
		fmt.Printf("Script '%s' executed successfully\n", scriptName)
	}

	return CommandResult{shouldRunService: false}, nil
}

// handleLogs displays logs for scripts
func handleLogs(args []string, _ string) (CommandResult, error) {
	flags, err := parseLogFlags(args)
	if err != nil {
		return CommandResult{shouldRunService: false}, err
	}

	// Determine logs directory path
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}
	logsDir := filepath.Join(dir, "logs")

	// Create log manager
	logManager := service.NewLogManager(logsDir)

	// Build query
	query := &service.LogQuery{}

	if scriptName, ok := flags["script"]; ok {
		query.ScriptName = scriptName
	}

	if exitCode, ok := flags["exit-code"]; ok {
		code, parseErr := strconv.Atoi(exitCode)
		if parseErr != nil {
			return CommandResult{shouldRunService: false}, fmt.Errorf("invalid exit-code: %v", parseErr)
		}
		query.ExitCode = &code
	}

	if limit, ok := flags["limit"]; ok {
		limitNum, parseErr := strconv.Atoi(limit)
		if parseErr != nil {
			return CommandResult{shouldRunService: false}, fmt.Errorf("invalid limit: %v", parseErr)
		}
		query.Limit = limitNum
	}

	// Query logs
	entries, err := logManager.QueryLogs(query)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to query logs: %v", err)
	}

	// Display logs
	if len(entries) == 0 {
		fmt.Println("No log entries found")
	} else {
		fmt.Printf("Found %d log entries:\n\n", len(entries))
		for _, entry := range entries {
			fmt.Printf("[%s] %s (exit: %d, duration: %dms)\n",
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.ScriptName,
				entry.ExitCode,
				entry.Duration)
			if entry.Stdout != "" {
				fmt.Printf("  STDOUT: %s\n", entry.Stdout)
			}
			if entry.Stderr != "" {
				fmt.Printf("  STDERR: %s\n", entry.Stderr)
			}
			fmt.Println()
		}
	}

	return CommandResult{shouldRunService: false}, nil
}

// handleClearLogs clears logs for a specific script
func handleClearLogs(args []string, _ string) (CommandResult, error) {
	flags, err := parseLogFlags(args)
	if err != nil {
		return CommandResult{shouldRunService: false}, err
	}

	scriptName, ok := flags["script"]
	if !ok {
		return CommandResult{shouldRunService: false},
			fmt.Errorf("usage: ./run-script-service clear-logs --script=<script-name>")
	}

	// Determine logs directory path
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}
	logsDir := filepath.Join(dir, "logs")

	// Clear the specific log file
	logFile := filepath.Join(logsDir, fmt.Sprintf("%s.log", scriptName))

	if _, statErr := os.Stat(logFile); os.IsNotExist(statErr) {
		fmt.Printf("No log file found for script '%s'\n", scriptName)
		return CommandResult{shouldRunService: false}, nil
	}

	err = os.Remove(logFile)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to clear logs: %v", err)
	}

	fmt.Printf("Logs cleared for script '%s'\n", scriptName)
	return CommandResult{shouldRunService: false}, nil
}

// parseLogFlags parses log command flags
func parseLogFlags(args []string) (map[string]string, error) {
	flags := make(map[string]string)

	for _, arg := range args {
		if arg == "--all" {
			// --all is equivalent to no script filter
			continue
		}

		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(arg[2:], "=", 2)
			if len(parts) == 2 {
				flags[parts[0]] = parts[1]
			} else {
				return nil, fmt.Errorf("invalid flag format: %s (expected --key=value)", arg)
			}
		} else {
			return nil, fmt.Errorf("invalid argument: %s", arg)
		}
	}

	return flags, nil
}

// handleSetWebPort sets the web server port
func handleSetWebPort(portStr, configPath string) (CommandResult, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("invalid port number: %v", err)
	}

	if port < 1 || port > 65535 {
		return CommandResult{shouldRunService: false}, fmt.Errorf("port must be between 1 and 65535")
	}

	// Load existing configuration
	var config service.ServiceConfig
	err = service.LoadServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to load config: %v", err)
	}

	// Update web port
	config.WebPort = port

	// Save configuration
	err = service.SaveServiceConfig(configPath, &config)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Web port set to %d\n", port)
	return CommandResult{shouldRunService: false}, nil
}

// runMultiScriptServiceWithWeb runs the service with web interface
func runMultiScriptServiceWithWeb(configPath string) {
	// Load service configuration
	var config service.ServiceConfig
	err := service.LoadServiceConfig(configPath, &config)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Set default web port if not configured
	if config.WebPort == 0 {
		config.WebPort = 8080
	}

	// Get current directory for file operations
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	// Create file manager for secure file operations
	fileManager := service.NewFileManager(dir)

	// Create script manager
	scriptManager := service.NewScriptManagerWithPath(&config, configPath)

	// Create system monitor
	systemMonitor := service.NewSystemMonitor()

	// Get secret key from environment or generate one
	secretKey := os.Getenv("WEB_SECRET_KEY")
	if secretKey == "" {
		// Generate a random secret key and warn about it
		secretKey = generateRandomKey()
		fmt.Printf("WARNING: No WEB_SECRET_KEY environment variable set!\n")
		fmt.Printf("Generated random secret key: %s\n", secretKey)
		fmt.Printf("Set WEB_SECRET_KEY environment variable to use a persistent key.\n")
		fmt.Printf("For production, use: export WEB_SECRET_KEY=your-secure-secret-here\n\n")
	}

	// Create web server (simplified, no LogManager dependency)
	webServer := web.NewWebServer(nil, config.WebPort, secretKey)
	webServer.SetScriptManager(scriptManager)
	webServer.SetFileManager(fileManager)
	webServer.SetSystemMonitor(systemMonitor)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all enabled scripts
	err = scriptManager.StartAllEnabled(ctx)
	if err != nil {
		fmt.Printf("Failed to start scripts: %v\n", err)
		cancel()
		os.Exit(1)
	}

	fmt.Println("Multi-script service with web interface started")
	fmt.Printf("Running scripts: %v\n", scriptManager.GetRunningScripts())
	fmt.Printf("Web interface available at http://localhost:%d\n", config.WebPort)

	// Start system metrics broadcasting (every 30 seconds)
	err = webServer.StartSystemMetricsBroadcasting(ctx, 30*time.Second)
	if err != nil {
		fmt.Printf("Failed to start system metrics broadcasting: %v\n", err)
	} else {
		fmt.Println("System metrics broadcasting started")
	}

	// Start web server in goroutine
	go func() {
		if err := webServer.Start(); err != nil {
			fmt.Printf("Web server failed: %v\n", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("Received shutdown signal")

	// Stop all scripts and web server
	scriptManager.StopAll()
	cancel()

	fmt.Println("Service stopped")
}

// PID file management functions
func getPidFilePath() string {
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}
	return filepath.Join(dir, "run-script-service.pid")
}

func writePidFile(pid int) error {
	pidFile := getPidFilePath()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func readPidFile() (int, error) {
	pidFile := getPidFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func removePidFile() error {
	pidFile := getPidFilePath()
	return os.Remove(pidFile)
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// handleDaemonCommand handles daemon subcommands (start/stop/status/restart/logs)
func handleDaemonCommand(subCommand, configPath string) (CommandResult, error) {
	switch subCommand {
	case "start":
		return handleDaemonStart(configPath)
	case "stop":
		return handleDaemonStop()
	case "status":
		return handleDaemonStatus()
	case "restart":
		return handleDaemonRestart(configPath)
	case "logs":
		return handleDaemonLogs()
	default:
		return CommandResult{shouldRunService: false},
			fmt.Errorf("unknown daemon subcommand: %s\navailable subcommands: start, stop, status, restart, logs", subCommand)
	}
}

// handleDaemonStart starts the service as a background daemon
func handleDaemonStart(configPath string) (CommandResult, error) {
	// Check if already running
	if pid, err := readPidFile(); err == nil && isProcessRunning(pid) {
		return CommandResult{shouldRunService: false},
			fmt.Errorf("service is already running (PID: %d)", pid)
	}

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return CommandResult{shouldRunService: false},
			fmt.Errorf("failed to get executable path: %v", err)
	}

	// Get working directory
	workDir := filepath.Dir(execPath)

	// Auto-build frontend if needed
	if err := ensureFrontendBuilt(workDir); err != nil {
		fmt.Printf("Warning: Frontend build failed: %v\n", err)
		fmt.Println("Continuing with existing build...")
	}

	// Create log file for daemon output
	logFile := filepath.Join(workDir, "daemon.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return CommandResult{shouldRunService: false},
			fmt.Errorf("failed to create log file: %v", err)
	}
	defer file.Close()

	// Start the daemon process
	cmd := exec.Command(execPath, "run")
	cmd.Dir = workDir
	cmd.Stdout = file
	cmd.Stderr = file
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session
	}

	err = cmd.Start()
	if err != nil {
		return CommandResult{shouldRunService: false},
			fmt.Errorf("failed to start daemon: %v", err)
	}

	// Write PID file
	err = writePidFile(cmd.Process.Pid)
	if err != nil {
		cmd.Process.Kill()
		return CommandResult{shouldRunService: false},
			fmt.Errorf("failed to write PID file: %v", err)
	}

	fmt.Printf("Service started successfully (PID: %d)\n", cmd.Process.Pid)
	fmt.Printf("Web interface available at http://localhost:8080\n")
	fmt.Printf("Logs: %s\n", logFile)

	return CommandResult{shouldRunService: false}, nil
}

// handleDaemonStop stops the running daemon
func handleDaemonStop() (CommandResult, error) {
	pid, err := readPidFile()
	if err != nil {
		if os.IsNotExist(err) {
			return CommandResult{shouldRunService: false}, fmt.Errorf("service is not running")
		}
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to read PID file: %v", err)
	}

	if !isProcessRunning(pid) {
		removePidFile()
		return CommandResult{shouldRunService: false}, fmt.Errorf("service is not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to find process: %v", err)
	}

	// Send SIGTERM for graceful shutdown
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to stop service: %v", err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	// Check if still running, force kill if necessary
	if isProcessRunning(pid) {
		err = process.Kill()
		if err != nil {
			return CommandResult{shouldRunService: false}, fmt.Errorf("failed to force kill service: %v", err)
		}
		fmt.Println("Service force killed")
	} else {
		fmt.Println("Service stopped gracefully")
	}

	// Remove PID file
	removePidFile()

	return CommandResult{shouldRunService: false}, nil
}

// handleDaemonStatus shows the status of the daemon
func handleDaemonStatus() (CommandResult, error) {
	pid, err := readPidFile()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Service is not running")
			return CommandResult{shouldRunService: false}, nil
		}
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to read PID file: %v", err)
	}

	if isProcessRunning(pid) {
		fmt.Printf("Service is running (PID: %d)\n", pid)
		fmt.Println("Web interface: http://localhost:8080")
	} else {
		fmt.Println("Service is not running (stale PID file)")
		removePidFile()
	}

	return CommandResult{shouldRunService: false}, nil
}

// handleDaemonRestart restarts the daemon
func handleDaemonRestart(configPath string) (CommandResult, error) {
	// Stop if running
	_, err := handleDaemonStop()
	if err != nil && !strings.Contains(err.Error(), "not running") {
		return CommandResult{shouldRunService: false}, err
	}

	// Wait a moment
	time.Sleep(1 * time.Second)

	// Start again
	return handleDaemonStart(configPath)
}

// handleDaemonLogs shows the daemon service logs
func handleDaemonLogs() (CommandResult, error) {
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	logFile := filepath.Join(dir, "daemon.log")

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("No daemon logs found. Start the service first with: ./run-script-service daemon start")
		return CommandResult{shouldRunService: false}, nil
	}

	// Read and display the log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		return CommandResult{shouldRunService: false}, fmt.Errorf("failed to read daemon logs: %v", err)
	}

	if len(content) == 0 {
		fmt.Println("Daemon log file is empty")
	} else {
		fmt.Print(string(content))
	}

	return CommandResult{shouldRunService: false}, nil
}

// ensureFrontendBuilt checks if frontend needs building and builds it if necessary
func ensureFrontendBuilt(workDir string) error {
	frontendDir := filepath.Join(workDir, "web", "frontend")
	distDir := filepath.Join(frontendDir, "dist")

	// Check if frontend project exists
	if _, err := os.Stat(filepath.Join(frontendDir, "package.json")); os.IsNotExist(err) {
		return fmt.Errorf("frontend project not found at %s", frontendDir)
	}

	// Check if dist directory exists and has files
	if info, err := os.Stat(distDir); os.IsNotExist(err) || !info.IsDir() {
		fmt.Println("Frontend dist directory not found, building frontend...")
		return buildFrontend(frontendDir)
	}

	// Check if dist directory is empty
	files, err := os.ReadDir(distDir)
	if err != nil {
		return fmt.Errorf("failed to read dist directory: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("Frontend dist directory is empty, building frontend...")
		return buildFrontend(frontendDir)
	}

	// Check if package.json is newer than dist (basic staleness check)
	packageJsonPath := filepath.Join(frontendDir, "package.json")
	packageInfo, err := os.Stat(packageJsonPath)
	if err != nil {
		return fmt.Errorf("failed to stat package.json: %v", err)
	}

	distInfo, err := os.Stat(distDir)
	if err != nil {
		return fmt.Errorf("failed to stat dist directory: %v", err)
	}

	if packageInfo.ModTime().After(distInfo.ModTime()) {
		fmt.Println("Frontend source appears newer than build, rebuilding frontend...")
		return buildFrontend(frontendDir)
	}

	fmt.Println("Frontend build appears up to date")
	return nil
}

// buildFrontend builds the frontend using pnpm/vite
func buildFrontend(frontendDir string) error {
	fmt.Printf("Building frontend in %s...\n", frontendDir)

	// Check if node_modules exists, install dependencies if not
	nodeModulesDir := filepath.Join(frontendDir, "node_modules")
	if _, err := os.Stat(nodeModulesDir); os.IsNotExist(err) {
		fmt.Println("Installing frontend dependencies...")
		if err := runCommand("pnpm", []string{"install"}, frontendDir); err != nil {
			return fmt.Errorf("pnpm install failed: %v", err)
		}
	}

	// Run the build command
	fmt.Println("Running frontend build...")
	if err := runCommand("pnpm", []string{"build"}, frontendDir); err != nil {
		return fmt.Errorf("pnpm build failed: %v", err)
	}

	fmt.Println("Frontend build completed successfully")
	return nil
}

// runCommand executes a command in the specified directory
func runCommand(command string, args []string, workingDir string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// generateRandomKey generates a cryptographically secure random key
func generateRandomKey() string {
	bytes := make([]byte, 32) // 256-bit key
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to time-based key if crypto/rand fails
		return fmt.Sprintf("fallback-key-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
