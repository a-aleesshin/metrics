package httpadapter

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/platform/hash"
)

type signingHTTPClientSpy struct {
	request *http.Request
	body    []byte
}

func (s *signingHTTPClientSpy) Do(request *http.Request) (*http.Response, error) {
	s.request = request

	if request.Body != nil {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		s.body = body
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestSigningClient_Do(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		body       string
		wantHeader bool
	}{
		{
			name:       "sets hash header when key is not empty",
			key:        "secret",
			body:       `{"ok":true}`,
			wantHeader: true,
		},
		{
			name:       "does not set hash header when key is empty",
			key:        "",
			body:       `{"ok":true}`,
			wantHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			spy := &signingHTTPClientSpy{}
			client := NewSigningClient(spy, tt.key)

			request, err := http.NewRequest(
				http.MethodPost,
				"http://example.com/update",
				strings.NewReader(tt.body),
			)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}

			// Act
			_, err = client.Do(request)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotHeader := spy.request.Header.Get(HashSHA256Header)
			if tt.wantHeader {
				wantHeader := hash.SumSHA256([]byte(tt.body), tt.key)
				if gotHeader != wantHeader {
					t.Fatalf("expected hash %q, got %q", wantHeader, gotHeader)
				}
				if spy.request.GetBody == nil {
					t.Fatal("expected GetBody to be set")
				}
			} else if gotHeader != "" {
				t.Fatalf("expected empty hash header, got %q", gotHeader)
			}

			if string(spy.body) != tt.body {
				t.Fatalf("expected body %q, got %q", tt.body, string(spy.body))
			}
		})
	}
}
