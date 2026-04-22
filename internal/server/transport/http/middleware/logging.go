package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseDataWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func newResponseDataWriter(w http.ResponseWriter) *responseDataWriter {
	return &responseDataWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseDataWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}

	n, err := rw.ResponseWriter.Write(b)

	rw.size += n

	return n, err
}

func (rw *responseDataWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(log *zap.Logger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := newResponseDataWriter(w)

			next.ServeHTTP(rw, r)

			log.Info(
				"http request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", time.Since(start)),
				zap.Int("status", rw.status),
				zap.Int("size", rw.size),
			)
		})
	}
}
