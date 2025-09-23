package common

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Test creating a new logger
	logger, err := NewLogger("test-service", INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test that log file was created
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Logs directory was not created")
	}

	// Test logging at different levels
	logger.Debug("This debug message should not appear")
	logger.Info("This info message should appear")
	logger.Warn("This warning message should appear")
	logger.Error("This error message should appear")

	// Test setting log level
	logger.SetLevel(DEBUG)
	logger.Debug("This debug message should now appear")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"WARN", WARN},
		{"ERROR", ERROR},
		{"FATAL", FATAL},
		{"INVALID", INFO}, // Default fallback
	}

	for _, test := range tests {
		result := ParseLogLevel(test.level)
		if result != test.expected {
			t.Errorf("ParseLogLevel(%s) = %v, expected %v", test.level, result, test.expected)
		}
	}
}

func TestLoggerMethods(t *testing.T) {
	logger, err := NewLogger("test-service", DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test LogRequest
	logger.LogRequest("GET", "/test", "127.0.0.1", 200, 100*time.Millisecond)

	// Test LogDatabase
	logger.LogDatabase("SELECT", "accounts", 50*time.Millisecond, nil)
	logger.LogDatabase("INSERT", "transactions", 75*time.Millisecond, os.ErrNotExist)

	// Test LogGRPC
	logger.LogGRPC("CreateAccount", 200*time.Millisecond, nil)
	logger.LogGRPC("CreateTransaction", 150*time.Millisecond, os.ErrPermission)

	// Test LogBusinessOperation
	details := map[string]interface{}{
		"account_id": "test-123",
		"amount":     100.50,
	}
	logger.LogBusinessOperation("CreateAccount", details, nil)
	logger.LogBusinessOperation("CreateTransaction", details, os.ErrInvalid)
}

func TestGlobalLogger(t *testing.T) {
	// Test global logger initialization
	err := InitGlobalLogger("global-test", INFO)
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}
	defer GlobalLogger.Close()

	// Test global logger functions
	Debug("Global debug message")
	Info("Global info message")
	Warn("Global warning message")
	Error("Global error message")
}

func TestLogFileCreation(t *testing.T) {
	// Clean up any existing logs directory
	os.RemoveAll("logs")

	logger, err := NewLogger("file-test", INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Check that log file exists
	logFiles, err := filepath.Glob("logs/file-test_*.log")
	if err != nil {
		t.Fatalf("Failed to glob log files: %v", err)
	}

	if len(logFiles) == 0 {
		t.Errorf("No log file was created")
	}

	// Write a test message
	logger.Info("Test message for file creation")

	// Check that the file has content
	logFile := logFiles[0]
	fileInfo, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Errorf("Log file is empty")
	}
}

func TestConcurrentLogging(t *testing.T) {
	logger, err := NewLogger("concurrent-test", INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				logger.Info("Concurrent log message %d-%d", id, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
