package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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

func (c *Client) PostWithBody(url string, body []byte) error {

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
	response, err := c.client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Length", strconv.Itoa(buf.Len())).
		SetHeader("Accept-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post(url)

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
		"status", response.Status(), // получаем перехваченный код статуса ответа
		"size", response.Size(), // получаем перехваченный размер ответа
		"content-type", response.Header().Get("Content-Type"),
		"content-encoding", response.Header().Get("Content-Encoding"),
	)
	return nil
}
