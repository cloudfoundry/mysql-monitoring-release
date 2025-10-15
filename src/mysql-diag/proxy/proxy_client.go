package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
)

type Client struct {
	username         string
	password         string
	host             string
	port             int
	backendsEndpoint string
	Name             string
	httpClient       *http.Client
}

func NewProxyClient(config config.Proxy) Client {
	return Client{
		username:         config.Username,
		password:         config.Password,
		host:             config.Host,
		port:             config.Port,
		backendsEndpoint: config.BackendsEndpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (c Client) Backends() (backends []Backend) {
	url := fmt.Sprintf("https://%s:%d/%s", c.host, c.port, c.backendsEndpoint)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("X-Forwarded-Proto", "https")
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(&backends)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Received %v backends\n", backends)

	return backends
}

type Backend struct {
	//Host                string `json:"host"`
	//Port                int    `json:"port"`
	//Healthy             bool   `json:"healthy"`
	Name string `json:"name"`
	//CurrentSessionCount int    `json:"currentSessionCount"`
	Active bool `json:"active"`
	//TrafficEnabled      bool   `json:"trafficEnabled"`
}

func (b Backend) String() string {
	return fmt.Sprintf("Name: %s, Active: %t", b.Name, b.Active)
}
