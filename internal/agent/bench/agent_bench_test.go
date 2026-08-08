package bench

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	"github.com/a-aleesshin/metrics/internal/agent/application/usecase"
	httpadapter "github.com/a-aleesshin/metrics/internal/agent/infra/http"
	"github.com/a-aleesshin/metrics/internal/agent/infra/persistence/memory"
	randomadapter "github.com/a-aleesshin/metrics/internal/agent/infra/random"
	runtimeadapter "github.com/a-aleesshin/metrics/internal/agent/infra/runtime"
)

const hashKey = "bench-key"

// fakeHTTPClient считает тело запроса и отвечает 200 — сеть не трогаем,
// чтобы профиль показывал аллокации кода агента, а не сокетов.
type fakeHTTPClient struct{}

func (fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}, nil
}

func newCollectUseCase(repo *memory.MemMetricRepository) *usecase.CollectMetricsUseCase {
	return usecase.NewCollectMetricsUseCase(
		runtimeadapter.NewMetricRuntimeReader(),
		repo,
		randomadapter.NewRandomValueAdapter(),
	)
}

func newReportUseCase(repo *memory.MemMetricRepository) *usecase.ReportMetricsUseCase {
	sender := httpadapter.NewMetricSender(
		"localhost:8080",
		httpadapter.NewSigningClient(fakeHTTPClient{}, hashKey),
	)

	return usecase.NewReportMetricsUseCase(repo, sender)
}

func seedRepo(b *testing.B, repo *memory.MemMetricRepository) {
	b.Helper()

	if err := newCollectUseCase(repo).Execute(); err != nil {
		b.Fatalf("seed repo: %v", err)
	}
}

func buildBatch(b *testing.B, size int) []dto.MetricDTO {
	b.Helper()

	batch := make([]dto.MetricDTO, 0, size)

	for i := 0; i < size; i++ {
		if i%2 == 0 {
			batch = append(batch, dto.MetricDTO{
				Type:  "gauge",
				Name:  fmt.Sprintf("Gauge%d", i),
				Value: fmt.Sprintf("%d.5", i),
			})
			continue
		}

		batch = append(batch, dto.MetricDTO{
			Type:  "counter",
			Name:  fmt.Sprintf("Counter%d", i),
			Value: fmt.Sprintf("%d", i),
		})
	}

	return batch
}

func BenchmarkCollectMetrics(b *testing.B) {
	repo := memory.NewMemMetricRepository()
	collect := newCollectUseCase(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := collect.Execute(); err != nil {
			b.Fatalf("collect: %v", err)
		}
	}
}

func BenchmarkBuildMetrics(b *testing.B) {
	repo := memory.NewMemMetricRepository()
	seedRepo(b, repo)

	report := newReportUseCase(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := report.BuildMetrics(); err != nil {
			b.Fatalf("build metrics: %v", err)
		}
	}
}

func BenchmarkReportMetrics(b *testing.B) {
	repo := memory.NewMemMetricRepository()
	seedRepo(b, repo)

	report := newReportUseCase(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := report.Execute(); err != nil {
			b.Fatalf("report: %v", err)
		}
	}
}

func BenchmarkSendBatch(b *testing.B) {
	sender := httpadapter.NewMetricSender(
		"localhost:8080",
		httpadapter.NewSigningClient(fakeHTTPClient{}, hashKey),
	)

	batch := buildBatch(b, 30)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := sender.SendBatch(batch); err != nil {
			b.Fatalf("send batch: %v", err)
		}
	}
}
