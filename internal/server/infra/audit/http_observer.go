package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/a-aleesshin/metrics/internal/server/audit"
)

const defaultTimeout = 5 * time.Second

// HTTPObserver отправляет события аудита POST-запросом с JSON-телом на заданный URL.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver создаёт наблюдателя, отправляющего события на url.
// При client == nil используется http.Client с таймаутом 5 секунд.
func NewHTTPObserver(url string, client *http.Client) *HTTPObserver {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	return &HTTPObserver{url: url, client: client}
}

// Notify отправляет событие POST-запросом с Content-Type: application/json.
// Статус ответа >= 300 считается ошибкой.
func (o *HTTPObserver) Notify(ctx context.Context, event audit.Event) error {
	body, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(body))

	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)

	if err != nil {
		return fmt.Errorf("send audit event: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit server returned status %d", resp.StatusCode)
	}

	return nil
}
