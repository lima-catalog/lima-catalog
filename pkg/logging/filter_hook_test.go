// Unit Tests: Log Filtering (logging package)
//
// High-level overview of what's being tested:
// - Filtering log messages by keyword (e.g., "EXPERIMENTAL")
// - Supporting multiple filter keywords simultaneously
// - Passing through non-filtered messages
// - Case-sensitive keyword matching
// - Filter working across all log levels (Trace, Debug, Info, Warn, Error)
// - Integration with logrus logger
// - Preserving normal log output while filtering specific keywords

package logging

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestFilterWriter_FiltersExperimentalMessages(t *testing.T) {
	// Create a logger with filtered output
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(NewFilterWriter(&buf, "EXPERIMENTAL"))
	logger.SetLevel(logrus.InfoLevel)

	// Test that EXPERIMENTAL messages are filtered out
	logger.Info("This is an EXPERIMENTAL feature")
	logger.Warn("Warning: EXPERIMENTAL functionality")

	// Test that normal messages pass through
	logger.Info("This is a normal info message")
	logger.Warn("This is a normal warning")

	output := buf.String()

	// Check that EXPERIMENTAL messages were filtered
	if bytes.Contains([]byte(output), []byte("EXPERIMENTAL")) {
		t.Errorf("Expected EXPERIMENTAL messages to be filtered, but found in output: %s", output)
	}

	// Check that normal messages were not filtered
	if !bytes.Contains([]byte(output), []byte("normal info message")) {
		t.Errorf("Expected normal info message to be present, but not found in output: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("normal warning")) {
		t.Errorf("Expected normal warning to be present, but not found in output: %s", output)
	}
}

func TestFilterWriter_MultipleKeywords(t *testing.T) {
	// Create a logger with multiple filter keywords
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(NewFilterWriter(&buf, "EXPERIMENTAL", "DEPRECATED"))
	logger.SetLevel(logrus.InfoLevel)

	// Test filtering of both keywords
	logger.Info("This is an EXPERIMENTAL feature")
	logger.Warn("This feature is DEPRECATED")
	logger.Info("This is a normal message")

	output := buf.String()

	// Check that both filtered keywords are absent
	if bytes.Contains([]byte(output), []byte("EXPERIMENTAL")) {
		t.Errorf("Expected EXPERIMENTAL messages to be filtered, but found in output: %s", output)
	}
	if bytes.Contains([]byte(output), []byte("DEPRECATED")) {
		t.Errorf("Expected DEPRECATED messages to be filtered, but found in output: %s", output)
	}

	// Check that normal message passed through
	if !bytes.Contains([]byte(output), []byte("normal message")) {
		t.Errorf("Expected normal message to be present, but not found in output: %s", output)
	}
}

func TestFilterWriter_CaseSensitive(t *testing.T) {
	// Create a logger that filters "EXPERIMENTAL" (all caps)
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(NewFilterWriter(&buf, "EXPERIMENTAL"))
	logger.SetLevel(logrus.InfoLevel)

	// Test that lowercase doesn't get filtered
	logger.Info("This is an experimental feature")
	logger.Info("This is an EXPERIMENTAL feature")

	output := buf.String()

	// Check that lowercase passes through
	if !bytes.Contains([]byte(output), []byte("experimental feature")) {
		t.Errorf("Expected lowercase 'experimental' to pass through, but not found in output: %s", output)
	}

	// Check that uppercase is filtered
	if bytes.Contains([]byte(output), []byte("EXPERIMENTAL")) {
		t.Errorf("Expected uppercase EXPERIMENTAL to be filtered, but found in output: %s", output)
	}
}

func TestFilterWriter_AllLevels(t *testing.T) {
	// Verify filter works on all log levels
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(NewFilterWriter(&buf, "EXPERIMENTAL"))
	logger.SetLevel(logrus.TraceLevel)

	// Test all levels
	logger.Trace("EXPERIMENTAL trace")
	logger.Debug("EXPERIMENTAL debug")
	logger.Info("EXPERIMENTAL info")
	logger.Warn("EXPERIMENTAL warn")
	logger.Error("EXPERIMENTAL error")

	// Also add some normal messages
	logger.Info("Normal info")
	logger.Warn("Normal warn")

	output := buf.String()

	// Check that no EXPERIMENTAL messages appear
	if bytes.Contains([]byte(output), []byte("EXPERIMENTAL")) {
		t.Errorf("Expected all EXPERIMENTAL messages to be filtered regardless of level, but found in output: %s", output)
	}

	// Check that normal messages pass through
	if !bytes.Contains([]byte(output), []byte("Normal info")) {
		t.Errorf("Expected normal messages to pass through, but not found in output: %s", output)
	}
}
