package logging

import (
	"github.com/sirupsen/logrus"
)

// WarningCaptureHook is a logrus hook that captures warning-level log entries
// This is useful for tracking validation warnings during template parsing
type WarningCaptureHook struct {
	Warnings []string
}

// NewWarningCaptureHook creates a new warning capture hook
func NewWarningCaptureHook() *WarningCaptureHook {
	return &WarningCaptureHook{Warnings: []string{}}
}

// Levels returns the log levels this hook should fire for
func (h *WarningCaptureHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel}
}

// Fire is called when a log entry is made at the warning level
func (h *WarningCaptureHook) Fire(entry *logrus.Entry) error {
	// Capture the warning message
	h.Warnings = append(h.Warnings, entry.Message)
	return nil
}

// Reset resets the warnings list
func (h *WarningCaptureHook) Reset() {
	h.Warnings = []string{}
}

// Count returns the number of warnings captured
func (h *WarningCaptureHook) Count() int {
	return len(h.Warnings)
}
