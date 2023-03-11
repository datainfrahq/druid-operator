package http

import (
	"net/http"
	"testing"
)

var (
	url = "https://httpbin.org/get"
)

func TestDo(t *testing.T) {
	client := NewHTTPClient(http.MethodGet, url, http.Client{}, nil)
	resp, err := client.Do()
	if err != nil {
		t.Errorf("HTTP Test Failed [%s]", err.Error())
	}
	t.Logf(string(resp))
}
