package httpadapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	dto2 "github.com/a-aleesshin/metrics/internal/agent/infra/dto"
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
	payload := dto2.MetricsSend{
		ID:    dto.Name,
		MType: dto.Type,
	}

	switch dto.Type {
	case "gauge":
		v, err := strconv.ParseFloat(dto.Value, 64)

		if err != nil {
			return fmt.Errorf("invalid gauge value %q: %w", dto.Value, err)
		}

		if math.IsNaN(v) || math.IsInf(v, 0) {
			v = 0
		}

		payload.Value = &v
	case "counter":
		d, err := strconv.ParseInt(dto.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid metric value: %s: %w", dto.Value, err)
		}
		payload.Delta = &d
	default:
		return fmt.Errorf("unsupported metric type: %s", dto.Type)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, m.url+"/update", bytes.NewReader(body))

	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := m.client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}

func normalizeBaseURL(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}
