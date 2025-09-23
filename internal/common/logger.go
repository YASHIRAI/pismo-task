package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger represents a structured logger
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       LogLevel
	logFile     *os.File
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string, logLevel LogLevel) (*Logger, error) {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("%s_%s.log", serviceName, timestamp)
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create loggers with different prefixes
	debugLogger := log.New(multiWriter, fmt.Sprintf("[%s][DEBUG] ", serviceName), log.LstdFlags|log.Lshortfile)
	infoLogger := log.New(multiWriter, fmt.Sprintf("[%s][INFO] ", serviceName), log.LstdFlags|log.Lshortfile)
	warnLogger := log.New(multiWriter, fmt.Sprintf("[%s][WARN] ", serviceName), log.LstdFlags|log.Lshortfile)
	errorLogger := log.New(multiWriter, fmt.Sprintf("[%s][ERROR] ", serviceName), log.LstdFlags|log.Lshortfile)
	fatalLogger := log.New(multiWriter, fmt.Sprintf("[%s][FATAL] ", serviceName), log.LstdFlags|log.Lshortfile)

	return &Logger{
		debugLogger: debugLogger,
		infoLogger:  infoLogger,
		warnLogger:  warnLogger,
		errorLogger: errorLogger,
		fatalLogger: fatalLogger,
		level:       logLevel,
		logFile:     logFile,
	}, nil
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debugLogger.Printf(format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.infoLogger.Printf(format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warnLogger.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.errorLogger.Printf(format, v...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.fatalLogger.Printf(format, v...)
	os.Exit(1)
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// LogRequest logs HTTP request details
func (l *Logger) LogRequest(method, path, clientIP string, statusCode int, duration time.Duration) {
	l.Info("HTTP %s %s from %s - Status: %d - Duration: %v", method, path, clientIP, statusCode, duration)
}

// LogDatabase logs database operations
func (l *Logger) LogDatabase(operation, table string, duration time.Duration, err error) {
	if err != nil {
		l.Error("DB %s on %s failed after %v: %v", operation, table, duration, err)
	} else {
		l.Debug("DB %s on %s completed in %v", operation, table, duration)
	}
}

// LogGRPC logs gRPC operations
func (l *Logger) LogGRPC(method string, duration time.Duration, err error) {
	if err != nil {
		l.Error("gRPC %s failed after %v: %v", method, duration, err)
	} else {
		l.Debug("gRPC %s completed in %v", method, duration)
	}
}

// LogBusinessOperation logs business logic operations
func (l *Logger) LogBusinessOperation(operation string, details map[string]interface{}, err error) {
	if err != nil {
		l.Error("Business operation %s failed: %v - Details: %+v", operation, err, details)
	} else {
		l.Info("Business operation %s completed successfully - Details: %+v", operation, details)
	}
}

// ParseLogLevel parses a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// Global logger instance
var GlobalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(serviceName string, logLevel LogLevel) error {
	var err error
	GlobalLogger, err = NewLogger(serviceName, logLevel)
	return err
}

// Convenience functions for global logger
func Debug(format string, v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Debug(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Info(format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Warn(format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Error(format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Fatal(format, v...)
	}
}
