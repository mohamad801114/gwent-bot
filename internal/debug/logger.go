package debug

import (
	"log"
	"os"
	"sync/atomic"
)

var (
	enabled int32
	prefix  string
)

// Init initializes the debug logger
func Init(p string) {
	prefix = p
	atomic.StoreInt32(&enabled, 1)
	log.Printf("[DEBUG:%s] Debug logger initialized", prefix)
}

// Enabled returns whether debug logging is enabled
func Enabled() bool {
	return atomic.LoadInt32(&enabled) == 1
}

// Enable enables debug logging
func Enable() {
	atomic.StoreInt32(&enabled, 1)
}

// Disable disables debug logging
func Disable() {
	atomic.StoreInt32(&enabled, 0)
}

// Logf logs a debug message if enabled
func Logf(format string, args ...any) {
	if atomic.LoadInt32(&enabled) == 1 {
		log.Printf("[DEBUG:%s] "+format, append([]any{prefix}, args...)...)
	}
}

// Errorf logs an error message (always enabled)
func Errorf(format string, args ...any) {
	log.Printf("[ERROR:%s] "+format, append([]any{prefix}, args...)...)
}

// Warnf logs a warning message (always enabled)
func Warnf(format string, args ...any) {
	log.Printf("[WARN:%s] "+format, append([]any{prefix}, args...)...)
}

// Infof logs an info message (always enabled)
func Infof(format string, args ...any) {
	log.Printf("[INFO:%s] "+format, append([]any{prefix}, args...)...)
}

// DebugContext holds debug context for a request
type DebugContext struct {
	RequestID string
	UserID    int64
	Command   string
}

// NewDebugContext creates a new debug context
func NewDebugContext(requestID string, userID int64, command string) *DebugContext {
	return &DebugContext{
		RequestID: requestID,
		UserID:    userID,
		Command:   command,
	}
}

// Logf logs with debug context
func (d *DebugContext) Logf(format string, args ...any) {
	if atomic.LoadInt32(&enabled) == 1 {
		log.Printf("[DEBUG:%s][req:%s][user:%d][cmd:%s] "+format, 
			append([]any{prefix, d.RequestID, d.UserID, d.Command}, args...)...)
	}
}