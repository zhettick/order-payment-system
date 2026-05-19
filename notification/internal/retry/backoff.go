package retry

import (
	"context"
	"time"
)

func Do(ctx context.Context, maxAttempts int, baseBackoff time.Duration, fn func() error) error {
	var lastErr error
	delay := baseBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == maxAttempts {
			break
		}

		select {
		case <-time.After(delay):
			delay *= 2
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}
