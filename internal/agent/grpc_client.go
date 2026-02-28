package agent

import (
	"context"

	mcs "github.com/annakonkova23/collect-metrics/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func (c *Client) SendRequest(ctx context.Context, host string, metrics []*mcs.Metric) error {
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.Sugar.Errorln("ошибка при установлении соединения с сервером", err)
		return err

	}
	defer conn.Close()
	client := mcs.NewMetricsClient(conn)

	ip, err := getLocalIP()
	if err != nil {
		c.Sugar.Error(err)
	}

	md := metadata.New(map[string]string{"x-real-ip": ip})
	ctx = metadata.NewOutgoingContext(ctx, md)

	request := &mcs.UpdateMetricsRequest{}
	request.SetMetrics(metrics)

	if _, err := client.UpdateMetrics(ctx, request); err != nil {
		c.Sugar.Errorln("ошибка отправки запроса", err)
		return err
	}
	return nil
}
