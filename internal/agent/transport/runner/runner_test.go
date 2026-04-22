package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	portlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
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

type loggerStub struct {
	errMsgs []string
}

func (l *loggerStub) Info(msg string, fields ...portlogger.Field) {}

func (l *loggerStub) Error(msg string, fields ...portlogger.Field) {
	l.errMsgs = append(l.errMsgs, msg)
}

func TestAgentRunner_Run_LogsErrorAndContinues(t *testing.T) {
	// Arrange
	c := &collectSpy{called: make(chan struct{}, 10), err: errors.New("collect failed")}
	rp := &reportSpy{called: make(chan struct{}, 10)}
	logStub := &loggerStub{}

	r := NewAgentRunner(
		c,
		rp,
		5*time.Millisecond,
		200*time.Millisecond,
		logStub,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	// Act
	select {
	case <-c.called:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("collect usecase was not called")
	}

	time.Sleep(20 * time.Millisecond)

	select {
	case err := <-done:
		t.Fatalf("runner should continue on collect error, got premature exit: %v", err)
	default:
	}

	if len(logStub.errMsgs) == 0 {
		t.Fatal("expected error log, got none")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil on context cancel, got %v", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("runner did not stop after cancel")
	}
}

func TestAgentRunner_Run_LogsReportErrorAndContinues(t *testing.T) {
	// Arrange
	c := &collectSpy{called: make(chan struct{}, 10)}
	rp := &reportSpy{called: make(chan struct{}, 10), err: errors.New("report failed")}
	logStub := &loggerStub{}

	r := NewAgentRunner(
		c,
		rp,
		200*time.Millisecond,
		5*time.Millisecond,
		logStub,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	// Act
	select {
	case <-rp.called:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("report usecase was not called")
	}

	time.Sleep(20 * time.Millisecond)

	// Assert
	select {
	case err := <-done:
		t.Fatalf("runner should continue on report error, got premature exit: %v", err)
	default:
	}

	if len(logStub.errMsgs) == 0 {
		t.Fatal("expected error log, got none")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil on context cancel, got %v", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("runner did not stop after cancel")
	}
}
