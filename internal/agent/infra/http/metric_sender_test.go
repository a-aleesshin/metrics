package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
)

func TestMetricSender_Send(t *testing.T) {
	tests := []struct {
		name           string
		metric         dto.MetricDTO
		responseStatus int
		wantPath       string
		wantErr        bool
	}{
		{
			name: "send gauge metric",
			metric: dto.MetricDTO{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "123.45",
			},
			responseStatus: http.StatusOK,
			wantPath:       "/update/gauge/Alloc/123.45",
			wantErr:        false,
		},
		{
			name: "send counter metric",
			metric: dto.MetricDTO{
				Type:  "counter",
				Name:  "PollCount",
				Value: "10",
			},
			responseStatus: http.StatusOK,
			wantPath:       "/update/counter/PollCount/10",
			wantErr:        false,
		},
		{
			name: "server returns non-ok status",
			metric: dto.MetricDTO{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "123.45",
			},
			responseStatus: http.StatusBadRequest,
			wantPath:       "/update/gauge/Alloc/123.45",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod string
			var gotPath string
			var gotContentType string

			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotContentType = r.Header.Get("Content-Type")

				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			sender := NewMetricSender(server.URL, server.Client())

			// Act
			err := sender.Send(tt.metric)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodPost {
				t.Fatalf("expected method %q, got %q", http.MethodPost, gotMethod)
			}

			if gotPath != tt.wantPath {
				t.Fatalf("expected path %q, got %q", tt.wantPath, gotPath)
			}

			if gotContentType != "text/plain" {
				t.Fatalf("expected Content-Type %q, got %q", "text/plain", gotContentType)
			}
		})
	}
}
