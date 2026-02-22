// Пакет работы с клиентом.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/annakonkova23/collect-metrics/internal/crypto"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Константы для retry.
const (
	countAttempt = 3 //количество попыток
	delayAttempt = 2 //задержка между попытками
)

// Client клиент.
type Client struct {
	client         *resty.Client
	Sugar          *zap.SugaredLogger
	gzipWriterPool sync.Pool
	bufferPool     sync.Pool
	key            *rsa.PublicKey
}

// NewClient создание клиента.
func NewClient(keyPath string, sugar *zap.SugaredLogger) *Client {
	key, err := crypto.ReadPublicKey(keyPath)
	if err != nil {
		sugar.Errorln(err)
	}

	return &Client{
		client: resty.New(),
		Sugar:  sugar,
		gzipWriterPool: sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(nil)
			},
		},
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		key: key,
	}
}

// Post отправляет POST-запрос в формате text/plain.
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

// Post	отправка post запроса в формате application/json с body.
func (c *Client) PostWithBody(ctx context.Context, url string, body []byte, hash string) error {
	delay := 1

	if c.key != nil {
		key, data, err := crypto.HybridEncrypt(c.key, body)
		if err != nil {
			return err
		}
		reqBody := model.EncryptedRequest{
			EncryptedKey:  key,
			EncryptedData: data,
		}

		data, err = json.Marshal(reqBody)
		if err != nil {
			return err
		}
		body = data
	}
	buf := c.bufferPool.Get().(*bytes.Buffer)
	gz := c.gzipWriterPool.Get().(*gzip.Writer)

	buf.Reset()
	gz.Reset(buf)

	if _, err := gz.Write(body); err != nil {
		gz.Close()
		c.bufferPool.Put(buf)
		c.gzipWriterPool.Put(gz)
		return err
	}
	if err := gz.Close(); err != nil {
		c.bufferPool.Put(buf)
		c.gzipWriterPool.Put(gz)
		return err
	}

	compressed := buf.Bytes()

	defer func() {
		c.bufferPool.Put(buf)
		c.gzipWriterPool.Put(gz)
	}()
	c.client.OnAfterResponse(c.WithLoggingResponse)
	var err error
	var response *resty.Response
	ip, err := getLocalIP()
	if err != nil {
		c.Sugar.Error(err)
	}
	for i := 0; i < countAttempt; i++ {
		response, err = c.client.R().
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Length", strconv.Itoa(buf.Len())).
			SetHeader("Accept-Encoding", "gzip").
			SetHeader("HashSHA256", hash).
			SetHeader("X-Real-IP", ip).
			SetBody(compressed).
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

// WithLoggingResponse middleware для логирования.
func (c *Client) WithLoggingResponse(client *resty.Client, response *resty.Response) error {
	c.Sugar.Infoln(
		"status", response.Status(),
		"size", response.Size(),
		"content-type", response.Header().Get("Content-Type"),
		"content-encoding", response.Header().Get("Content-Encoding"),
	)
	return nil
}

// isTemporaryResponseError проверка на временную ошибку.
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

// isConnectionRefused проверка на ошибку соединения
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

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
