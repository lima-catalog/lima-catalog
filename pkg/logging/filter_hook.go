// Package logging provides log filtering utilities for suppressing specific log messages.
//
// The package implements a FilterWriter that intercepts log output and suppresses lines
// containing specified keywords. This is useful for reducing noise from third-party
// libraries or filtering out experimental/deprecated feature warnings.
//
// Example usage:
//
//	// Filter out Lima's experimental feature warnings
//	logging.SetupLogrusFilter("EXPERIMENTAL", "DEPRECATED")
//
//	// Or create a custom filter writer
//	filtered := logging.NewFilterWriter(os.Stderr, "EXPERIMENTAL")
//	logger.SetOutput(filtered)
//
// The filtering is case-sensitive and works by checking if log lines contain
// any of the specified keywords. Matching lines are suppressed entirely.
package logging

import (
	"bytes"
	"io"

	"github.com/sirupsen/logrus"
)

// FilterWriter is an io.Writer that filters out log lines containing specific keywords
type FilterWriter struct {
	underlying io.Writer
	keywords   []string
}

// NewFilterWriter creates a new filter writer that suppresses log lines containing any of the specified keywords
func NewFilterWriter(underlying io.Writer, keywords ...string) *FilterWriter {
	return &FilterWriter{
		underlying: underlying,
		keywords:   keywords,
	}
}

// Write filters the input and writes only non-matching lines to the underlying writer
func (fw *FilterWriter) Write(p []byte) (n int, err error) {
	// Check if this log line contains any filtered keywords
	for _, keyword := range fw.keywords {
		if bytes.Contains(p, []byte(keyword)) {
			// Suppress this line by not writing it
			// Return the original length to indicate all bytes were "written"
			return len(p), nil
		}
	}

	// No keywords found, write to underlying writer
	return fw.underlying.Write(p)
}

// SetupLogrusFilter configures logrus to filter out messages containing the specified keywords
// This should be called early in main() before any logging occurs
func SetupLogrusFilter(keywords ...string) {
	// Get the logrus standard logger
	// Note: This is a singleton, so this affects all logrus usage in the process
	// including Lima's internal logging
	logrus.StandardLogger().SetOutput(NewFilterWriter(logrus.StandardLogger().Out, keywords...))
}
