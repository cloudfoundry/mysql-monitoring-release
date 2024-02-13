//
// This is a partial implementation of the Java hattery library:
// https://github.com/stickfigure/hattery
//

package hattery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type basicAuth struct {
	username string
	password string
}

type Request struct {
	client  *http.Client
	url     string
	timeout time.Duration
	headers map[string]string
	auth    *basicAuth
}

func Url(url string) Request {
	return Request{}.Url(url)
}

func (r Request) Url(url string) Request {
	return Request{
		url:     url,
		timeout: r.timeout,
		headers: r.headers,
		auth:    r.auth,
	}
}

func (r Request) Client(c *http.Client) Request {
	return Request{
		client:  c,
		url:     r.url,
		timeout: r.timeout,
		headers: r.headers,
		auth:    r.auth,
	}
}

func (r Request) Timeout(timeout time.Duration) Request {
	return Request{
		url:     r.url,
		timeout: timeout,
		headers: r.headers,
		auth:    r.auth,
	}
}

func (r Request) BasicAuth(username string, password string) Request {
	return Request{
		url:     r.url,
		timeout: r.timeout,
		headers: r.headers,
		auth:    &basicAuth{username: username, password: password},
	}
}

func (r Request) Fetch(into interface{}) error {
	req, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		return err
	}

	client := r.client
	if r.client == nil {
		client = &http.Client{}
	}

	if r.timeout != 0 {
		client.Timeout = r.timeout
	}

	if r.auth != nil {
		req.SetBasicAuth(r.auth.username, r.auth.password)
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("Unexpected status code: %d (%s)", response.StatusCode, bytes.TrimSpace(body))
	}

	if err = json.NewDecoder(response.Body).Decode(&into); err != nil {
		return err
	}

	return nil
}
