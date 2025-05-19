package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
)

// Logger handles all logging operations
type Logger struct {
	logger    *log.Logger
	logWriter io.WriteCloser
	verbose   bool
}

// New creates a new Logger
func New(logFile string, maxSize, maxBackups, maxAge int, verbose bool) *Logger {
	// Create directory for log if it doesn't exist
	logDir := filepath.Dir(logFile)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	// Configure log rotation
	var writer io.Writer
	if logFile == "stdout" {
		writer = os.Stdout
	} else {
		writer = &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    maxSize,    // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,     // days
			Compress:   true,
		}
	}

	// If verbose, log to stdout as well
	if verbose && logFile != "stdout" {
		writer = io.MultiWriter(writer, os.Stdout)
	}

	return &Logger{
		logger:    log.New(writer, "", log.LstdFlags),
		logWriter: writer.(io.WriteCloser),
		verbose:   verbose,
	}
}

// Info logs informational messages
func (l *Logger) Info(format string, v ...interface{}) {
	l.logger.Printf("[INFO] "+format, v...)
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Printf("[ERROR] "+format, v...)
	if l.verbose {
		fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", v...)
	}
}

// Debug logs debug messages (only in verbose mode)
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.verbose {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Close closes the log writer
func (l *Logger) Close() {
	if closer, ok := l.logWriter.(io.Closer); ok {
		closer.Close()
	}
}