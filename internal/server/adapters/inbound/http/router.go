package http

import (
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/adapters/inbound/http/metrics"
)

func NewRouter(metricsHandler *metrics.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/", metricsHandler.Update)

	return mux
}
