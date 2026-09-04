package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultOtelEndpoint   = "127.0.0.1:9564"
	DefaultOtelCAPath     = "/var/vcap/jobs/mysql-metrics/certs/loggregator-ca.pem"
	DefaultOtelCertPath   = "/var/vcap/jobs/mysql-metrics/certs/loggregator-client-cert.pem"
	DefaultOtelKeyPath    = "/var/vcap/jobs/mysql-metrics/certs/loggregator-client-key.pem"
	DefaultOtelServerName = "otel-collector"
)

type Config struct {
	MetricsFrequency          int    `yaml:"metrics_frequency"`
	Host                      string `yaml:"host"`
	Port                      int    `yaml:"port"`
	Password                  string `yaml:"password"`
	Username                  string `yaml:"username"`
	InstanceID                string `yaml:"instance_id"`
	Deployment                string `yaml:"deployment"`
	JobName                   string `yaml:"job_name"`
	JobIndex                  string `yaml:"job_index"`
	JobIP                     string `yaml:"job_ip"`
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
	OtelEndpoint              string `yaml:"otel_endpoint"`
	OtelCAPath                string `yaml:"otel_ca_path"`
	OtelCertPath              string `yaml:"otel_cert_path"`
	OtelKeyPath               string `yaml:"otel_key_path"`
	OtelServerName            string `yaml:"otel_server_name"`
}

func LoadFromFile(filepath string, cfg *Config) error {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(contents, cfg); err != nil {
		return err
	}

	cfg.applyDefaults()

	return nil
}

func (c *Config) applyDefaults() {
	if c.OtelEndpoint == "" {
		c.OtelEndpoint = DefaultOtelEndpoint
	}
	if c.OtelCAPath == "" {
		if c.LoggregatorCAPath != "" {
			c.OtelCAPath = c.LoggregatorCAPath
		} else {
			c.OtelCAPath = DefaultOtelCAPath
		}
	}
	if c.OtelCertPath == "" {
		if c.LoggregatorClientCertPath != "" {
			c.OtelCertPath = c.LoggregatorClientCertPath
		} else {
			c.OtelCertPath = DefaultOtelCertPath
		}
	}
	if c.OtelKeyPath == "" {
		if c.LoggregatorClientKeyPath != "" {
			c.OtelKeyPath = c.LoggregatorClientKeyPath
		} else {
			c.OtelKeyPath = DefaultOtelKeyPath
		}
	}
	if c.OtelServerName == "" {
		c.OtelServerName = DefaultOtelServerName
	}
}
