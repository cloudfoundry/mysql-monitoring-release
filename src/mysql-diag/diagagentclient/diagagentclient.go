package diagagentclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/hattery"
)

type DiskInfo struct {
	BytesTotal  uint64 `json:"bytes_total"`
	BytesFree   uint64 `json:"bytes_free"`
	InodesTotal uint64 `json:"inodes_total"`
	InodesFree  uint64 `json:"inodes_free"`
}

type InfoResponse struct {
	Persistent DiskInfo `json:"persistent"`
	Ephemeral  DiskInfo `json:"ephemeral"`
}

type DiagAgentClient struct {
	username   string
	password   string
	useTLS     bool
	httpClient *http.Client
}

func NewDiagAgentClient(agent config.AgentConfig) *DiagAgentClient {
	return &DiagAgentClient{
		username:   agent.Username,
		password:   agent.Password,
		httpClient: agent.TLS.HTTPClient(),
		useTLS:     agent.TLS.Enabled,
	}
}

func (c *DiagAgentClient) Info(address string, useTLS bool) (*InfoResponse, error) {
	url := fmt.Sprintf("http://%s/api/v1/info", address)
	if useTLS {
		url = fmt.Sprintf("https://%s/api/v1/info", address)
	}

	var info InfoResponse

	err := hattery.Url(url).
		Timeout(time.Second*10).
		BasicAuth(c.username, c.password).
		Client(c.httpClient).
		Fetch(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}
