// Package retry provides exponential backoff retry logic with context cancellation support.
//
// The package implements flexible retry mechanisms for operations that may fail transiently,
// such as network requests or API calls. It supports:
//
//   - Exponential backoff with configurable delays and multipliers
//   - Context cancellation for early termination
//   - Custom retry policies via ShouldRetry functions
//   - Generic return values with Do[T] functions
//
// Example usage:
//
//	// Simple retry with default config
//	err := retry.WithExponentialBackoff(func() error {
//	    return makeAPICall()
//	}, 3)
//
//	// Generic function with return value
//	result, err := retry.Do(func() (string, error) {
//	    return fetchData()
//	}, 3)
//
//	// Custom configuration
//	config := &retry.Config{
//	    MaxRetries:   5,
//	    InitialDelay: 100 * time.Millisecond,
//	    MaxDelay:     10 * time.Second,
//	    Multiplier:   2.0,
//	    ShouldRetry:  retry.IsRetryable,
//	}
//	err = retry.WithConfig(fn, config)
package retry

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Config holds retry configuration
type Config struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the initial delay before first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the backoff multiplier (typically 2.0 for exponential)
	Multiplier float64
	// ShouldRetry determines if an error is retryable
	ShouldRetry func(error) bool
}

// DefaultConfig returns sensible defaults for retry logic
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		ShouldRetry:  IsRetryable,
	}
}

// IsRetryable returns true for errors that should be retried
// Currently retries all errors - can be customized for specific error types
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	// By default, retry all errors
	// This can be extended to check for specific error types
	// (e.g., network errors, rate limits, timeouts)
	return true
}

// WithExponentialBackoff retries a function with exponential backoff
func WithExponentialBackoff(fn func() error, maxRetries int) error {
	config := DefaultConfig()
	config.MaxRetries = maxRetries
	return WithConfig(fn, config)
}

// WithExponentialBackoffContext retries with exponential backoff and context
func WithExponentialBackoffContext(ctx context.Context, fn func() error, maxRetries int) error {
	config := DefaultConfig()
	config.MaxRetries = maxRetries
	return WithConfigContext(ctx, fn, config)
}

// WithConfig retries a function with custom configuration
func WithConfig(fn func() error, config *Config) error {
	return WithConfigContext(context.Background(), fn, config)
}

// WithConfigContext retries a function with custom configuration and context
func WithConfigContext(ctx context.Context, fn func() error, config *Config) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success!
		}

		lastErr = err

		// Check if we should retry this error
		if !config.ShouldRetry(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Sleep with exponential backoff
		time.Sleep(delay)

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}

// Do retries a function that returns a value with exponential backoff
func Do[T any](fn func() (T, error), maxRetries int) (T, error) {
	config := DefaultConfig()
	config.MaxRetries = maxRetries
	return DoWithConfig(fn, config)
}

// DoWithContext retries a function with context
func DoWithContext[T any](ctx context.Context, fn func() (T, error), maxRetries int) (T, error) {
	config := DefaultConfig()
	config.MaxRetries = maxRetries
	return DoWithConfigContext(ctx, fn, config)
}

// DoWithConfig retries a function with custom configuration
func DoWithConfig[T any](fn func() (T, error), config *Config) (T, error) {
	return DoWithConfigContext(context.Background(), fn, config)
}

// DoWithConfigContext retries a function with custom configuration and context
func DoWithConfigContext[T any](ctx context.Context, fn func() (T, error), config *Config) (T, error) {
	var result T
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		val, err := fn()
		if err == nil {
			return val, nil // Success!
		}

		lastErr = err

		// Check if we should retry this error
		if !config.ShouldRetry(err) {
			return result, fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Sleep with exponential backoff
		time.Sleep(delay)

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return result, fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}

// CalculateDelay calculates the delay for a given attempt number
func CalculateDelay(attempt int, config *Config) time.Duration {
	delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt)))
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	return delay
}
