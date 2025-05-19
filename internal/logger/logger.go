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
	logWriter io.Writer // 修改：从 io.WriteCloser 改为 io.Writer
	verbose   bool
	closers   []io.Closer // 新增：保存需要关闭的写入器列表
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

	var writers []io.Writer
	var closers []io.Closer

	// Configure log rotation
	if logFile != "stdout" {
		logRotator := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge, // days
			Compress:   true,
		}
		writers = append(writers, logRotator)
		closers = append(closers, logRotator)
	}

	// Add stdout if verbose or if stdout is explicitly requested
	if verbose || logFile == "stdout" {
		writers = append(writers, os.Stdout)
	}

	// Create multi-writer if needed
	var writer io.Writer
	if len(writers) > 1 {
		writer = io.MultiWriter(writers...)
	} else if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = os.Stdout // 默认为标准输出
	}

	return &Logger{
		logger:    log.New(writer, "", log.LstdFlags),
		logWriter: writer,
		verbose:   verbose,
		closers:   closers,
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

// Close closes all registered closers
func (l *Logger) Close() {
	for _, closer := range l.closers {
		closer.Close()
	}
}
