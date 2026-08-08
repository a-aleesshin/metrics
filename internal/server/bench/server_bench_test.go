package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/platform/health"
	sharedrouter "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/platform/id"
	platformlogger "github.com/a-aleesshin/metrics/internal/platform/logger"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/audit"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/healths"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/metrics"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/middleware"
	"go.uber.org/zap"
)

const hashKey = "bench-key"

type metricsPayload struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func newBenchServer(b *testing.B) (http.Handler, *memory.MemStorage) {
	b.Helper()

	storage := memory.NewMemStorage()
	appLogger := platformlogger.NewZapLogger(zap.NewNop())

	updateUC := usecase.NewUpdateMetric(storage, appLogger, nil)
	valueUC := usecase.NewGetValueMetricUseCase(storage)
	listUC := usecase.NewListMetricUseCase(storage)
	updatesUC := usecase.NewUpdatesMetricsUseCase(storage, id.NewUUIDV7Generator())

	auditPublisher := audit.NewPublisher(appLogger)

	metricsHandler := metrics.NewHandler(
		metrics.NewUpdateHandler(updateUC, auditPublisher),
		metrics.NewUpdateJsonHandler(updateUC, auditPublisher),
		metrics.NewUpdatesHandler(updatesUC, auditPublisher),
		metrics.NewValueHandler(valueUC),
		metrics.NewValueJsonHandler(valueUC),
		metrics.NewListMetricsHandler(listUC),
	)

	healthHandler := healths.NewHandler(healths.NewPingHandler(health.NewService()))

	router := sharedrouter.New(
		[]func(http.Handler) http.Handler{
			middleware.WithHashSHA256(hashKey),
			middleware.DecompressRequest,
			middleware.CompressResponse,
			middleware.RequestLogger(zap.NewNop()),
		},
		metricsHandler,
		healthHandler,
	)

	return router, storage
}

func seedStorage(b *testing.B, storage *memory.MemStorage, gauges, counters int) {
	b.Helper()

	ctx := context.Background()

	for i := 0; i < gauges; i++ {
		g, err := metric.NewGauge(fmt.Sprintf("id-g-%d", i), fmt.Sprintf("Gauge%d", i), float64(i)*1.5)
		if err != nil {
			b.Fatalf("new gauge: %v", err)
		}

		if err := storage.SaveGauge(ctx, g); err != nil {
			b.Fatalf("save gauge: %v", err)
		}
	}

	for i := 0; i < counters; i++ {
		c, err := metric.NewCounter(fmt.Sprintf("id-c-%d", i), fmt.Sprintf("Counter%d", i), int64(i))
		if err != nil {
			b.Fatalf("new counter: %v", err)
		}

		if err := storage.SaveCounter(ctx, c); err != nil {
			b.Fatalf("save counter: %v", err)
		}
	}
}

func mustJSON(b *testing.B, v any) []byte {
	b.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		b.Fatalf("marshal payload: %v", err)
	}

	return data
}

func serve(b *testing.B, handler http.Handler, req *http.Request, wantStatus int) {
	b.Helper()

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != wantStatus {
		b.Fatalf("unexpected status: got %d, want %d", rec.Code, wantStatus)
	}
}

func BenchmarkHTTP_UpdatePlain(b *testing.B) {
	handler, _ := newBenchServer(b)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
		serve(b, handler, req, http.StatusOK)
	}
}

func BenchmarkHTTP_UpdateJSON(b *testing.B) {
	handler, _ := newBenchServer(b)

	value := 123.45
	body := mustJSON(b, metricsPayload{ID: "Alloc", MType: "gauge", Value: &value})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		serve(b, handler, req, http.StatusOK)
	}
}

func BenchmarkHTTP_UpdatesBatch(b *testing.B) {
	handler, _ := newBenchServer(b)

	const batchSize = 30

	payload := make([]metricsPayload, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		if i%2 == 0 {
			value := float64(i) * 1.5
			payload = append(payload, metricsPayload{ID: fmt.Sprintf("Gauge%d", i), MType: "gauge", Value: &value})
			continue
		}

		delta := int64(i)
		payload = append(payload, metricsPayload{ID: fmt.Sprintf("Counter%d", i), MType: "counter", Delta: &delta})
	}

	body := mustJSON(b, payload)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		serve(b, handler, req, http.StatusOK)
	}
}

func BenchmarkHTTP_ValueJSON(b *testing.B) {
	handler, storage := newBenchServer(b)
	seedStorage(b, storage, 100, 50)

	body := mustJSON(b, metricsPayload{ID: "Gauge1", MType: "gauge"})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		serve(b, handler, req, http.StatusOK)
	}
}

func BenchmarkHTTP_List(b *testing.B) {
	handler, storage := newBenchServer(b)
	seedStorage(b, storage, 100, 50)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		serve(b, handler, req, http.StatusOK)
	}
}

func BenchmarkMemStorage_UpdateBatch(b *testing.B) {
	storage := memory.NewMemStorage()
	ctx := context.Background()

	const batchSize = 30

	batch := repository.MetricBatch{
		Gauges:   make([]*metric.Gauge, 0, batchSize/2),
		Counters: make([]*metric.Counter, 0, batchSize/2),
	}

	for i := 0; i < batchSize/2; i++ {
		g, err := metric.NewGauge(fmt.Sprintf("id-g-%d", i), fmt.Sprintf("Gauge%d", i), float64(i))
		if err != nil {
			b.Fatalf("new gauge: %v", err)
		}
		batch.Gauges = append(batch.Gauges, g)

		c, err := metric.NewCounter(fmt.Sprintf("id-c-%d", i), fmt.Sprintf("Counter%d", i), int64(i))
		if err != nil {
			b.Fatalf("new counter: %v", err)
		}
		batch.Counters = append(batch.Counters, c)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := storage.UpdateBatch(ctx, batch); err != nil {
			b.Fatalf("update batch: %v", err)
		}
	}
}

type noopObserver struct{}

func (noopObserver) Notify(_ context.Context, _ audit.Event) error { return nil }

func BenchmarkAuditPublisher_Publish(b *testing.B) {
	publisher := audit.NewPublisher(nil, noopObserver{})
	ctx := context.Background()
	names := []string{"Alloc", "Frees", "HeapAlloc", "PollCount"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		publisher.Publish(ctx, names, "192.168.0.42")
	}
}
