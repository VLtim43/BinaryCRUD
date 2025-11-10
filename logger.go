package main

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// LogEntry represents a single log message with timestamp and level
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// InMemoryHandler captures logs in memory for UI display
type InMemoryHandler struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	next    slog.Handler // Chain to file handler
}

// NewInMemoryHandler creates a new in-memory handler
func NewInMemoryHandler(maxSize int, next slog.Handler) *InMemoryHandler {
	return &InMemoryHandler{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
		next:    next,
	}
}

// Handle implements slog.Handler interface
func (h *InMemoryHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Store in memory
	entry := LogEntry{
		Timestamp: r.Time.Format("15:04:05.000"),
		Level:     r.Level.String(),
		Message:   r.Message,
	}

	// If we've reached max size, remove oldest entry
	if len(h.entries) >= h.maxSize {
		h.entries = h.entries[1:]
	}

	h.entries = append(h.entries, entry)

	// Chain to next handler (file) if set
	if h.next != nil {
		return h.next.Handle(ctx, r)
	}
	return nil
}

// Enabled implements slog.Handler interface
func (h *InMemoryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true // Accept all log levels
}

// WithAttrs implements slog.Handler interface
func (h *InMemoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.next != nil {
		return &InMemoryHandler{
			entries: h.entries,
			maxSize: h.maxSize,
			next:    h.next.WithAttrs(attrs),
		}
	}
	return h
}

// WithGroup implements slog.Handler interface
func (h *InMemoryHandler) WithGroup(name string) slog.Handler {
	if h.next != nil {
		return &InMemoryHandler{
			entries: h.entries,
			maxSize: h.maxSize,
			next:    h.next.WithGroup(name),
		}
	}
	return h
}

// GetLogs returns all current log entries
func (h *InMemoryHandler) GetLogs() []LogEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid race conditions
	logs := make([]LogEntry, len(h.entries))
	copy(logs, h.entries)
	return logs
}

// Clear removes all log entries
func (h *InMemoryHandler) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]LogEntry, 0, h.maxSize)
}

// Logger wraps slog.Logger with in-memory handler
type Logger struct {
	logger  *slog.Logger
	handler *InMemoryHandler
}

// NewLogger creates a new logger with in-memory and file handlers
func NewLogger(maxSize int) *Logger {
	// Create logs directory if it doesn't exist
	os.MkdirAll("logs", 0755)

	// Create file handler for persistent logs
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to stderr if file creation fails
		logFile = os.Stderr
	}

	// Create text handler for file output
	fileHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Create in-memory handler chained with file handler
	memHandler := NewInMemoryHandler(maxSize, fileHandler)

	// Create logger
	logger := slog.New(memHandler)

	return &Logger{
		logger:  logger,
		handler: memHandler,
	}
}

// Info logs at INFO level
func (l *Logger) Info(message string) {
	l.logger.Info(message)
}

// Debug logs at DEBUG level
func (l *Logger) Debug(message string) {
	l.logger.Debug(message)
}

// Warn logs at WARN level
func (l *Logger) Warn(message string) {
	l.logger.Warn(message)
}

// Error logs at ERROR level
func (l *Logger) Error(message string) {
	l.logger.Error(message)
}

// Log logs at INFO level (for backward compatibility)
func (l *Logger) Log(message string) {
	l.logger.Info(message)
}

// GetLogs returns all current log entries
func (l *Logger) GetLogs() []LogEntry {
	return l.handler.GetLogs()
}

// Clear removes all log entries
func (l *Logger) Clear() {
	l.handler.Clear()
}
