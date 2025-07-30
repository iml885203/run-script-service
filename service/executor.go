// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
func (e *Executor) ExecuteScript() *ExecutionResult {
	timestamp := time.Now()
	result := &ExecutionResult{
		Timestamp: timestamp,
	}

	cmd := exec.Command(e.scriptPath)
	cmd.Dir = filepath.Dir(e.scriptPath)

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

	// Write to log
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

	return result
}

// logError logs an error message
func (e *Executor) logError(timestamp time.Time, message string) {
	errorMsg := fmt.Sprintf("[%s] ERROR: %s\n%s\n",
		timestamp.Format("2006-01-02 15:04:05"), message, strings.Repeat("-", 50))
	if err := e.WriteLog(errorMsg); err != nil {
		fmt.Printf("Error writing error to log: %v\n", err)
	}
	fmt.Printf("Error executing script: %s\n", message)
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
