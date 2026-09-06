package metrics_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	sharedrouter "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/platform/id"
	platformlogger "github.com/a-aleesshin/metrics/internal/platform/logger"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/metrics"
	"go.uber.org/zap"
)

// newExampleServer собирает роутер сервера метрик на in-memory хранилище —
// минимальная конфигурация без middleware.
func newExampleServer() http.Handler {
	storage := memory.NewMemStorage()
	logger := platformlogger.NewZapLogger(zap.NewNop())

	updateUC := usecase.NewUpdateMetric(storage, logger, nil)
	valueUC := usecase.NewGetValueMetricUseCase(storage)
	listUC := usecase.NewListMetricUseCase(storage)
	updatesUC := usecase.NewUpdatesMetricsUseCase(storage, id.NewUUIDV7Generator())

	handler := metrics.NewHandler(
		metrics.NewUpdateHandler(updateUC, nil),
		metrics.NewUpdateJSONHandler(updateUC, nil),
		metrics.NewUpdatesHandler(updatesUC, nil),
		metrics.NewValueHandler(valueUC),
		metrics.NewValueJSONHandler(valueUC),
		metrics.NewListMetricsHandler(listUC),
	)

	return sharedrouter.New(nil, handler)
}

// ExampleUpdateHandler_Update — обновление метрики через URL-параметры:
// POST /update/{type}/{name}/{value}.
func ExampleUpdateHandler_Update() {
	server := newExampleServer()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

// ExampleUpdateJSONHandler_UpdateJSON — обновление метрики JSON-телом:
// POST /update. В ответ сервер возвращает принятую метрику.
func ExampleUpdateJSONHandler_UpdateJSON() {
	server := newExampleServer()

	body := `{"id":"Alloc","type":"gauge","value":123.45}`
	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":123.45}
}

// ExampleUpdatesHandler_Updates — батч-обновление метрик: POST /updates
// принимает массив метрик одним запросом.
func ExampleUpdatesHandler_Updates() {
	server := newExampleServer()

	body := `[
		{"id":"Alloc","type":"gauge","value":123.45},
		{"id":"PollCount","type":"counter","delta":5}
	]`
	req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

// ExampleValueHandler_Value — чтение значения метрики в plain text:
// GET /value/{type}/{name}.
func ExampleValueHandler_Value() {
	server := newExampleServer()

	update := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
	server.ServeHTTP(httptest.NewRecorder(), update)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: 123.45
}

// ExampleValueJSONHandler_ValueJSON — чтение значения метрики JSON-запросом:
// POST /value с телом {"id": ..., "type": ...}.
func ExampleValueJSONHandler_ValueJSON() {
	server := newExampleServer()

	update := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/5", nil)
	server.ServeHTTP(httptest.NewRecorder(), update)

	body := `{"id":"PollCount","type":"counter"}`
	req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output: {"id":"PollCount","type":"counter","delta":5}
}

// ExampleListMetricsHandler_List — HTML-страница со всеми метриками: GET /.
func ExampleListMetricsHandler_List() {
	server := newExampleServer()

	update := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
	server.ServeHTTP(httptest.NewRecorder(), update)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(strings.Contains(rec.Body.String(), "gauge Alloc = 123.45"))
	// Output:
	// 200
	// true
}
