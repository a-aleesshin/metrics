package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoWithDelays(t *testing.T) {
	retriableErr := errors.New("temporary error")
	nonRetriableErr := errors.New("permanent error")

	tests := []struct {
		name      string
		delays    []time.Duration
		retriable func(error) bool
		operation func(*int) error
		wantErr   error
		wantCalls int
	}{
		{
			name:      "success first try",
			delays:    []time.Duration{time.Nanosecond, time.Nanosecond, time.Nanosecond},
			retriable: func(error) bool { return true },
			operation: func(calls *int) error {
				(*calls)++
				return nil
			},
			wantCalls: 1,
		},
		{
			name:      "success after retriable error",
			delays:    []time.Duration{time.Nanosecond, time.Nanosecond, time.Nanosecond},
			retriable: func(err error) bool { return errors.Is(err, retriableErr) },
			operation: func(calls *int) error {
				(*calls)++
				if *calls < 3 {
					return retriableErr
				}
				return nil
			},
			wantCalls: 3,
		},
		{
			name:      "returns last retriable error after retries exhausted",
			delays:    []time.Duration{time.Nanosecond, time.Nanosecond, time.Nanosecond},
			retriable: func(err error) bool { return errors.Is(err, retriableErr) },
			operation: func(calls *int) error {
				(*calls)++
				return retriableErr
			},
			wantErr:   retriableErr,
			wantCalls: 4,
		},
		{
			name:      "non retriable error is not retried",
			delays:    []time.Duration{time.Nanosecond, time.Nanosecond, time.Nanosecond},
			retriable: func(err error) bool { return errors.Is(err, retriableErr) },
			operation: func(calls *int) error {
				(*calls)++
				return nonRetriableErr
			},
			wantErr:   nonRetriableErr,
			wantCalls: 1,
		},
		{
			name:      "nil retriable function is not retried",
			delays:    []time.Duration{time.Nanosecond, time.Nanosecond, time.Nanosecond},
			retriable: nil,
			operation: func(calls *int) error {
				(*calls)++
				return retriableErr
			},
			wantErr:   retriableErr,
			wantCalls: 1,
		},
		{
			name:      "empty delays returns first retriable error",
			delays:    nil,
			retriable: func(err error) bool { return errors.Is(err, retriableErr) },
			operation: func(calls *int) error {
				(*calls)++
				return retriableErr
			},
			wantErr:   retriableErr,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			calls := 0

			// Act
			err := DoWithDelays(
				context.Background(),
				tt.delays,
				tt.retriable,
				func() error {
					return tt.operation(&calls)
				},
			)

			// Assert
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}

				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if calls != tt.wantCalls {
				t.Fatalf("expected calls=%d, got %d", tt.wantCalls, calls)
			}
		})
	}
}

func TestDoWithDelays_ContextCanceled(t *testing.T) {
	retriableErr := errors.New("temporary error")

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0

	// Act
	err := DoWithDelays(
		ctx,
		[]time.Duration{time.Second},
		func(err error) bool { return errors.Is(err, retriableErr) },
		func() error {
			calls++
			return retriableErr
		},
	)

	// Assert
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error %v, got %v", context.Canceled, err)
	}

	if calls != 1 {
		t.Fatalf("expected calls=1, got %d", calls)
	}
}
