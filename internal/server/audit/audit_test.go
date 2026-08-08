package audit

import (
	"context"
	"errors"
	"testing"
	"time"
)

type observerSpy struct {
	events []Event
	err    error
}

func (o *observerSpy) Notify(_ context.Context, event Event) error {
	o.events = append(o.events, event)
	return o.err
}

func TestPublisher_Publish_NotifiesAllObservers(t *testing.T) {
	first := &observerSpy{}
	second := &observerSpy{}

	publisher := NewPublisher(nil, first, second)
	publisher.now = func() time.Time { return time.Unix(12345678, 0) }

	publisher.Publish(context.Background(), []string{"Alloc", "Frees"}, "192.168.0.42")

	want := Event{
		TS:        12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	for i, spy := range []*observerSpy{first, second} {
		if len(spy.events) != 1 {
			t.Fatalf("observer %d: expected 1 event, got %d", i, len(spy.events))
		}

		got := spy.events[0]
		if got.TS != want.TS || got.IPAddress != want.IPAddress {
			t.Fatalf("observer %d: expected event %+v, got %+v", i, want, got)
		}

		if len(got.Metrics) != 2 || got.Metrics[0] != "Alloc" || got.Metrics[1] != "Frees" {
			t.Fatalf("observer %d: unexpected metrics %v", i, got.Metrics)
		}
	}
}

func TestPublisher_Publish_ObserverErrorDoesNotStopOthers(t *testing.T) {
	failing := &observerSpy{err: errors.New("sink unavailable")}
	healthy := &observerSpy{}

	publisher := NewPublisher(nil, failing, healthy)

	publisher.Publish(context.Background(), []string{"Alloc"}, "10.0.0.1")

	if len(healthy.events) != 1 {
		t.Fatalf("expected healthy observer to receive event, got %d", len(healthy.events))
	}
}

func TestPublisher_Register(t *testing.T) {
	publisher := NewPublisher(nil)

	spy := &observerSpy{}
	publisher.Register(spy)
	publisher.Register(nil)

	publisher.Publish(context.Background(), []string{"Alloc"}, "10.0.0.1")

	if len(spy.events) != 1 {
		t.Fatalf("expected registered observer to receive event, got %d", len(spy.events))
	}
}
