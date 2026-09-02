package grpc

import (
	"context"

	pb "github.com/a-aleesshin/metrics/internal/proto"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdatesMetricsUseCase interface {
	Execute(ctx context.Context, command usecase.UpdatesMetricsCommand) error
}

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	useCase UpdatesMetricsUseCase
}

func NewMetricsServer(useCase UpdatesMetricsUseCase) *MetricsServer {
	return &MetricsServer{useCase: useCase}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	commands := make([]usecase.MetricUpdatesCommand, 0, len(req.GetMetrics()))

	for _, metric := range req.GetMetrics() {
		command, err := toCommand(metric)
		if err != nil {
			return nil, err
		}

		commands = append(commands, command)
	}

	if len(commands) == 0 {
		return &pb.UpdateMetricsResponse{}, nil
	}

	if err := s.useCase.Execute(ctx, usecase.UpdatesMetricsCommand{Metrics: commands}); err != nil {
		return nil, status.Errorf(codes.Internal, "update metrics: %v", err)
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func toCommand(metric *pb.Metric) (usecase.MetricUpdatesCommand, error) {
	if metric.GetId() == "" {
		return usecase.MetricUpdatesCommand{}, status.Error(codes.InvalidArgument, "metric id is required")
	}

	switch metric.GetType() {
	case pb.Metric_GAUGE:
		value := metric.GetValue()

		return usecase.MetricUpdatesCommand{
			Name:  metric.GetId(),
			MType: "gauge",
			Value: &value,
		}, nil

	case pb.Metric_COUNTER:
		delta := metric.GetDelta()

		return usecase.MetricUpdatesCommand{
			Name:  metric.GetId(),
			MType: "counter",
			Delta: &delta,
		}, nil

	default:
		return usecase.MetricUpdatesCommand{}, status.Errorf(codes.InvalidArgument, "unsupported metric type: %v", metric.GetType())
	}
}
