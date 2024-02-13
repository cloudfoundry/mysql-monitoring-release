//
// This is a partial implementation of the Java hattery library:
// https://github.com/stickfigure/hattery
//

package hattery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type basicAuth struct {
	username string
	password string
}

type Request struct {
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

func (r Request) Path(path string) Request {
	// TODO: if path doesn't start with leading /, add it
	return Request{
		url:     r.url + path,
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

	client := &http.Client{}

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

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&into)
	if err != nil {
		return err
	}

	return nil
}
