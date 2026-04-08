package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

type collectSpy struct {
	called chan struct{}
	err    error
}

func (s *collectSpy) Execute() error {
	select {
	case s.called <- struct{}{}:
	default:
	}
	return s.err
}

type reportSpy struct {
	called chan struct{}
	err    error
}

func (s *reportSpy) Execute() error {
	select {
	case s.called <- struct{}{}:
	default:
	}
	return s.err
}

func TestAgentRunner_Run(t *testing.T) {
	t.Run("returns nil when context canceled", func(t *testing.T) {
		// Arrange
		c := &collectSpy{called: make(chan struct{}, 1)}
		rp := &reportSpy{called: make(chan struct{}, 1)}
		r := NewAgentRunner(c, rp, 10*time.Millisecond, 10*time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		err := r.Run(ctx)

		// Assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("returns collect error", func(t *testing.T) {
		// Arrange
		wantErr := errors.New("collect failed")
		c := &collectSpy{called: make(chan struct{}, 1), err: wantErr}
		rp := &reportSpy{called: make(chan struct{}, 1)}
		r := NewAgentRunner(c, rp, 5*time.Millisecond, time.Hour)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- r.Run(ctx)
		}()

		// Act + Assert
		select {
		case err := <-done:
			if !errors.Is(err, wantErr) {
				t.Fatalf("expected %v, got %v", wantErr, err)
			}
		case <-time.After(300 * time.Millisecond):
			t.Fatal("expected collect error, got timeout")
		}
	})
}
