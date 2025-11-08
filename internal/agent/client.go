package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	countAttempt = 3
	delayAttempt = 2
)

type Client struct {
	client *resty.Client
	Sugar  *zap.SugaredLogger
}

func NewClient(sugar *zap.SugaredLogger) *Client {
	return &Client{
		client: resty.New(),
		Sugar:  sugar,
	}
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (c *Client) Post(url string) error {
	c.client.OnAfterResponse(c.WithLoggingResponse)
	response, err := c.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)

	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("%s", response.Body())
	}
	return nil
}

func (c *Client) PostWithBody(ctx context.Context, url string, body []byte) error {
	delay := 1
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer gz.Close()

	if _, err := gz.Write(body); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	c.client.OnAfterResponse(c.WithLoggingResponse)
	var err error
	var response *resty.Response
	for i := 0; i < countAttempt; i++ {
		response, err = c.client.R().
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Length", strconv.Itoa(buf.Len())).
			SetHeader("Accept-Encoding", "gzip").
			SetBody(buf.Bytes()).
			Post(url)

		if err == nil && response.StatusCode() == http.StatusOK {
			return nil
		}

		if c.isTemporaryResponseError(err, response) {
			c.Sugar.Info(fmt.Sprintf("Попытка [%d] отправки запроса завершилась с ошибкой:%v", i+1, err))
			select {
			case <-ctx.Done():
				c.Sugar.Info("Отмена контекста")
				return nil
			case <-time.After(time.Duration(delay) * time.Second):
				delay = delay + delayAttempt
				continue
			}
		}
	}
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("%s", response.Body())
	}
	return nil
}

func (c *Client) WithLoggingResponse(client *resty.Client, response *resty.Response) error {
	c.Sugar.Infoln(
		"status", response.Status(),
		"size", response.Size(),
		"content-type", response.Header().Get("Content-Type"),
		"content-encoding", response.Header().Get("Content-Encoding"),
	)
	return nil
}

func (c *Client) isTemporaryResponseError(err error, resp *resty.Response) bool {
	if err != nil && resp != nil && resp.StatusCode() >= 500 {
		return true
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	if c.isConnectionRefused(err) {
		return true
	}
	return false
}

func (c *Client) isConnectionRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Err != nil &&
			(strings.Contains(opErr.Err.Error(), "actively refused") ||
				strings.Contains(opErr.Err.Error(), "connection refused")) {
			return true
		}
	}
	return false
}
