package handler

import (
	"context"
	"fmt"

	mcs "github.com/annakonkova23/collect-metrics/pkg/metrics"
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
