package retry

import (
	"context"
	"time"
)

// Retry retries a specific function with the given frequency.
// A context with timeout or deadline can be passed to set a timeout.
// The passed function will be retried until it either returns no error, or a timeout happens.
// An error is only returned by this function if a timeout happened.
func Retry(ctx context.Context, frequency time.Duration, f func() error) error {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if f() == nil {
				return nil
			}
		}
	}
}
