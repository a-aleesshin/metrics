package grpc

import (
	"context"
	"net"
	"strings"

	pb "github.com/a-aleesshin/metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TrustedSubnetInterceptor(subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		ip := clientIPFromMetadata(ctx)
		if ip == nil || !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "client ip is not in trusted subnet")
		}

		return handler(ctx, req)
	}
}

func clientIPFromMetadata(ctx context.Context) net.IP {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	values := md.Get(pb.RealIPMetadataKey)
	if len(values) == 0 {
		return nil
	}

	return net.ParseIP(strings.TrimSpace(values[0]))
}
