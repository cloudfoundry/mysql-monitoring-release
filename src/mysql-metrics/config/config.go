package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	MetricsFrequency          int    `yaml:"metrics_frequency"`
	Origin                    string `yaml:"origin"`
	Host                      string `yaml:"host"`
	Password                  string `yaml:"password"`
	Username                  string `yaml:"username"`
	InstanceId                string `yaml:"instance_id"`
	EmitCPUMetrics            bool   `yaml:"emit_cpu_metrics"`
	EmitMysqlMetrics          bool   `yaml:"emit_mysql_metrics"`
	EmitLeaderFollowerMetrics bool   `yaml:"emit_leader_follower_metrics"`
	EmitGaleraMetrics         bool   `yaml:"emit_galera_metrics"`
	EmitDiskMetrics           bool   `yaml:"emit_disk_metrics"`
	EmitBrokerMetrics         bool   `yaml:"emit_broker_metrics"`
	HeartbeatDatabase         string `yaml:"heartbeat_database"`
	HeartbeatTable            string `yaml:"heartbeat_table"`
}

func LoadFromFile(filepath string, object interface{}) error {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(contents, object); err != nil {
		return err
	}

	return err
}
