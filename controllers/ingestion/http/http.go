package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// DruidHTTP interface
type DruidHTTP interface {
	Do() (*Response, error)
}

// HTTP client
type Client struct {
	Method     string
	URL        string
	HTTPClient http.Client
	Body       []byte
	Auth       Auth
}

func NewHTTPClient(method, url string, client http.Client, body []byte, auth Auth) DruidHTTP {
	newClient := &Client{
		Method:     method,
		URL:        url,
		HTTPClient: client,
		Body:       body,
		Auth:       auth,
	}

	return newClient
}

// Auth mechanisms supported by Druid control plane to authenticate
// with druid clusters
type Auth struct {
	BasicAuth BasicAuth
}

// BasicAuth
type BasicAuth struct {
	UserName string
	Password string
}

// Response passed to controller
type Response struct {
	ResponseBody string
	StatusCode   int
}

// Do method to be used schema and tenant controller.
func (c *Client) Do() (*Response, error) {

	req, err := http.NewRequest(c.Method, c.URL, bytes.NewBuffer(c.Body))
	if err != nil {
		return nil, err
	}

	if c.Auth.BasicAuth != (BasicAuth{}) {
		req.SetBasicAuth(c.Auth.BasicAuth.UserName, c.Auth.BasicAuth.Password)
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{ResponseBody: string(responseBody), StatusCode: resp.StatusCode}, nil
}
