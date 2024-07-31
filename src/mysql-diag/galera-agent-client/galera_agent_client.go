package galera_agent_client

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/hattery"
	"github.com/cloudfoundry/mysql-diag/msg"
)

const DEFAULT_SEQNO = -1

type GaleraAgentClient struct {
	address    string
	username   string
	password   string
	useTLS     bool
	httpClient *http.Client
}

func NewGaleraAgentClient(host string, galeraAgent config.GaleraAgentConfig) *GaleraAgentClient {
	return &GaleraAgentClient{
		address:    net.JoinHostPort(host, strconv.Itoa(int(galeraAgent.ApiPort))),
		httpClient: galeraAgent.TLS.HTTPClient(),
		useTLS:     galeraAgent.TLS.Enabled,
		username:   galeraAgent.Username,
		password:   galeraAgent.Password,
	}
}

func (g *GaleraAgentClient) SequenceNumber() (int, error) {
	url := g.constructURL()

	var seqNo int
	err := hattery.Url(url).
		Timeout(time.Second*30).
		BasicAuth(g.username, g.password).
		Client(g.httpClient).
		Fetch(&seqNo)

	return seqNo, err
}

func GetSequenceNumbers(galeraAgentConfig *config.GaleraAgentConfig, nodeClusterStatus map[string]*database.NodeClusterStatus) {
	if galeraAgentConfig == nil {
		fmt.Println("Galera Agent not configured, skipping sequence number check")
		return
	}
	channel := make(chan database.NodeClusterStatus, len(nodeClusterStatus))
	channelCount := 0
	for _, status := range nodeClusterStatus {
		n := status.Node
		s := status.Status
		if s != nil {
			continue
		}
		channelCount += 1
		go func() {
			s = &database.GaleraStatus{LastApplied: DEFAULT_SEQNO}
			agentClient := NewGaleraAgentClient(n.Host, *galeraAgentConfig)
			sequenceNumber, err := agentClient.SequenceNumber()
			if err != nil {
				msg.PrintfErrorIntro("", "error retrieving galera agent sequence number: %v", err)
			} else {
				s.LastApplied = sequenceNumber
			}
			channel <- database.NodeClusterStatus{Node: n, Status: s}
		}()
	}

	for i := 0; i < channelCount; i++ {
		ns := <-channel
		nodeClusterStatus[ns.Node.Host] = &ns
	}
}

func (c *GaleraAgentClient) constructURL() string {
	if c.useTLS {
		return fmt.Sprintf("https://%s/sequence_number", c.address)
	}
	return fmt.Sprintf("http://%s/sequence_number", c.address)
}
