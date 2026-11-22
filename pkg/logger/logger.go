package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents log severity level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger is a simple leveled logger.
type Logger struct {
	mu     sync.Mutex
	level  Level
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	SetDefault(New(INFO, os.Stdout))
}

// New creates a new Logger.
func New(level Level, out io.Writer) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(out, "", 0),
	}
}

// SetDefault sets the global default logger.
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Default returns the global default logger.
func Default() *Logger {
	return defaultLogger
}

// SetLevel changes the log level of the default logger.
func SetLevel(level Level) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.level = level
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	l.logger.Printf("[%s] [%s] %s", timestamp, levelNames[level], msg)

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs at DEBUG level.
func Debug(format string, args ...interface{}) {
	defaultLogger.log(DEBUG, format, args...)
}

// Info logs at INFO level.
func Info(format string, args ...interface{}) {
	defaultLogger.log(INFO, format, args...)
}

// Warn logs at WARN level.
func Warn(format string, args ...interface{}) {
	defaultLogger.log(WARN, format, args...)
}

// Error logs at ERROR level.
func Error(format string, args ...interface{}) {
	defaultLogger.log(ERROR, format, args...)
}

// Fatal logs at FATAL level and exits.
func Fatal(format string, args ...interface{}) {
	defaultLogger.log(FATAL, format, args...)
}

// Debugf logs at DEBUG level.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Infof logs at INFO level.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warnf logs at WARN level.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Errorf logs at ERROR level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatalf logs at FATAL level and exits.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}
