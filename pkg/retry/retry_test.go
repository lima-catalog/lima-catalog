package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithExponentialBackoff_Success(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := WithExponentialBackoff(fn, 3)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestWithExponentialBackoff_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := WithExponentialBackoff(fn, 3)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestWithExponentialBackoff_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	testErr := errors.New("persistent error")
	fn := func() error {
		callCount++
		return testErr
	}

	err := WithExponentialBackoff(fn, 2)
	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount != 3 { // Initial attempt + 2 retries = 3 total
		t.Errorf("expected 3 calls (initial + 2 retries), got %d", callCount)
	}

	if !errors.Is(err, testErr) {
		t.Errorf("expected error to wrap %v, got %v", testErr, err)
	}
}

func TestWithExponentialBackoffContext_Cancellation(t *testing.T) {
	callCount := 0
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() error {
		callCount++
		if callCount == 2 {
			cancel() // Cancel after second attempt
		}
		return errors.New("error")
	}

	config := DefaultConfig()
	config.MaxRetries = 5
	config.InitialDelay = 100 * time.Millisecond

	err := WithConfigContext(ctx, fn, config)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Should be called twice before cancellation
	if callCount != 2 {
		t.Errorf("expected 2 calls before cancellation, got %d", callCount)
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestDo_Success(t *testing.T) {
	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "success", nil
	}

	result, err := Do(fn, 3)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got %q", result)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	fn := func() (int, error) {
		callCount++
		if callCount < 3 {
			return 0, errors.New("temporary error")
		}
		return 42, nil
	}

	result, err := Do(fn, 3)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestDo_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	testErr := errors.New("persistent error")
	fn := func() (bool, error) {
		callCount++
		return false, testErr
	}

	result, err := Do(fn, 2)
	if err == nil {
		t.Error("expected error, got nil")
	}

	if result != false {
		t.Errorf("expected false (zero value), got %v", result)
	}

	if callCount != 3 { // Initial attempt + 2 retries
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestWithConfig_CustomBackoff(t *testing.T) {
	callCount := 0
	attempts := []time.Time{}

	fn := func() error {
		callCount++
		attempts = append(attempts, time.Now())
		if callCount < 3 {
			return errors.New("error")
		}
		return nil
	}

	config := &Config{
		MaxRetries:   3,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   2.0,
		ShouldRetry:  IsRetryable,
	}

	err := WithConfig(fn, config)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}

	// Verify exponential backoff delays
	if len(attempts) >= 2 {
		delay1 := attempts[1].Sub(attempts[0])
		if delay1 < 50*time.Millisecond || delay1 > 100*time.Millisecond {
			t.Errorf("first delay should be ~50ms, got %v", delay1)
		}
	}

	if len(attempts) >= 3 {
		delay2 := attempts[2].Sub(attempts[1])
		if delay2 < 100*time.Millisecond || delay2 > 150*time.Millisecond {
			t.Errorf("second delay should be ~100ms, got %v", delay2)
		}
	}
}

func TestWithConfig_NonRetryableError(t *testing.T) {
	callCount := 0
	nonRetryableErr := errors.New("non-retryable error")

	fn := func() error {
		callCount++
		return nonRetryableErr
	}

	config := &Config{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		ShouldRetry: func(err error) bool {
			return false // Never retry
		},
	}

	err := WithConfig(fn, config)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Should only be called once since error is not retryable
	if callCount != 1 {
		t.Errorf("expected 1 call for non-retryable error, got %d", callCount)
	}

	if !errors.Is(err, nonRetryableErr) {
		t.Errorf("expected error to wrap %v, got %v", nonRetryableErr, err)
	}
}

func TestCalculateDelay(t *testing.T) {
	config := &Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{
			name:    "Attempt 0",
			attempt: 0,
			want:    100 * time.Millisecond,
		},
		{
			name:    "Attempt 1",
			attempt: 1,
			want:    200 * time.Millisecond,
		},
		{
			name:    "Attempt 2",
			attempt: 2,
			want:    400 * time.Millisecond,
		},
		{
			name:    "Attempt 3",
			attempt: 3,
			want:    800 * time.Millisecond,
		},
		{
			name:    "Attempt 4 (capped at MaxDelay)",
			attempt: 4,
			want:    1 * time.Second,
		},
		{
			name:    "Attempt 5 (still capped)",
			attempt: 5,
			want:    1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDelay(tt.attempt, config)
			if got != tt.want {
				t.Errorf("CalculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries = 3, got %d", config.MaxRetries)
	}

	if config.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay = 1s, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay = 30s, got %v", config.MaxDelay)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier = 2.0, got %f", config.Multiplier)
	}

	if config.ShouldRetry == nil {
		t.Error("expected ShouldRetry function, got nil")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Nil error",
			err:  nil,
			want: false,
		},
		{
			name: "Any error",
			err:  errors.New("some error"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
