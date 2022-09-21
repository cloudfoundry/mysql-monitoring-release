package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	MetricsFrequency          int    `yaml:"metrics_frequency"`
	Host                      string `yaml:"host"`
	Port                      int    `yaml:"port"`
	Password                  string `yaml:"password"`
	Username                  string `yaml:"username"`
	InstanceID                string `yaml:"instance_id"`
	Origin                    string `yaml:"origin"`
	SourceID                  string `yaml:"source_id"`
	EmitCPUMetrics            bool   `yaml:"emit_cpu_metrics"`
	EmitMysqlMetrics          bool   `yaml:"emit_mysql_metrics"`
	EmitLeaderFollowerMetrics bool   `yaml:"emit_leader_follower_metrics"`
	EmitGaleraMetrics         bool   `yaml:"emit_galera_metrics"`
	EmitDiskMetrics           bool   `yaml:"emit_disk_metrics"`
	EmitBrokerMetrics         bool   `yaml:"emit_broker_metrics"`
	EmitBackupMetrics         bool   `yaml:"emit_backup_metrics"`
	HeartbeatDatabase         string `yaml:"heartbeat_database"`
	HeartbeatTable            string `yaml:"heartbeat_table"`
	LoggregatorCAPath         string `yaml:"loggregator_ca_path"`
	LoggregatorClientCertPath string `yaml:"loggregator_client_cert_path"`
	LoggregatorClientKeyPath  string `yaml:"loggregator_client_key_path"`
}

func LoadFromFile(filepath string, cfg *Config) error {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(contents, cfg); err != nil {
		return err
	}

	return err
}
