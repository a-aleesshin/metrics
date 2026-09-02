// Package grpcadapter реализует отправку метрик на сервер по gRPC.
package grpcadapter

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	pb "github.com/a-aleesshin/metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const defaultTimeout = 5 * time.Second

type MetricSender struct {
	conn    *grpc.ClientConn
	client  pb.MetricsClient
	localIP string
	timeout time.Duration
}

func NewMetricSender(address string, opts ...grpc.DialOption) (*MetricSender, error) {
	dialOpts := append(
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		opts...,
	)

	conn, err := grpc.NewClient(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("create grpc client for %s: %w", address, err)
	}

	return &MetricSender{
		conn:    conn,
		client:  pb.NewMetricsClient(conn),
		localIP: localIPFor(address),
		timeout: defaultTimeout,
	}, nil
}

func (s *MetricSender) Close() error {
	return s.conn.Close()
}

func (s *MetricSender) Send(metric dto.MetricDTO) error {
	return s.SendBatch([]dto.MetricDTO{metric})
}

func (s *MetricSender) SendBatch(metrics []dto.MetricDTO) error {
	if len(metrics) == 0 {
		return nil
	}

	payload := make([]*pb.Metric, 0, len(metrics))

	for _, metric := range metrics {
		protoMetric, err := toProtoMetric(metric)
		if err != nil {
			return err
		}

		payload = append(payload, protoMetric)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if s.localIP != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, pb.RealIPMetadataKey, s.localIP)
	}

	if _, err := s.client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{Metrics: payload}); err != nil {
		return fmt.Errorf("send metrics batch: %w", err)
	}

	return nil
}

func toProtoMetric(metric dto.MetricDTO) (*pb.Metric, error) {
	switch metric.Type {
	case "gauge":
		value, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid gauge value %q: %w", metric.Value, err)
		}

		if math.IsNaN(value) || math.IsInf(value, 0) {
			value = 0
		}

		return &pb.Metric{Id: metric.Name, Type: pb.Metric_GAUGE, Value: value}, nil

	case "counter":
		delta, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid counter value %q: %w", metric.Value, err)
		}

		return &pb.Metric{Id: metric.Name, Type: pb.Metric_COUNTER, Delta: delta}, nil

	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metric.Type)
	}
}

func localIPFor(address string) string {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return ""
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}

	return udpAddr.IP.String()
}
