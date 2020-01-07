package canaryclient

import (
	"fmt"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/config"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/hattery"
	"time"
)

type CanaryStatus struct {
	Healthy bool `json:"healthy"`
}

type CanaryClient struct {
	host     string
	port     uint
	username string
	password string
}

func NewCanaryClient(host string, port uint, canary config.CanaryConfig) *CanaryClient {
	return &CanaryClient{
		host:     host,
		port:     port,
		username: canary.Username,
		password: canary.Password,
	}
}

func (c *CanaryClient) Status() (bool, error) {
	url := constructURL(c.host, c.port)

	var canaryData CanaryStatus
	err := hattery.Url(url).Timeout(time.Second*10).BasicAuth(c.username, c.password).Fetch(&canaryData)

	return canaryData.Healthy, err
}

func constructURL(host string, port uint) string {
	return fmt.Sprintf("http://%s:%d/api/v1/status", host, port)
}
