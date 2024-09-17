package canaryclient

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/hattery"
	"github.com/cloudfoundry/mysql-diag/msg"
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

// Returns true if the canary is unhealthy. Otherwise, it's either healthy or unknown.
func Check(config *config.CanaryConfig) bool {
	if config == nil {
		return false
	}

	intro := "Checking canary status... "
	fmt.Println(intro)

	client := NewCanaryClient("127.0.0.1", config.ApiPort, *config)
	healthy, err := client.Status()
	if err != nil {
		msg.PrintfErrorIntro(intro, "%v", err)
		return false
	} else {
		if healthy {
			fmt.Println(intro + msg.Happy("healthy"))
			return false
		} else {
			fmt.Println(intro + msg.Alert("unhealthy"))
			return true
		}
	}
}
