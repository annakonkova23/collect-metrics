package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/http"
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
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
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
	c.client.OnAfterResponse(c.WithLoggingResponse)
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
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
	)
	return nil
}
