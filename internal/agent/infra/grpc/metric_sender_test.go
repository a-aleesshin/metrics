package grpcadapter

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	"github.com/a-aleesshin/metrics/internal/platform/grpccreds"
	pb "github.com/a-aleesshin/metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type captureServer struct {
	pb.UnimplementedMetricsServer

	mu      sync.Mutex
	request *pb.UpdateMetricsRequest
	md      metadata.MD
}

func (s *captureServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.request = req
	s.md, _ = metadata.FromIncomingContext(ctx)

	return &pb.UpdateMetricsResponse{}, nil
}

func (s *captureServer) captured() (*pb.UpdateMetricsRequest, metadata.MD) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.request, s.md
}

func newTestSender(t *testing.T) (*MetricSender, *captureServer) {
	t.Helper()

	capture := &captureServer{}

	certPEM, keyPEM, err := grpccreds.GenerateSelfSigned("bufnet")
	if err != nil {
		t.Fatalf("generate certificate: %v", err)
	}

	serverCreds, err := grpccreds.ServerFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("server creds: %v", err)
	}

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(caPath, certPEM, 0o600); err != nil {
		t.Fatalf("write ca file: %v", err)
	}

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.Creds(serverCreds))
	pb.RegisterMetricsServer(server, capture)

	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	sender, err := NewMetricSender(
		"passthrough:///bufnet",
		caPath,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatalf("create sender: %v", err)
	}
	t.Cleanup(func() { _ = sender.Close() })

	return sender, capture
}

func TestSendBatch_SendsMetricsWithRealIPMetadata(t *testing.T) {
	sender, capture := newTestSender(t)
	sender.localIP = "10.0.0.42"

	batch := []dto.MetricDTO{
		{Type: "gauge", Name: "Alloc", Value: "123.45"},
		{Type: "counter", Name: "PollCount", Value: "5"},
	}

	if err := sender.SendBatch(batch); err != nil {
		t.Fatalf("send batch: %v", err)
	}

	request, md := capture.captured()

	if request == nil || len(request.GetMetrics()) != 2 {
		t.Fatalf("expected 2 metrics in request, got %+v", request)
	}

	gauge := request.GetMetrics()[0]
	if gauge.GetId() != "Alloc" || gauge.GetType() != pb.Metric_GAUGE || gauge.GetValue() != 123.45 {
		t.Fatalf("unexpected gauge: %+v", gauge)
	}

	counter := request.GetMetrics()[1]
	if counter.GetId() != "PollCount" || counter.GetType() != pb.Metric_COUNTER || counter.GetDelta() != 5 {
		t.Fatalf("unexpected counter: %+v", counter)
	}

	ips := md.Get(pb.RealIPMetadataKey)
	if len(ips) != 1 || ips[0] != "10.0.0.42" {
		t.Fatalf("expected x-real-ip metadata 10.0.0.42, got %v", ips)
	}
}

func TestSendBatch_EmptyBatchIsNoop(t *testing.T) {
	sender, capture := newTestSender(t)

	if err := sender.SendBatch(nil); err != nil {
		t.Fatalf("empty batch: %v", err)
	}

	if request, _ := capture.captured(); request != nil {
		t.Fatal("expected no request for empty batch")
	}
}

func TestSendBatch_InvalidValueRejected(t *testing.T) {
	sender, _ := newTestSender(t)

	err := sender.SendBatch([]dto.MetricDTO{{Type: "gauge", Name: "Alloc", Value: "not-a-number"}})
	if err == nil {
		t.Fatal("expected error for invalid gauge value")
	}

	err = sender.SendBatch([]dto.MetricDTO{{Type: "unknown", Name: "X", Value: "1"}})
	if err == nil {
		t.Fatal("expected error for unsupported metric type")
	}
}

func TestSend_SingleMetricGoesAsBatchOfOne(t *testing.T) {
	sender, capture := newTestSender(t)

	if err := sender.Send(dto.MetricDTO{Type: "counter", Name: "PollCount", Value: "7"}); err != nil {
		t.Fatalf("send: %v", err)
	}

	request, _ := capture.captured()
	if request == nil || len(request.GetMetrics()) != 1 {
		t.Fatalf("expected batch of one metric, got %+v", request)
	}
}
