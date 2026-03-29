package runtimeadapter

import (
	"runtime"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
)

type MetricRuntimeReader struct {
}

func NewMetricRuntimeReader() *MetricRuntimeReader {
	return &MetricRuntimeReader{}
}

func (m *MetricRuntimeReader) Read() []reader.RuntimeMetric {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return []reader.RuntimeMetric{
		{Name: "Alloc", Value: float64(stats.Alloc)},
		{Name: "BuckHashSys", Value: float64(stats.BuckHashSys)},
		{Name: "Frees", Value: float64(stats.Frees)},
		{Name: "GCCPUFraction", Value: float64(stats.GCCPUFraction)},
		{Name: "GCSys", Value: float64(stats.GCSys)},
		{Name: "HeapAlloc", Value: float64(stats.HeapAlloc)},
		{Name: "HeapIdle", Value: float64(stats.HeapIdle)},
		{Name: "HeapInuse", Value: float64(stats.HeapInuse)},
		{Name: "HeapObjects", Value: float64(stats.HeapObjects)},
		{Name: "HeapReleased", Value: float64(stats.HeapReleased)},
		{Name: "HeapSys", Value: float64(stats.HeapSys)},
		{Name: "LastGC", Value: float64(stats.LastGC)},
		{Name: "Lookups", Value: float64(stats.Lookups)},
		{Name: "MCacheInuse", Value: float64(stats.MCacheInuse)},
		{Name: "MCacheSys", Value: float64(stats.MCacheSys)},
		{Name: "MSpanInuse", Value: float64(stats.MSpanInuse)},
		{Name: "MSpanSys", Value: float64(stats.MSpanSys)},
		{Name: "Mallocs", Value: float64(stats.Mallocs)},
		{Name: "NextGC", Value: float64(stats.NextGC)},
		{Name: "NumForcedGC", Value: float64(stats.NumForcedGC)},
		{Name: "NumGC", Value: float64(stats.NumGC)},
		{Name: "OtherSys", Value: float64(stats.OtherSys)},
		{Name: "PauseTotalNs", Value: float64(stats.PauseTotalNs)},
		{Name: "StackInuse", Value: float64(stats.StackInuse)},
		{Name: "StackSys", Value: float64(stats.StackSys)},
		{Name: "Sys", Value: float64(stats.Sys)},
		{Name: "TotalAlloc", Value: float64(stats.TotalAlloc)},
	}
}
