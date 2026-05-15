package httpadapter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-aleesshin/metrics/internal/platform/retry"
)

type RetryClient struct {
	client HTTPClient
	delays []time.Duration
}

func NewRetryClient(client HTTPClient) *RetryClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &RetryClient{
		client: client,
	}
}

func NewRetryClientWithDelays(client HTTPClient, delays []time.Duration) *RetryClient {
	retryClient := NewRetryClient(client)
	retryClient.delays = delays

	return retryClient
}

func (c *RetryClient) Do(request *http.Request) (*http.Response, error) {
	var response *http.Response

	operation := func() error {
		retryRequest, err := cloneRequest(request)
		if err != nil {
			return err
		}

		response, err = c.client.Do(retryRequest)
		if err != nil {
			return err
		}

		return nil
	}

	var err error
	if c.delays == nil {
		err = retry.Do(request.Context(), isRetriableHTTPError, operation)
	} else {
		err = retry.DoWithDelays(request.Context(), c.delays, isRetriableHTTPError, operation)
	}

	if err != nil {
		return nil, err
	}

	return response, nil
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
