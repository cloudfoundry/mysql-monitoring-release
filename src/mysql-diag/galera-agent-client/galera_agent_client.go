package galera_agent_client

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/hattery"
)

type GaleraAgentClient struct {
	address    string
	username   string
	password   string
	useTLS     bool
	httpClient *http.Client
}

type GaleraAgentSequenceNumber struct {
	SequenceNumber string `json:"sequence_number"`
}

func NewGaleraAgentClient(host string, port uint, canary config.GaleraAgentConfig) *GaleraAgentClient {
	return &GaleraAgentClient{
		address:    net.JoinHostPort(host, strconv.Itoa(int(port))),
		httpClient: canary.TLS.HTTPClient(),
		useTLS:     canary.TLS.Enabled,
		username:   canary.Username,
		password:   canary.Password,
	}
}

func (g *GaleraAgentClient) SequenceNumber() (string, error) {
	url := g.constructURL()

	var res GaleraAgentSequenceNumber
	err := hattery.Url(url).
		Timeout(time.Second*10).
		BasicAuth(g.username, g.password).
		Client(g.httpClient).
		Fetch(&res)

	return res.SequenceNumber, err
}

func (c *GaleraAgentClient) constructURL() string {
	if c.useTLS {
		return fmt.Sprintf("https://%s/sequence_number", c.address)
	}
	return fmt.Sprintf("http://%s/sequence_number", c.address)
}
