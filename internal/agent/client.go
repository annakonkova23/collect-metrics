package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
)

type Client struct {
	client *resty.Client
}

func NewClient() *Client {
	return &Client{
		client: resty.New(),
	}
}

func (c *Client) Post(url string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
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
