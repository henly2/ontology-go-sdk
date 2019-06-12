package httpclient

import (
	"net/http"
	"time"
	"bytes"
	"io/ioutil"
	"io"
)

type (
	HttpClient struct{
		BaseUrl string

		HTTP http.Client
	}

	HandleRequest func(req *http.Request)
)

func NewHttpClient(baseUrl string) *HttpClient {
	exaClient := &HttpClient{
		BaseUrl: baseUrl,
	}
	exaClient.HTTP = http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		ResponseHeaderTimeout: time.Second * 15,
	}}

	return exaClient
}

func (c *HttpClient)Do(method, path string, query, body string, cb HandleRequest) (int, []byte, error) {
	var (
		url string
		reader io.Reader
	)
	url = c.BaseUrl +path
	if query != ""{
		url += "?"
		url += query
	}

	if body != ""{
		reader = bytes.NewReader([]byte(body))
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil{
		return -1, nil, err
	}

	if cb != nil {
		cb(req)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, data, err
}

func (c *HttpClient)Post(path string, query, body string, cb HandleRequest) (int, []byte, error) {
	return c.Do("POST", path, query, body, cb)
}

func (c *HttpClient)Get(path string, query string, cb HandleRequest) (int, []byte, error) {
	return c.Do("GET", path, query, "", cb)
}

func (c *HttpClient)Put(path string, query, body string, cb HandleRequest) (int, []byte, error) {
	return c.Do("PUT", path, query, body, cb)
}

func (c *HttpClient)Delete(path string, query string, cb HandleRequest) (int, []byte, error) {
	return c.Do("DELETE", path, query, "", cb)
}