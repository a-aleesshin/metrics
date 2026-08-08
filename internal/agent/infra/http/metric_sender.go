// Package httpadapter реализует отправку метрик на сервер по HTTP:
// JSON → gzip → подпись HMAC → POST с ретраями.
package httpadapter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	dto2 "github.com/a-aleesshin/metrics/internal/agent/infra/dto"
	"github.com/a-aleesshin/metrics/internal/agent/infra/mapper"
)

// gzip.Writer держит ~800 КБ внутренних буферов flate — создание на каждую
// отправку было главным источником аллокаций по профилю, поэтому переиспользуем.
var gzipWriterPool = sync.Pool{
	New: func() any { return gzip.NewWriter(io.Discard) },
}

var gzipBufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// HTTPClient — минимальный интерфейс HTTP-клиента; ему удовлетворяет *http.Client.
type HTTPClient interface {
	// Do выполняет HTTP-запрос и возвращает ответ.
	Do(request *http.Request) (*http.Response, error)
}

// MetricSender шлёт метрики на сервер gzip-сжатым JSON через POST /update и /updates.
type MetricSender struct {
	url    string
	client HTTPClient
}

// NewMetricSender создаёт отправитель метрик; url нормализуется до базового адреса сервера.
func NewMetricSender(url string, HTTPClient HTTPClient) *MetricSender {
	return &MetricSender{
		url:    normalizeBaseURL(url),
		client: HTTPClient,
	}
}

// Send отправляет одну метрику запросом POST /update.
func (m *MetricSender) Send(dto dto.MetricDTO) error {
	payload, err := mapper.ToSendMetric(dto)

	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return m.sendGzippedJSON("/update", body)
}

// SendBatch отправляет пачку метрик одним запросом POST /updates; пустая пачка не отправляется.
func (m *MetricSender) SendBatch(metrics []dto.MetricDTO) error {
	if len(metrics) == 0 {
		return nil
	}

	payload := make([]dto2.MetricsSend, 0, len(metrics))

	for _, metric := range metrics {
		metricSendDTO, err := mapper.ToSendMetric(metric)

		if err != nil {
			return err
		}

		payload = append(payload, metricSendDTO)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return m.sendGzippedJSON("/updates", body)
}

func normalizeBaseURL(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}

func (m *MetricSender) sendGzippedJSON(path string, body []byte) error {
	gzBuf := gzipBufferPool.Get().(*bytes.Buffer)
	gzBuf.Reset()
	// буфер нельзя вернуть в пул раньше: gzBody ссылается на него,
	// пока client.Do (включая ретраи) не завершится
	defer gzipBufferPool.Put(gzBuf)

	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(gzBuf)

	if _, err := gz.Write(body); err != nil {
		_ = gz.Close()
		gzipWriterPool.Put(gz)
		return err
	}

	if err := gz.Close(); err != nil {
		gzipWriterPool.Put(gz)
		return err
	}

	gzipWriterPool.Put(gz)

	gzBody := gzBuf.Bytes()

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		m.url+path,
		bytes.NewReader(gzBody),
	)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	response, err := m.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return unexpectedStatusError{code: response.StatusCode}
	}

	return nil
}
