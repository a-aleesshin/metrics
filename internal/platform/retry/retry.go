package retry

import (
	"context"
	"time"
)

var defaultDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

type IsRetriableFunc func(error) bool

func Do(ctx context.Context, isRetriable IsRetriableFunc, operation func() error) error {
	return DoWithDelays(ctx, defaultDelays, isRetriable, operation)
}

func DoWithDelays(
	ctx context.Context,
	delays []time.Duration,
	isRetriable IsRetriableFunc,
	operation func() error,
) error {
	var err error

	for attempt := 0; ; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		if isRetriable == nil || !isRetriable(err) {
			return err
		}

		if attempt >= len(delays) {
			return err
		}

		if sleepErr := sleep(ctx, delays[attempt]); sleepErr != nil {
			return sleepErr
		}
	}
}

func sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
