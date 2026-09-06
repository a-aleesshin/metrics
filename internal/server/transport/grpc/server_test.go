package grpc_test

import (
	"context"
	"net"
	"testing"

	"github.com/a-aleesshin/metrics/internal/platform/grpccreds"
	"github.com/a-aleesshin/metrics/internal/platform/id"
	pb "github.com/a-aleesshin/metrics/internal/proto"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	grpctransport "github.com/a-aleesshin/metrics/internal/server/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func startTestServer(t *testing.T, subnetCIDR string) (pb.MetricsClient, *memory.MemStorage) {
	t.Helper()

	storage := memory.NewMemStorage()
	useCase := usecase.NewUpdatesMetricsUseCase(storage, id.NewUUIDV7Generator())

	var subnet *net.IPNet

	if subnetCIDR != "" {
		_, parsed, err := net.ParseCIDR(subnetCIDR)
		if err != nil {
			t.Fatalf("parse subnet: %v", err)
		}

		subnet = parsed
	}

	certPEM, keyPEM, err := grpccreds.GenerateSelfSigned("bufnet")
	if err != nil {
		t.Fatalf("generate certificate: %v", err)
	}

	serverCreds, err := grpccreds.ServerFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("server creds: %v", err)
	}

	clientCreds, err := grpccreds.ClientFromPEM(certPEM)
	if err != nil {
		t.Fatalf("client creds: %v", err)
	}

	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.UnaryInterceptor(grpctransport.TrustedSubnetInterceptor(subnet)),
	)
	pb.RegisterMetricsServer(server, grpctransport.NewMetricsServer(useCase, nil))

	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return pb.NewMetricsClient(conn), storage
}

func withRealIP(ctx context.Context, ip string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, pb.RealIPMetadataKey, ip)
}

func TestUpdateMetrics_StoresBatch(t *testing.T) {
	client, storage := startTestServer(t, "")
	ctx := t.Context()

	request := pb.UpdateMetricsRequest_builder{Metrics: []*pb.Metric{
		pb.Metric_builder{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 123.45}.Build(),
		pb.Metric_builder{Id: "PollCount", Type: pb.Metric_COUNTER, Delta: 5}.Build(),
	}}.Build()

	if _, err := client.UpdateMetrics(ctx, request); err != nil {
		t.Fatalf("update metrics: %v", err)
	}

	if _, err := client.UpdateMetrics(ctx, request); err != nil {
		t.Fatalf("second update metrics: %v", err)
	}

	value, found, err := storage.FindGaugeByName(ctx, metric.Name("Alloc"))
	if err != nil || !found {
		t.Fatalf("gauge not stored: found=%v err=%v", found, err)
	}

	if value != 123.45 {
		t.Fatalf("expected gauge 123.45, got %v", value)
	}

	delta, found, err := storage.FindCounterByName(ctx, metric.Name("PollCount"))
	if err != nil || !found {
		t.Fatalf("counter not stored: found=%v err=%v", found, err)
	}

	if delta != 10 {
		t.Fatalf("expected counter accumulated to 10, got %d", delta)
	}
}

func TestUpdateMetrics_EmptyBatchIsNoop(t *testing.T) {
	client, _ := startTestServer(t, "")

	if _, err := client.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{}); err != nil {
		t.Fatalf("empty batch should succeed, got %v", err)
	}
}

func TestUpdateMetrics_MissingIDRejected(t *testing.T) {
	client, _ := startTestServer(t, "")

	request := pb.UpdateMetricsRequest_builder{Metrics: []*pb.Metric{
		pb.Metric_builder{Type: pb.Metric_GAUGE, Value: 1}.Build(),
	}}.Build()

	_, err := client.UpdateMetrics(t.Context(), request)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestTrustedSubnet_AllowsIPFromSubnet(t *testing.T) {
	client, storage := startTestServer(t, "10.0.0.0/8")

	ctx := withRealIP(t.Context(), "10.1.2.3")

	request := pb.UpdateMetricsRequest_builder{Metrics: []*pb.Metric{
		pb.Metric_builder{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 1.5}.Build(),
	}}.Build()

	if _, err := client.UpdateMetrics(ctx, request); err != nil {
		t.Fatalf("expected success for trusted ip, got %v", err)
	}

	if _, found, _ := storage.FindGaugeByName(t.Context(), metric.Name("Alloc")); !found {
		t.Fatal("metric was not stored")
	}
}

func TestTrustedSubnet_DeniesIPOutsideSubnet(t *testing.T) {
	client, _ := startTestServer(t, "10.0.0.0/8")

	ctx := withRealIP(t.Context(), "192.168.1.1")

	_, err := client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestTrustedSubnet_DeniesRequestWithoutMetadata(t *testing.T) {
	client, _ := startTestServer(t, "10.0.0.0/8")

	_, err := client.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied without metadata, got %v", err)
	}
}

func TestTrustedSubnet_DisabledAllowsAnyRequest(t *testing.T) {
	client, _ := startTestServer(t, "")

	if _, err := client.UpdateMetrics(t.Context(), &pb.UpdateMetricsRequest{}); err != nil {
		t.Fatalf("expected success without subnet check, got %v", err)
	}
}
