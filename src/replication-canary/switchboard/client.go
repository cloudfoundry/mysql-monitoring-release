package switchboard

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/lager"
)

type Client struct {
	client   *http.Client
	rootURL  string
	username string
	password string
	logger   lager.Logger
}

func NewClient(
	rootURL string,
	username string,
	password string,
	skipSSLCertVerify bool,
	logger lager.Logger,
) *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 5,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipSSLCertVerify,
				},
			},
		},
		rootURL:  rootURL,
		username: username,
		password: password,
		logger:   logger,
	}
}

type backends []backend

type backend struct {
	Host   string `json:"host"`
	Active bool   `json:"active"`
}

func (c *Client) DisableClusterTraffic() error {
	message := "Disabling cluster traffic"

	return c.sendClusterTrafficRequest(enable(false), message)
}

func (c *Client) EnableClusterTraffic() error {
	message := "Enabling cluster traffic"

	return c.sendClusterTrafficRequest(enable(true), message)
}

func (c *Client) sendClusterTrafficRequest(enabled enable, message string) error {
	u := c.rootURL + "/v0/cluster"

	req, err := http.NewRequest("PATCH", u, nil)
	if err != nil {
		return err
	}

	v := url.Values{}
	v.Add("trafficEnabled", fmt.Sprintf("%v", enabled))
	v.Add("message", message)
	req.URL.RawQuery = v.Encode()

	req.SetBasicAuth(c.username, c.password)

	c.logger.Debug("Making request to proxy", lager.Data{
		"method":         "PATCH",
		"url":            req.URL,
		"trafficEnabled": enabled,
		"message":        message,
	})
	res, err := c.client.Do(req)
	if err != nil {
		c.logger.Debug("error making request to proxy", lager.Data{
			"errorMessage": err.Error(),
		})
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		var action string
		if enabled {
			action = "enabling cluster"
		} else {
			action = "disabling cluster"
		}

		b, _ := ioutil.ReadAll(res.Body)

		c.logger.Debug("received bad status code from proxy", lager.Data{
			"statusCode": res.StatusCode,
		})
		return fmt.Errorf("bad response %s (%d) - %s", action, res.StatusCode, string(b))
	}

	return nil
}

type enable bool

func (c *Client) ActiveBackendHost() (string, error) {
	u := c.rootURL + "/v0/backends"

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.username, c.password)

	// The proxy issues a redirect to https if we're missing this header. That's to keep public requests
	// through the router from passing basic auth creds over http. Since we're also on the internal network,
	// we can just pretend to be the router.
	req.Header.Set("X-Forwarded-Proto", "https")

	c.logger.Debug("Making request to proxy", lager.Data{
		"method": "GET",
		"url":    req.URL,
	})
	res, err := c.client.Do(req)
	if err != nil {
		c.logger.Debug("error making request to proxy", lager.Data{
			"errorMessage": err.Error(),
		})
		return "", err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// Untested as it is too hard to force ioutil.ReadAll to return an error
		c.logger.Debug("error reading proxy response body", lager.Data{
			"errorMessage": err.Error(),
		})
		return "", err
	}

	if res.StatusCode >= http.StatusBadRequest {
		c.logger.Debug("received bad status code from proxy", lager.Data{
			"statusCode": res.StatusCode,
		})
		return "", fmt.Errorf("bad response (%d) - %s", res.StatusCode, string(b))
	}

	var resp backends
	err = json.Unmarshal(b, &resp)
	if err != nil {
		c.logger.Debug("error unmarshalling json proxy response body", lager.Data{
			"errorMessage": err.Error(),
		})
		return "", err
	}

	for _, backend := range resp {
		if backend.Active {
			return backend.Host, nil
		}
	}

	return "", fmt.Errorf("no active backend found")
}
