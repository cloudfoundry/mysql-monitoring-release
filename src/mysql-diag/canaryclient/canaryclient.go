package canaryclient

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/hattery"
)

type CanaryStatus struct {
	Healthy bool `json:"healthy"`
}

type CanaryClient struct {
	address    string
	username   string
	password   string
	useTLS     bool
	httpClient *http.Client
}

func NewCanaryClient(host string, port uint, canary config.CanaryConfig) *CanaryClient {
	return &CanaryClient{
		address:    net.JoinHostPort(host, strconv.Itoa(int(port))),
		httpClient: canary.TLS.HTTPClient(),
		useTLS:     canary.TLS.Enabled,
		username:   canary.Username,
		password:   canary.Password,
	}
}

func (c *CanaryClient) Status() (bool, error) {
	url := c.constructURL()

	var canaryData CanaryStatus
	err := hattery.Url(url).
		Timeout(time.Second*10).
		BasicAuth(c.username, c.password).
		Client(c.httpClient).
		Fetch(&canaryData)

	return canaryData.Healthy, err
}

func (c CanaryClient) constructURL() string {
	if c.useTLS {
		return fmt.Sprintf("https://%s/api/v1/status", c.address)
	}
	return fmt.Sprintf("http://%s/api/v1/status", c.address)
}
