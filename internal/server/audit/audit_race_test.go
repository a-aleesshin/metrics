package audit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

type countingObserver struct {
	calls atomic.Int64
}

func (o *countingObserver) Notify(_ context.Context, _ Event) error {
	o.calls.Add(1)
	return nil
}

func TestPublisher_ConcurrentRegisterAndPublish(t *testing.T) {
	publisher := NewPublisher(nil, &countingObserver{})
	ctx := context.Background()

	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				publisher.Register(&countingObserver{})
			}
		}()

		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				publisher.Publish(ctx, []string{"Alloc"}, "10.0.0.1")
			}
		}()
	}

	wg.Wait()
}
