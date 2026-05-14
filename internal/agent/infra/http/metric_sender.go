package httpadapter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	dto2 "github.com/a-aleesshin/metrics/internal/agent/infra/dto"
	"github.com/a-aleesshin/metrics/internal/agent/infra/mapper"
	"github.com/a-aleesshin/metrics/internal/platform/retry"
)

type MetricSender struct {
	url    string
	client *http.Client
}

func NewMetricSender(url string, HTTPClient *http.Client) *MetricSender {
	return &MetricSender{
		url:    normalizeBaseURL(url),
		client: HTTPClient,
	}
}

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
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)

	if _, err := gz.Write(body); err != nil {
		_ = gz.Close()
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}

	gzBody := gzBuf.Bytes()

	return retry.Do(context.Background(), isRetriableHTTPError, func() error {
		request, err := http.NewRequest(
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
	})
}
