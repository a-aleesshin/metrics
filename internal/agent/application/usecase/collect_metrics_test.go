package usecase

import (
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

type collectMetricsTest struct {
	name         string
	runtimeRider metricRuntimeReaderStub
	repository   metricRepositoryStub
	randomValue  randomValueStub
	wantErr      bool
}

type metricRuntimeReaderStub struct {
	metrics []reader.RuntimeMetric
}

func (m *metricRuntimeReaderStub) Read() []reader.RuntimeMetric {
	return m.metrics
}

type randomValueStub struct {
	value float64
}

func (m *randomValueStub) GenerateFloat64() float64 {
	return m.value
}

type metricRepositorySpy struct {
	gauges   []*metric.Gauge
	counters []*metric.Counter

	setGaugeErr   error
	addCounterErr error
}

func (m *metricRepositorySpy) SetGauge(gauge *metric.Gauge) error {
	if m.setGaugeErr != nil {
		return m.setGaugeErr
	}

	m.gauges = append(m.gauges, gauge)
	return nil
}

func (m *metricRepositorySpy) AddCounter(counter *metric.Counter) error {
	if m.addCounterErr != nil {
		return m.addCounterErr
	}

	m.counters = append(m.counters, counter)
	return nil
}

func (m *metricRepositorySpy) GetMetrics() (repository.MetricsState, error) {
	return repository.MetricsState{}, nil
}

func runtimeMetricsToMap(metrics []reader.RuntimeMetric) map[string]float64 {
	out := make(map[string]float64, len(metrics))
	for _, runtimeMetric := range metrics {
		out[runtimeMetric.Name] = runtimeMetric.Value
	}
	return out
}

func gaugesToMap(gauges []*metric.Gauge) map[string]float64 {
	out := make(map[string]float64, len(gauges))
	for _, gauge := range gauges {
		out[gauge.Name().String()] = gauge.Value()
	}

	return out
}

func countersToMap(counters []*metric.Counter) map[string]int64 {
	out := make(map[string]int64, len(counters))
	for _, counter := range counters {
		out[counter.Name().String()] = counter.Value()
	}
	return out
}

func TestCollectMetricsUseCase_Execute(t *testing.T) {
	runtimeMetrics := []reader.RuntimeMetric{
		{Name: "Alloc", Value: 100},
		{Name: "BuckHashSys", Value: 101},
		{Name: "Frees", Value: 102},
		{Name: "GCCPUFraction", Value: 103},
		{Name: "GCSys", Value: 104},
		{Name: "HeapAlloc", Value: 105},
		{Name: "HeapIdle", Value: 106},
		{Name: "HeapInuse", Value: 107},
		{Name: "HeapObjects", Value: 108},
		{Name: "HeapReleased", Value: 109},
		{Name: "HeapSys", Value: 110},
		{Name: "LastGC", Value: 111},
		{Name: "Lookups", Value: 112},
		{Name: "MCacheInuse", Value: 113},
		{Name: "MCacheSys", Value: 114},
		{Name: "MSpanInuse", Value: 115},
		{Name: "MSpanSys", Value: 116},
		{Name: "Mallocs", Value: 117},
		{Name: "NextGC", Value: 118},
		{Name: "NumForcedGC", Value: 119},
		{Name: "NumGC", Value: 120},
		{Name: "OtherSys", Value: 121},
		{Name: "PauseTotalNs", Value: 122},
		{Name: "StackInuse", Value: 123},
		{Name: "StackSys", Value: 124},
		{Name: "Sys", Value: 125},
		{Name: "TotalAlloc", Value: 126},
	}

	runtimeRider := &metricRuntimeReaderStub{
		metrics: runtimeMetrics,
	}

	randomValue := &randomValueStub{
		value: 127.324,
	}

	repositorySpy := &metricRepositorySpy{}

	uc := NewCollectMetricsUseCase(runtimeRider, repositorySpy, randomValue)
	err := uc.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantGauges := runtimeMetricsToMap(runtimeMetrics)
	wantGauges["RandomValue"] = 127.324

	gotGauges := gaugesToMap(repositorySpy.gauges)

	if len(gotGauges) != len(wantGauges) {
		t.Fatalf("expected %d gauges, got %d", len(wantGauges), len(gotGauges))
	}

	for name, wantValue := range wantGauges {
		gotValue, ok := gotGauges[name]
		if !ok {
			t.Fatalf("expected gauge %q to be stored", name)
		}

		if gotValue != wantValue {
			t.Fatalf("gauge %q: expected %v, got %v", name, wantValue, gotValue)
		}
	}

	wantCounters := map[string]int64{
		"PollCount": 1,
	}

	gotCounters := countersToMap(repositorySpy.counters)

	if len(gotCounters) != len(wantCounters) {
		t.Fatalf("expected %d counters, got %d", len(wantCounters), len(gotCounters))
	}

	for name, wantValue := range wantCounters {
		gotValue, ok := gotCounters[name]
		if !ok {
			t.Fatalf("expected counter %q to be stored", name)
		}

		if gotValue != wantValue {
			t.Fatalf("counter %q: expected %v, got %v", name, wantValue, gotValue)
		}
	}
}
