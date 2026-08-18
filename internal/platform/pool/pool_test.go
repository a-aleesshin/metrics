package pool_test

import (
	"sync"
	"testing"

	"github.com/a-aleesshin/metrics/internal/platform/pool"
	"github.com/a-aleesshin/metrics/internal/platform/resettest"
)

func TestPool_GetUsesFactoryWhenEmpty(t *testing.T) {
	calls := 0

	p := pool.New(func() *resettest.Container {
		calls++
		return &resettest.Container{}
	})

	first := p.Get()
	if first == nil {
		t.Fatal("expected object from factory, got nil")
	}

	if calls != 1 {
		t.Fatalf("expected 1 factory call, got %d", calls)
	}
}

func TestPool_NilFactoryReturnsZeroValue(t *testing.T) {
	p := pool.New[*resettest.Container](nil)

	if got := p.Get(); got != nil {
		t.Fatalf("expected zero value (nil) from empty pool without factory, got %+v", got)
	}

	// Возвращённые в пул объекты переиспользуются и без фабрики.
	p.Put(&resettest.Container{Num: 42})

	got := p.Get()
	if got == nil {
		t.Fatal("expected pooled object after Put, got nil")
	}

	if got.Num != 0 {
		t.Fatalf("expected object reset on Put, got Num=%d", got.Num)
	}
}

func TestPool_PutResetsObject(t *testing.T) {
	p := pool.New(func() *resettest.Container {
		return &resettest.Container{}
	})

	c := p.Get()
	c.Num = 42
	c.Str = "dirty"
	c.Items = append(c.Items, "a", "b")
	c.Index = map[string]int{"a": 1}

	itemsCap := cap(c.Items)

	p.Put(c)

	// Put сбрасывает объект на месте — проверяем по сохранённому указателю.
	if c.Num != 0 || c.Str != "" {
		t.Fatalf("expected primitives reset on Put, got %+v", c)
	}

	if len(c.Items) != 0 || cap(c.Items) != itemsCap {
		t.Fatalf("expected slice truncated with capacity %d preserved, got len=%d cap=%d",
			itemsCap, len(c.Items), cap(c.Items))
	}

	if len(c.Index) != 0 {
		t.Fatalf("expected cleared map, got %v", c.Index)
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	p := pool.New(func() *resettest.Container {
		return &resettest.Container{}
	})

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < 1000; j++ {
				c := p.Get()

				if c.Num != 0 || c.Str != "" || len(c.Items) != 0 {
					t.Error("got dirty object from pool")
					return
				}

				c.Num = j
				c.Str = "busy"
				c.Items = append(c.Items, "x")

				p.Put(c)
			}
		}()
	}

	wg.Wait()
}
