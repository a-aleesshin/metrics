package pool_test

import (
	"fmt"

	"github.com/a-aleesshin/metrics/internal/platform/pool"
	"github.com/a-aleesshin/metrics/internal/server/audit"
)

// ExampleNew показывает типичный цикл работы с пулом: взять объект,
// заполнить, использовать и вернуть Put сам сбросит состояние.
func ExampleNew() {
	events := pool.New(func() *audit.Event {
		return &audit.Event{}
	})

	event := events.Get()
	event.TS = 12345678
	event.Metrics = append(event.Metrics, "Alloc", "Frees")
	event.IPAddress = "192.168.0.42"

	fmt.Println(len(event.Metrics))

	events.Put(event)

	fmt.Println(event.TS, len(event.Metrics), event.IPAddress == "")
	// Output:
	// 2
	// 0 0 true
}
