package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Method     string
	URL        string
	HTTPClient http.Client
	Body       []byte
}

func NewHTTPClient(method, url string, client http.Client, body []byte) *Client {
	return &Client{
		Method:     method,
		URL:        url,
		HTTPClient: client,
		Body:       body,
	}
}

func (c *Client) Do() (respBody []byte, err error) {

	req, err := http.NewRequest(c.Method, c.URL, bytes.NewBuffer(c.Body))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(responseBody)

	return respBody, nil

}
