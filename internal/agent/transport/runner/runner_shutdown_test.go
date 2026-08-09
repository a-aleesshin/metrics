package runner

import (
	"context"
	"testing"
	"time"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
)

func TestAgentRunner_Run_FlushesFinalReportOnCancel(t *testing.T) {
	c := &collectSpy{called: make(chan struct{}, 10)}
	rp := &reportSpy{
		buildCalled: make(chan struct{}, 10),
		sendCalled:  make(chan struct{}, 10),
		metrics:     []dto.MetricDTO{{Type: "gauge", Name: "Alloc", Value: "1"}},
	}
	logStub := &loggerStub{}

	r := NewAgentRunner(
		c,
		nil,
		rp,
		time.Hour, // сборщик и тикер отчёта не успеют сработать
		time.Hour,
		1,
		logStub,
	)

	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil on cancel, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runner did not stop after cancel")
	}

	select {
	case <-rp.sendCalled:
	default:
		t.Fatal("expected final report to be sent on shutdown")
	}
}
