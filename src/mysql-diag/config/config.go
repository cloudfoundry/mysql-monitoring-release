package config

import (
	"crypto/x509"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/tlsconfig"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"

	"github.com/cloudfoundry/mysql-diag/msg"
)

type Config struct {
	Mysql       MysqlConfig        `yaml:"mysql"`
	GaleraAgent *GaleraAgentConfig `yaml:"galera_agent"`
	Proxies     []Proxy            `yaml:"proxies"`
}

type GaleraAgentConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	ApiPort  uint   `yaml:"api_port"`
	TLS      TLS    `yaml:"tls"`
}

type MysqlConfig struct {
	Username  string           `yaml:"username"`
	Password  string           `yaml:"password"`
	Port      uint             `yaml:"port"`
	Agent     *AgentConfig     `yaml:"agent"`
	Threshold *ThresholdConfig `yaml:"threshold"`
	Nodes     []MysqlNode      `yaml:"nodes"`
}

type AgentConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     uint   `yaml:"port"`
	TLS      TLS    `yaml:"tls"`
}

type MysqlNode struct {
	Host string `yaml:"host"`
	Name string `yaml:"name"`
	UUID string `yaml:"uuid"`
}

type ThresholdConfig struct {
	DiskUsedWarningPercent       uint `yaml:"disk_used_warning_percent"`
	DiskInodesUsedWarningPercent uint `yaml:"disk_inodes_used_warning_percent"`
}

type TLS struct {
	Enabled    bool   `yaml:"enabled"`
	CA         string `yaml:"ca"`
	ServerName string `yaml:"server_name"`
}

type Proxy struct {
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	BackendsEndpoint string `yaml:"backends_endpoint"`
	Name             string `yaml:"name"`
	TLS              TLS    `yaml:"tls"`
}

func (mysqlConfig *MysqlConfig) ConnectionString(node MysqlNode) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=10s&readTimeout=10s&tls=preferred", mysqlConfig.Username, mysqlConfig.Password, node.Host, mysqlConfig.Port)
}

func (mysqlConfig *MysqlConfig) Connection(node MysqlNode) *sql.DB {
	conn, err := sql.Open("mysql", mysqlConfig.ConnectionString(node))
	if err != nil {
		msg.PrintfError("Database configuration problem: %v", err)
		os.Exit(1)
	}

	return conn
}

func (tls *TLS) HTTPClient() *http.Client {
	httpClient := &http.Client{}

	if tls.Enabled {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM([]byte(tls.CA))

		// This call will never fail with an error given the current options
		// If different options are used in the future, we should check the error
		tlsClientConfig, _ := tlsconfig.Build(
			tlsconfig.WithInternalServiceDefaults(),
		).Client(
			tlsconfig.WithAuthority(certPool),
			tlsconfig.WithServerName(tls.ServerName),
		)
		httpClient.Transport = &http.Transport{TLSClientConfig: tlsClientConfig}
	}

	return httpClient
}

func (c *Config) HostsWithLogs() []string {
	var result []string
	for _, node := range c.Mysql.Nodes {
		result = append(result, node.Host)
	}

	return result
}

func LoadFromFile(filepath string) (*Config, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(contents, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
