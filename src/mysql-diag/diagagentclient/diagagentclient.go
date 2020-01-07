package diagagentclient

import (
	"fmt"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/hattery"
	"time"
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
	host     string
	port     uint
	username string
	password string
}

func NewDiagAgentClient(host string, port uint, username string, password string) *DiagAgentClient {
	return &DiagAgentClient{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (c *DiagAgentClient) Info() (*InfoResponse, error) {
	url := constructURL(c.host, c.port)

	var info InfoResponse

	err := hattery.Url(url).Timeout(time.Second*10).BasicAuth(c.username, c.password).Fetch(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func constructURL(host string, port uint) string {
	return fmt.Sprintf("http://%s:%d/api/v1/info", host, port)
}
