package httpadapter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-aleesshin/metrics/internal/platform/retry"
)

// RetryClient — декоратор HTTPClient, повторяющий запрос при сетевых ошибках.
type RetryClient struct {
	client HTTPClient
	delays []time.Duration
}

// NewRetryClient создаёт RetryClient с задержками по умолчанию;
// при nil client используется http.DefaultClient.
func NewRetryClient(client HTTPClient) *RetryClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &RetryClient{
		client: client,
	}
}

// NewRetryClientWithDelays создаёт RetryClient с заданными задержками между повторами.
func NewRetryClientWithDelays(client HTTPClient, delays []time.Duration) *RetryClient {
	retryClient := NewRetryClient(client)
	retryClient.delays = delays

	return retryClient
}

// Do выполняет запрос с повторами: тело клонируется через GetBody перед каждой
// попыткой. Успешный ответ возвращается вызывающему — закрытие Body на нём.
func (c *RetryClient) Do(request *http.Request) (*http.Response, error) {
	delays := c.delays
	if delays == nil {
		delays = retry.DefaultDelays()
	}

	var lastErr error

	for attempt := 0; attempt <= len(delays); attempt++ {
		if attempt > 0 {
			if waitErr := retry.Wait(request.Context(), delays[attempt-1]); waitErr != nil {
				return nil, waitErr
			}
		}

		retryRequest, err := cloneRequest(request)
		if err != nil {
			return nil, err
		}

		response, err := c.client.Do(retryRequest)
		if err == nil {
			return response, nil
		}

		if !isRetriableHTTPError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func cloneRequest(request *http.Request) (*http.Request, error) {
	retryRequest := request.Clone(request.Context())

	if request.Body == nil {
		return retryRequest, nil
	}

	if request.GetBody == nil {
		return nil, fmt.Errorf("request body cannot be retried")
	}

	body, err := request.GetBody()
	if err != nil {
		return nil, fmt.Errorf("get request body: %w", err)
	}

	retryRequest.Body = body

	return retryRequest, nil
}
