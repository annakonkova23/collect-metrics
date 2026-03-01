package handler

import (
	"context"
	"fmt"

	mcs "github.com/annakonkova23/collect-metrics/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateMetrics(ctx context.Context, req *mcs.UpdateMetricsRequest) (*mcs.UpdateMetricsResponse, error) {
	s.logger.Info(fmt.Sprintf("Запрос на установку метрик %v", req.GetMetrics()))
	err := s.Collector.SaveMetricsByProto(ctx, req.GetMetrics())
	if err != nil {
		return nil, err
	}
	s.logger.Info("Метрики успешно установлены")
	return &mcs.UpdateMetricsResponse{}, nil
}

func InterceptorCheckIsIPInCIDR(cidr string) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var ip string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				ip = values[0]
				if is, err := IsIPInCIDR(ip, cidr); err != nil {
					return nil, status.Errorf(codes.PermissionDenied, "Ошибка определения принадлежности IP: %v", err)
				} else {
					if !is {
						return nil, status.Errorf(codes.PermissionDenied, "IP %s не принадлежит доверенной сети %s", ip, cidr)
					}
				}
			}
		}
		return handler(ctx, req)
	}
}
