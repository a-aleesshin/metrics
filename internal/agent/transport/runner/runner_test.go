package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
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
	buildCalled chan struct{}
	sendCalled  chan struct{}
	metrics     []dto.MetricDTO
	buildErr    error
	sendErr     error
}

func (s *reportSpy) BuildMetrics() ([]dto.MetricDTO, error) {
	select {
	case s.buildCalled <- struct{}{}:
	default:
	}
	return s.metrics, s.buildErr
}

func (s *reportSpy) SendMetrics(_ []dto.MetricDTO) error {
	select {
	case s.sendCalled <- struct{}{}:
	default:
	}
	return s.sendErr
}

type loggerStub struct {
	mu      sync.Mutex
	errMsgs []string
}

func (l *loggerStub) Info(msg string, fields ...portlogger.Field) {}

func (l *loggerStub) Error(msg string, fields ...portlogger.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.errMsgs = append(l.errMsgs, msg)
}

func (l *loggerStub) ErrorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.errMsgs)
}

func TestAgentRunner_Run_LogsErrorAndContinues(t *testing.T) {
	// Arrange
	c := &collectSpy{called: make(chan struct{}, 10), err: errors.New("collect failed")}
	rp := &reportSpy{
		buildCalled: make(chan struct{}, 10),
		sendCalled:  make(chan struct{}, 10),
	}
	logStub := &loggerStub{}

	r := NewAgentRunner(
		c,
		nil,
		rp,
		5*time.Millisecond,
		200*time.Millisecond,
		1,
		logStub,
	)

	ctx, cancel := context.WithCancel(t.Context())
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

	if logStub.ErrorCount() == 0 {
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
	rp := &reportSpy{
		buildCalled: make(chan struct{}, 10),
		sendCalled:  make(chan struct{}, 10),
		metrics: []dto.MetricDTO{
			{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "1",
			},
		},
		sendErr: errors.New("report failed"),
	}
	logStub := &loggerStub{}

	r := NewAgentRunner(
		c,
		nil,
		rp,
		200*time.Millisecond,
		5*time.Millisecond,
		1,
		logStub,
	)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	// Act
	select {
	case <-rp.sendCalled:
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

	if logStub.ErrorCount() == 0 {
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

type slowReportSpy struct {
	mu        sync.Mutex
	active    int
	maxActive int
	calls     int
	called    chan struct{}
	blockFor  time.Duration
}

func (s *slowReportSpy) BuildMetrics() ([]dto.MetricDTO, error) {
	return []dto.MetricDTO{
		{
			Type:  "gauge",
			Name:  "Alloc",
			Value: "1",
		},
	}, nil
}

func (s *slowReportSpy) SendMetrics(_ []dto.MetricDTO) error {
	s.mu.Lock()
	s.active++
	s.calls++
	if s.active > s.maxActive {
		s.maxActive = s.active
	}
	s.mu.Unlock()

	select {
	case s.called <- struct{}{}:
	default:
	}

	time.Sleep(s.blockFor)

	s.mu.Lock()
	s.active--
	s.mu.Unlock()

	return nil
}

func (s *slowReportSpy) MaxActive() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.maxActive
}

func TestAgentRunner_Run_LimitsConcurrentReports(t *testing.T) {
	// Arrange
	c := &collectSpy{called: make(chan struct{}, 10)}
	rp := &slowReportSpy{
		called:   make(chan struct{}, 10),
		blockFor: 50 * time.Millisecond,
	}
	logStub := &loggerStub{}

	r := NewAgentRunner(
		c,
		nil,
		rp,
		200*time.Millisecond,
		5*time.Millisecond,
		2,
		logStub,
	)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	// Act
	for i := 0; i < 3; i++ {
		select {
		case <-rp.called:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("report usecase was not called")
		}
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil on context cancel, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runner did not stop after cancel")
	}

	// Assert
	if rp.MaxActive() > 2 {
		t.Fatalf("expected max concurrent reports <= 2, got %d", rp.MaxActive())
	}
}
