package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{},
	}
}

func (c *Client) Post(url string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	response, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", responseBody)
	}
	return nil
}
