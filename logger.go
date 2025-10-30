package main

import (
	"fmt"
	"sync"
	"time"
)

// LogEntry represents a single log message with timestamp
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// Logger manages application logs in memory
type Logger struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
}

// NewLogger creates a new logger instance
func NewLogger(maxSize int) *Logger {
	return &Logger{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Log adds a new log entry
func (l *Logger) Log(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format("15:04:05.000"),
		Message:   message,
	}

	// If we've reached max size, remove oldest entry
	if len(l.entries) >= l.maxSize {
		l.entries = l.entries[1:]
	}

	l.entries = append(l.entries, entry)
}

// GetLogs returns all current log entries
func (l *Logger) GetLogs() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Return a copy to avoid race conditions
	logs := make([]LogEntry, len(l.entries))
	copy(logs, l.entries)
	return logs
}

// Clear removes all log entries
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = make([]LogEntry, 0, l.maxSize)
}

// LogPrintf is a wrapper around fmt.Printf that also logs to the logger
func (l *Logger) LogPrintf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.Log(message)
	fmt.Print(message)
}

// LogPrintln is a wrapper around fmt.Println that also logs to the logger
func (l *Logger) LogPrintln(args ...interface{}) {
	message := fmt.Sprintln(args...)
	l.Log(message)
	fmt.Print(message)
}
