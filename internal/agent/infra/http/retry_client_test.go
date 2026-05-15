package httpadapter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type httpClientStub struct {
	responses []*http.Response
	errors    []error
	calls     int
}

func (c *httpClientStub) Do(request *http.Request) (*http.Response, error) {
	index := c.calls
	c.calls++

	if index < len(c.errors) && c.errors[index] != nil {
		return nil, c.errors[index]
	}

	if index < len(c.responses) && c.responses[index] != nil {
		return c.responses[index], nil
	}

	return responseWithStatus(http.StatusOK), nil
}

func responseWithStatus(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func TestRetryClient_Do(t *testing.T) {
	temporaryErr := errors.New("temporary error")
	permanentErr := unexpectedStatusError{code: http.StatusBadRequest}

	tests := []struct {
		name      string
		client    *httpClientStub
		wantErr   error
		wantCalls int
	}{
		{
			name: "success first try",
			client: &httpClientStub{
				responses: []*http.Response{responseWithStatus(http.StatusOK)},
			},
			wantCalls: 1,
		},
		{
			name: "success after retriable error",
			client: &httpClientStub{
				errors:    []error{temporaryErr, nil},
				responses: []*http.Response{nil, responseWithStatus(http.StatusOK)},
			},
			wantCalls: 2,
		},
		{
			name: "non retriable error is not retried",
			client: &httpClientStub{
				errors: []error{permanentErr},
			},
			wantErr:   permanentErr,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			client := NewRetryClientWithDelays(
				tt.client,
				[]time.Duration{time.Nanosecond, time.Nanosecond, time.Nanosecond},
			)

			request, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"http://example.com/update",
				nil,
			)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}

			// Act
			response, err := client.Do(request)

			// Assert
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}

				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr == nil && response == nil {
				t.Fatal("expected response, got nil")
			}

			if tt.client.calls != tt.wantCalls {
				t.Fatalf("expected calls=%d, got %d", tt.wantCalls, tt.client.calls)
			}
		})
	}
}

func TestRetryClient_Do_ContextCanceled(t *testing.T) {
	temporaryErr := errors.New("temporary error")

	// Arrange
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://example.com/update",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	stub := &httpClientStub{
		errors: []error{temporaryErr},
	}
	client := NewRetryClientWithDelays(stub, []time.Duration{time.Second})

	// Act
	response, err := client.Do(request)

	// Assert
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error %v, got %v", context.Canceled, err)
	}

	if response != nil {
		t.Fatalf("expected nil response, got %+v", response)
	}

	if stub.calls != 1 {
		t.Fatalf("expected calls=1, got %d", stub.calls)
	}
}
