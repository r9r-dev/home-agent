package services

import (
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID        int64     `json:"id"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// LogSubscriber is a channel that receives log entries
type LogSubscriber chan LogEntry

// LogService provides in-memory logging with real-time streaming
type LogService struct {
	mu          sync.RWMutex
	entries     []LogEntry
	maxEntries  int
	nextID      int64
	subscribers map[LogSubscriber]struct{}
	subMu       sync.RWMutex

	// Track highest level for indicator
	hasWarning  bool
	hasError    bool
}

// NewLogService creates a new LogService with a maximum buffer size
func NewLogService(maxEntries int) *LogService {
	if maxEntries <= 0 {
		maxEntries = 100
	}
	return &LogService{
		entries:     make([]LogEntry, 0, maxEntries),
		maxEntries:  maxEntries,
		subscribers: make(map[LogSubscriber]struct{}),
	}
}

// Log adds a new log entry and notifies subscribers
func (ls *LogService) Log(level LogLevel, message string) {
	ls.mu.Lock()

	entry := LogEntry{
		ID:        ls.nextID,
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}
	ls.nextID++

	// Add to buffer (circular)
	if len(ls.entries) >= ls.maxEntries {
		ls.entries = ls.entries[1:]
	}
	ls.entries = append(ls.entries, entry)

	// Track levels for indicator
	if level == LogLevelWarning {
		ls.hasWarning = true
	} else if level == LogLevelError {
		ls.hasError = true
	}

	ls.mu.Unlock()

	// Notify subscribers (non-blocking)
	ls.subMu.RLock()
	for sub := range ls.subscribers {
		select {
		case sub <- entry:
		default:
			// Skip if subscriber is slow
		}
	}
	ls.subMu.RUnlock()
}

// Info logs an info message
func (ls *LogService) Info(message string) {
	ls.Log(LogLevelInfo, message)
}

// Warning logs a warning message
func (ls *LogService) Warning(message string) {
	ls.Log(LogLevelWarning, message)
}

// Error logs an error message
func (ls *LogService) Error(message string) {
	ls.Log(LogLevelError, message)
}

// GetEntries returns all current log entries
func (ls *LogService) GetEntries() []LogEntry {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	// Return a copy
	entries := make([]LogEntry, len(ls.entries))
	copy(entries, ls.entries)
	return entries
}

// GetStatus returns the current log status (highest level seen)
func (ls *LogService) GetStatus() LogLevel {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if ls.hasError {
		return LogLevelError
	}
	if ls.hasWarning {
		return LogLevelWarning
	}
	return LogLevelInfo
}

// ClearStatus resets the warning/error indicators
func (ls *LogService) ClearStatus() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.hasWarning = false
	ls.hasError = false
}

// Subscribe adds a subscriber to receive log entries
func (ls *LogService) Subscribe() LogSubscriber {
	sub := make(LogSubscriber, 10)

	ls.subMu.Lock()
	ls.subscribers[sub] = struct{}{}
	ls.subMu.Unlock()

	return sub
}

// Unsubscribe removes a subscriber
func (ls *LogService) Unsubscribe(sub LogSubscriber) {
	ls.subMu.Lock()
	delete(ls.subscribers, sub)
	ls.subMu.Unlock()

	close(sub)
}
