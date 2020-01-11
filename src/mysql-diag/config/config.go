package config

import (
	"database/sql"
	"io/ioutil"

	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	yaml "gopkg.in/yaml.v2"
	"mysql-diag/msg"
)

type Config struct {
	Canary *CanaryConfig `yaml:"canary"`
	Mysql  MysqlConfig   `yaml:"mysql"`
}

type CanaryConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	ApiPort  uint   `yaml:"api_port"`
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

func (mysqlConfig *MysqlConfig) ConnectionString(node MysqlNode) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=10s", mysqlConfig.Username, mysqlConfig.Password, node.Host, mysqlConfig.Port)
}

func (mysqlConfig *MysqlConfig) Connection(node MysqlNode) *sql.DB {
	conn, err := sql.Open("mysql", mysqlConfig.ConnectionString(node))
	if err != nil {
		msg.PrintfError("Database configuration problem: %v", err)
		os.Exit(1)
	}

	return conn
}

func (c *Config) HostsWithLogs() []string {
	result := []string{}
	for _, node := range c.Mysql.Nodes {
		result = append(result, node.Host)
	}

	return result
}

func LoadFromFile(filepath string) (*Config, error) {
	contents, err := ioutil.ReadFile(filepath)
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
