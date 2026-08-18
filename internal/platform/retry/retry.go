// Package retry выполняет операции с повторами и настраиваемыми задержками между попытками.
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

// IsRetriableFunc решает, стоит ли повторять операцию после данной ошибки.
type IsRetriableFunc func(error) bool

// DefaultDelays возвращает копию задержек по умолчанию (1s, 3s, 5s).
func DefaultDelays() []time.Duration {
	out := make([]time.Duration, len(defaultDelays))
	copy(out, defaultDelays)

	return out
}

// Wait ждёт delay или отмену ctx; при отмене возвращает ctx.Err().
func Wait(ctx context.Context, delay time.Duration) error {
	return sleep(ctx, delay)
}

// Do выполняет operation с задержками по умолчанию (1s, 3s, 5s) между повторами.
func Do(ctx context.Context, isRetriable IsRetriableFunc, operation func() error) error {
	return DoWithDelays(ctx, defaultDelays, isRetriable, operation)
}

// DoWithDelays выполняет operation до len(delays)+1 раз, ожидая delays[i] перед i-м повтором.
// Повторы прекращаются, если isRetriable вернул false или ctx отменён.
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
