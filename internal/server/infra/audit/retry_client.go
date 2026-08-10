package audit

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/a-aleesshin/metrics/internal/platform/retry"
)

// httpDoer — минимальный контракт HTTP-клиента; ему удовлетворяет *http.Client.
type httpDoer interface {
	Do(request *http.Request) (*http.Response, error)
}

// retryHTTPClient — прозрачная обёртка над HTTP-клиентом: повторяет запрос
type retryHTTPClient struct {
	client httpDoer
	delays []time.Duration // nil — задержки по умолчанию (1s, 3s, 5s)
}

func newRetryHTTPClient(client httpDoer, delays []time.Duration) *retryHTTPClient {
	return &retryHTTPClient{client: client, delays: delays}
}

// retriableStatusError — ответ 5xx: запрос имеет смысл повторить.
type retriableStatusError struct {
	code int
}

func (e retriableStatusError) Error() string {
	return fmt.Sprintf("audit server returned status %d", e.code)
}

// permanentError помечает ошибки, при которых повторы бессмысленны.
type permanentError struct {
	err error
}

func (e permanentError) Error() string { return e.err.Error() }

func (e permanentError) Unwrap() error { return e.err }

// Do выполняет запрос с повторами; успешным считается любой ответ со статусом < 500.
func (c *retryHTTPClient) Do(request *http.Request) (*http.Response, error) {
	var response *http.Response

	operation := func() error {
		attemptRequest, err := cloneRequest(request)
		if err != nil {
			return permanentError{err: err}
		}

		resp, err := c.client.Do(attemptRequest)
		if err != nil {
			return err
		}

		if resp.StatusCode >= http.StatusInternalServerError {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			return retriableStatusError{code: resp.StatusCode}
		}

		response = resp

		return nil
	}

	isRetriable := func(err error) bool {
		var permanent permanentError
		return !errors.As(err, &permanent)
	}

	var err error
	if c.delays == nil {
		err = retry.Do(request.Context(), isRetriable, operation)
	} else {
		err = retry.DoWithDelays(request.Context(), c.delays, isRetriable, operation)
	}

	if err != nil {
		var permanent permanentError
		if errors.As(err, &permanent) {
			return nil, permanent.err
		}

		return nil, err
	}

	return response, nil
}

func cloneRequest(request *http.Request) (*http.Request, error) {
	clone := request.Clone(request.Context())

	if request.Body == nil {
		return clone, nil
	}

	if request.GetBody == nil {
		return nil, errors.New("request body cannot be retried")
	}

	body, err := request.GetBody()
	if err != nil {
		return nil, fmt.Errorf("get request body: %w", err)
	}

	clone.Body = body

	return clone, nil
}
