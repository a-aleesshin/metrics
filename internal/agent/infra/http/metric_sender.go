package httpadapter

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
)

type MetricSender struct {
	url    string
	client *http.Client
}

func NewMetricSender(url string, HTTPClient *http.Client) *MetricSender {
	return &MetricSender{
		url:    url,
		client: HTTPClient,
	}
}

func (m *MetricSender) Send(dto dto.MetricDTO) error {
	var path string

	path = fmt.Sprintf(
		"%s/update/%s/%s/%s",
		m.url,
		url.PathEscape(dto.Type),
		url.PathEscape(dto.Name),
		url.PathEscape(dto.Value),
	)

	request, err := http.NewRequest(http.MethodPost, path, nil)

	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "text/plain")

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
