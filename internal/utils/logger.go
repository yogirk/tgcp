package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

// InitLogger initializes the logger to write to ~/.tgcp/debug.log
func InitLogger() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(home, ".tgcp")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(logDir, "debug.log")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logFile = f
	
	// Write header
	fmt.Fprintf(logFile, "\n--- Log session started at %s ---\n", time.Now().Format(time.RFC3339))
	return nil
}

// Log writes a formatted message to the log file
func Log(format string, args ...interface{}) {
	if logFile != nil {
		timestamp := time.Now().Format("15:04:05.000")
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(logFile, "[%s] %s\n", timestamp, msg)
	}
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
