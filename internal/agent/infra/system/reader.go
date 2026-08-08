package systemadapter

import (
	"fmt"
	"time"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type virtualMemoryReader func() (*mem.VirtualMemoryStat, error)
type cpuPercentReader func(interval time.Duration, percpu bool) ([]float64, error)

type GopsutilReader struct {
	virtualMemory virtualMemoryReader
	cpuPercent    cpuPercentReader
}

func NewGopsutilReader() *GopsutilReader {
	return newGopsutilReader(mem.VirtualMemory, cpu.Percent)
}

func newGopsutilReader(virtualMemory virtualMemoryReader, cpuPercent cpuPercentReader) *GopsutilReader {
	return &GopsutilReader{
		virtualMemory: virtualMemory,
		cpuPercent:    cpuPercent,
	}
}

func (r *GopsutilReader) Read() ([]reader.SystemMetric, error) {
	memoryStats, err := r.virtualMemory()
	if err != nil {
		return nil, fmt.Errorf("read virtual memory: %w", err)
	}

	cpuStats, err := r.cpuPercent(0, true)
	if err != nil {
		return nil, fmt.Errorf("read cpu percent: %w", err)
	}

	metrics := make([]reader.SystemMetric, 0, 2+len(cpuStats))
	metrics = append(metrics,
		reader.SystemMetric{Name: "TotalMemory", Value: float64(memoryStats.Total)},
		reader.SystemMetric{Name: "FreeMemory", Value: float64(memoryStats.Free)},
	)

	for i, value := range cpuStats {
		metrics = append(metrics, reader.SystemMetric{
			Name:  fmt.Sprintf("CPUutilization%d", i+1),
			Value: value,
		})
	}

	return metrics, nil
}
