package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/a-aleesshin/metrics/internal/platform/hash"
)

const HashSHA256Header = "HashSHA256"

var hashBufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func WithHashSHA256(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqBuf := hashBufferPool.Get().(*bytes.Buffer)
			reqBuf.Reset()
			defer hashBufferPool.Put(reqBuf)

			if _, err := reqBuf.ReadFrom(r.Body); err != nil {
				http.Error(w, "read request body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			body := reqBuf.Bytes()

			if got := r.Header.Get(HashSHA256Header); got != "" && !hash.VerifySHA256(body, key, got) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			rec := newHashResponseRecorder(w)
			defer func() {
				rec.body.Reset()
				hashBufferPool.Put(rec.body)
			}()

			next.ServeHTTP(rec, r)

			for headerName, values := range rec.header {
				for _, value := range values {
					w.Header().Add(headerName, value)
				}
			}

			w.Header().Set(HashSHA256Header, hash.SumSHA256(rec.body.Bytes(), key))
			w.WriteHeader(rec.statusCode)
			w.Write(rec.body.Bytes())
		})
	}
}

type hashResponseRecorder struct {
	writer     http.ResponseWriter
	header     http.Header
	body       *bytes.Buffer
	statusCode int
}

func newHashResponseRecorder(w http.ResponseWriter) *hashResponseRecorder {
	body := hashBufferPool.Get().(*bytes.Buffer)
	body.Reset()

	return &hashResponseRecorder{
		writer:     w,
		header:     make(http.Header),
		body:       body,
		statusCode: http.StatusOK,
	}
}

func (r *hashResponseRecorder) Header() http.Header {
	return r.header
}

func (r *hashResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *hashResponseRecorder) Write(data []byte) (int, error) {
	return r.body.Write(data)
}
