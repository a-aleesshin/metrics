package httpadapter

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/platform/hash"
)

const HashSHA256Header = "HashSHA256"

type SigningClient struct {
	client HTTPClient
	key    string
}

func NewSigningClient(client HTTPClient, key string) *SigningClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &SigningClient{
		client: client,
		key:    key,
	}
}

func (c *SigningClient) Do(request *http.Request) (*http.Response, error) {
	if c.key == "" {
		return c.client.Do(request)
	}

	body, err := readRequestBody(request)
	if err != nil {
		return nil, err
	}

	request.Body = io.NopCloser(bytes.NewReader(body))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	request.Header.Set(HashSHA256Header, hash.SumSHA256(body, c.key))

	return c.client.Do(request)
}

func readRequestBody(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	if err := request.Body.Close(); err != nil {
		return nil, fmt.Errorf("close request body: %w", err)
	}

	return body, nil
}
