// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ExecutionResult contains the results of script execution
type ExecutionResult struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	Timestamp time.Time
}

// Executor handles script execution and logging
type Executor struct {
	scriptPath string
	logPath    string
	maxLines   int
	logHandler LogHandler
}

// NewExecutor creates a new script executor
func NewExecutor(scriptPath, logPath string, maxLines int) *Executor {
	return &Executor{
		scriptPath: scriptPath,
		logPath:    logPath,
		maxLines:   maxLines,
	}
}

// ExecuteScript executes the configured script and logs the results
func (e *Executor) ExecuteScript(args ...string) *ExecutionResult {
	// Use context with timeout for backward compatibility
	ctx := context.Background()
	return e.ExecuteScriptWithContext(ctx, args...)
}

// ExecuteScriptWithContext executes the configured script with context support
func (e *Executor) ExecuteScriptWithContext(ctx context.Context, args ...string) *ExecutionResult {
	timestamp := time.Now()
	result := &ExecutionResult{
		Timestamp: timestamp,
	}

	cmd := exec.CommandContext(ctx, e.scriptPath, args...)
	cmd.Dir = filepath.Dir(e.scriptPath)

	// Set process group to enable proper cleanup
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		e.logError(timestamp, fmt.Sprintf("Error creating stdout pipe: %v", err))
		result.ExitCode = -1
		return result
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		e.logError(timestamp, fmt.Sprintf("Error creating stderr pipe: %v", err))
		result.ExitCode = -1
		return result
	}

	if startErr := cmd.Start(); startErr != nil {
		e.logError(timestamp, fmt.Sprintf("Error starting command: %v", startErr))
		result.ExitCode = -1
		return result
	}

	// Ensure process cleanup on exit
	defer func() {
		if cmd.Process != nil {
			// Kill the entire process group to clean up any child processes
			if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
				// Only kill if the process is still running and we can get the pgid
				_ = syscall.Kill(-pgid, syscall.SIGTERM)

				// Wait a moment for graceful shutdown, then force kill if needed
				go func() {
					time.Sleep(100 * time.Millisecond)
					if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
						_ = syscall.Kill(-pgid, syscall.SIGKILL)
					}
				}()
			}
		}
	}()

	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)

	err = cmd.Wait()
	result.ExitCode = 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			e.logError(timestamp, fmt.Sprintf("Error waiting for command: %v", err))
			result.ExitCode = -1
			return result
		}
	}

	result.Stdout = strings.TrimSpace(string(stdoutBytes))
	result.Stderr = strings.TrimSpace(string(stderrBytes))

	// Write execution result to log
	e.writeExecutionLog(timestamp, result)

	return result
}

// logError logs an error message
func (e *Executor) logError(timestamp time.Time, message string) {
	if e.logPath != "" {
		errorMsg := fmt.Sprintf("[%s] ERROR: %s\n%s\n",
			timestamp.Format("2006-01-02 15:04:05"), message, strings.Repeat("-", 50))
		if err := e.WriteLog(errorMsg); err != nil {
			fmt.Printf("Error writing error to log: %v\n", err)
		}
	}
	fmt.Printf("Error executing script: %s\n", message)
}

// writeExecutionLog writes execution result to log file
func (e *Executor) writeExecutionLog(timestamp time.Time, result *ExecutionResult) {
	if e.logPath != "" {
		logEntry := fmt.Sprintf("[%s] Exit code: %d\n", timestamp.Format("2006-01-02 15:04:05"), result.ExitCode)
		if result.Stdout != "" {
			logEntry += fmt.Sprintf("STDOUT: %s\n", result.Stdout)
		}
		if result.Stderr != "" {
			logEntry += fmt.Sprintf("STDERR: %s\n", result.Stderr)
		}
		logEntry += strings.Repeat("-", 50) + "\n"

		if err := e.WriteLog(logEntry); err != nil {
			fmt.Printf("Error writing to log: %v\n", err)
		}

		if err := e.TrimLog(); err != nil {
			fmt.Printf("Error trimming log: %v\n", err)
		}
	}
}

// WriteLog writes content to the log file
func (e *Executor) WriteLog(content string) error {
	file, err := os.OpenFile(e.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// TrimLog keeps only the last maxLines lines in the log file
func (e *Executor) TrimLog() error {
	file, err := os.Open(e.logPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return scanErr
	}

	if len(lines) <= e.maxLines {
		return nil
	}

	// Keep only the last maxLines lines
	linesToKeep := lines[len(lines)-e.maxLines:]

	outFile, err := os.Create(e.logPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, line := range linesToKeep {
		if _, err := outFile.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteWithStreaming executes the script with streaming output
func (e *Executor) ExecuteWithStreaming(ctx context.Context, args ...string) *ExecutionResult {
	timestamp := time.Now()
	result := &ExecutionResult{
		Timestamp: timestamp,
	}

	// Notify handler of execution start
	if e.logHandler != nil {
		e.logHandler.HandleExecutionStart(timestamp)
	}

	cmd := exec.CommandContext(ctx, e.scriptPath, args...)
	cmd.Dir = filepath.Dir(e.scriptPath)

	// Set process group to enable proper cleanup
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		e.logError(timestamp, fmt.Sprintf("Error creating stdout pipe: %v", err))
		result.ExitCode = -1
		if e.logHandler != nil {
			e.logHandler.HandleExecutionEnd(timestamp, -1)
		}
		return result
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		e.logError(timestamp, fmt.Sprintf("Error creating stderr pipe: %v", err))
		result.ExitCode = -1
		if e.logHandler != nil {
			e.logHandler.HandleExecutionEnd(timestamp, -1)
		}
		return result
	}

	if startErr := cmd.Start(); startErr != nil {
		e.logError(timestamp, fmt.Sprintf("Error starting command: %v", startErr))
		result.ExitCode = -1
		if e.logHandler != nil {
			e.logHandler.HandleExecutionEnd(timestamp, -1)
		}
		return result
	}

	// Ensure process cleanup on exit
	defer func() {
		if cmd.Process != nil {
			// Kill the entire process group to clean up any child processes
			if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
				// Only kill if the process is still running and we can get the pgid
				_ = syscall.Kill(-pgid, syscall.SIGTERM)

				// Wait a moment for graceful shutdown, then force kill if needed
				go func() {
					time.Sleep(100 * time.Millisecond)
					if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
						_ = syscall.Kill(-pgid, syscall.SIGKILL)
					}
				}()
			}
		}
	}()

	// Stream output in real-time using goroutines
	var stdoutBuilder strings.Builder
	var stderrBuilder strings.Builder
	var wg sync.WaitGroup

	// Start streaming goroutines
	wg.Add(2)
	go func() {
		defer wg.Done()
		e.streamOutputToBuilder(stdout, "STDOUT", &stdoutBuilder)
	}()

	go func() {
		defer wg.Done()
		e.streamOutputToBuilder(stderr, "STDERR", &stderrBuilder)
	}()

	err = cmd.Wait()
	// Wait for all streaming to complete before proceeding
	wg.Wait()
	result.ExitCode = 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			e.logError(timestamp, fmt.Sprintf("Error waiting for command: %v", err))
			result.ExitCode = -1
		}
	}

	result.Stdout = strings.TrimSpace(stdoutBuilder.String())
	result.Stderr = strings.TrimSpace(stderrBuilder.String())

	// Notify handler of execution end
	if e.logHandler != nil {
		e.logHandler.HandleExecutionEnd(time.Now(), result.ExitCode)
	}

	// Write execution result to log
	e.writeExecutionLog(timestamp, result)

	return result
}

// streamOutputToBuilder processes output from a reader line by line, calls the log handler, and builds output string
func (e *Executor) streamOutputToBuilder(reader io.Reader, streamType string, builder *strings.Builder) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		timestamp := time.Now()

		// Add to builder for backward compatibility
		builder.WriteString(line + "\n")

		// Send to log handler if available
		if e.logHandler != nil {
			e.logHandler.HandleLogLine(timestamp, streamType, line)
		}
	}
}

// SetLogHandler sets the log handler for streaming output
func (e *Executor) SetLogHandler(handler LogHandler) {
	e.logHandler = handler
}

// ExecuteWithResult executes the script and returns both result and error
// This method provides a simpler interface for script execution with error handling
func (e *Executor) ExecuteWithResult(ctx context.Context, args ...string) (*ExecutionResult, error) {
	ctx = e.ensureContext(ctx)
	result := e.ExecuteScriptWithContext(ctx, args...)
	return e.handleExecutionResult(result)
}

// ExecuteWithResultStreaming executes the script with streaming output and returns both result and error
// This method combines streaming capabilities with error handling interface, allowing real-time
// output processing while maintaining the same error handling semantics as ExecuteWithResult
func (e *Executor) ExecuteWithResultStreaming(ctx context.Context, args ...string) (*ExecutionResult, error) {
	ctx = e.ensureContext(ctx)
	result := e.ExecuteWithStreaming(ctx, args...)
	return e.handleExecutionResult(result)
}

// ensureContext ensures we have a valid context, using background context as fallback
func (e *Executor) ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// handleExecutionResult converts execution results to the result/error pattern
// Non-zero exit codes are converted to errors while preserving the full result
func (e *Executor) handleExecutionResult(result *ExecutionResult) (*ExecutionResult, error) {
	if result.ExitCode != 0 {
		return result, fmt.Errorf("script execution failed with exit code %d", result.ExitCode)
	}
	return result, nil
}
