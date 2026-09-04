package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/cloudfoundry/mysql-metrics/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		config                    *Config
		configFilepath            string
		tempDir                   string
		host                      string
		password                  string
		username                  string
		instanceId                string
		deployment                string
		jobName                   string
		jobIndex                  string
		jobIP                     string
		metricFrequency           int
		sourceId                  string
		origin                    string
		emitBrokerMetrics         bool
		emitMysqlMetrics          bool
		emitLeaderFollowerMetrics bool
		emitGaleraMetrics         bool
		emitDiskMetrics           bool
		emitBackupMetrics         bool
		heartbeatDatabase         string
		heartbeatTable            string
	)

	BeforeEach(func() {
		config = &Config{}
	})

	Describe("with a fully formed, correct yaml file", func() {
		BeforeEach(func() {
			host = "localhost"
			username = "user"
			password = "secret"
			sourceId = "p-mysql"
			origin = "origin"
			instanceId = "vm-123456"
			deployment = "service-instance-xyz"
			jobName = "mysql"
			jobIndex = "0"
			jobIP = "10.0.0.1"
			metricFrequency = 1
			emitBrokerMetrics = true
			emitMysqlMetrics = true
			emitLeaderFollowerMetrics = true
			emitGaleraMetrics = true
			emitDiskMetrics = true
			emitBackupMetrics = true
			heartbeatDatabase = "someDatabase"
			heartbeatTable = "someTable"

			var err error
			tempDir, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			configFilepath = filepath.Join(tempDir, "metric-config.yml")
			configString := fmt.Sprintf(`{
				"instance_id":"%s",
				"deployment":"%s",
				"job_name":"%s",
				"job_index":"%s",
				"job_ip":"%s",
				"host":"%s",
				"port":6033,
				"username":"%s",
				"password":"%s",
				"metrics_frequency":%d,
				"source_id":"%s",
				"origin": "%s",
				"emit_broker_metrics":%t,
				"emit_mysql_metrics":%t,
				"emit_leader_follower_metrics":%t,
				"emit_galera_metrics":%t,
				"emit_disk_metrics":%t,
				"emit_backup_metrics":%t,
				"heartbeat_database":"%s",
				"heartbeat_table":"%s",
				"loggregator_ca_path":"/var/vcap/jobs/mysql-metrics/certs/loggregator-ca.pem",
				"loggregator_client_cert_path":"/var/vcap/jobs/mysql-metrics/certs/loggregator-client-cert.pem",
				"loggregator_client_key_path":"/var/vcap/jobs/mysql-metrics/certs/loggregator-client-key.pem"
			}`, instanceId, deployment, jobName, jobIndex, jobIP, host, username, password, metricFrequency, sourceId, origin, emitBrokerMetrics, emitMysqlMetrics, emitLeaderFollowerMetrics, emitGaleraMetrics, emitDiskMetrics, emitBackupMetrics, heartbeatDatabase, heartbeatTable)

			err = os.WriteFile(configFilepath, []byte(configString), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reads the config file and applies default otel settings", func() {
			var err error
			err = LoadFromFile(configFilepath, config)

			Expect(err).NotTo(HaveOccurred())
			Expect(config).NotTo(BeNil())
			Expect(config.InstanceID).To(Equal(instanceId))
			Expect(config.Deployment).To(Equal(deployment))
			Expect(config.JobName).To(Equal(jobName))
			Expect(config.JobIndex).To(Equal(jobIndex))
			Expect(config.JobIP).To(Equal(jobIP))
			Expect(config.Host).To(Equal(host))
			Expect(config.Port).To(Equal(6033))
			Expect(config.Username).To(Equal(username))
			Expect(config.Password).To(Equal(password))
			Expect(config.MetricsFrequency).To(Equal(metricFrequency))
			Expect(config.SourceID).To(Equal(sourceId))
			Expect(config.Origin).To(Equal(origin))
			Expect(config.EmitBrokerMetrics).To(Equal(emitBrokerMetrics))
			Expect(config.EmitMysqlMetrics).To(Equal(emitMysqlMetrics))
			Expect(config.EmitLeaderFollowerMetrics).To(Equal(emitLeaderFollowerMetrics))
			Expect(config.EmitGaleraMetrics).To(Equal(emitGaleraMetrics))
			Expect(config.EmitDiskMetrics).To(Equal(emitDiskMetrics))
			Expect(config.EmitBackupMetrics).To(Equal(emitBackupMetrics))
			Expect(config.HeartbeatDatabase).To(Equal(heartbeatDatabase))
			Expect(config.HeartbeatTable).To(Equal(heartbeatTable))
			Expect(config.OtelEndpoint).To(Equal(DefaultOtelEndpoint))
			Expect(config.OtelCAPath).To(Equal(DefaultOtelCAPath))
			Expect(config.OtelCertPath).To(Equal(DefaultOtelCertPath))
			Expect(config.OtelKeyPath).To(Equal(DefaultOtelKeyPath))
			Expect(config.OtelServerName).To(Equal(DefaultOtelServerName))
		})
	})

	Describe("with custom otel configurations", func() {
		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			configFilepath = filepath.Join(tempDir, "metric-config.yml")
			configString := `{
				"otel_endpoint": "custom-endpoint:1234",
				"otel_ca_path": "/custom/ca.crt",
				"otel_cert_path": "/custom/cert.crt",
				"otel_key_path": "/custom/key.key",
				"otel_server_name": "custom-server"
			}`

			err = os.WriteFile(configFilepath, []byte(configString), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("preserves custom otel settings", func() {
			err := LoadFromFile(configFilepath, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.OtelEndpoint).To(Equal("custom-endpoint:1234"))
			Expect(config.OtelCAPath).To(Equal("/custom/ca.crt"))
			Expect(config.OtelCertPath).To(Equal("/custom/cert.crt"))
			Expect(config.OtelKeyPath).To(Equal("/custom/key.key"))
			Expect(config.OtelServerName).To(Equal("custom-server"))
		})
	})

	Describe("when the yaml file is not fully formed", func() {
		BeforeEach(func() {
			configString := `"field1value1"}`
			var err error
			tempDir, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			configFilepath = filepath.Join(tempDir, "metric-config.yml")

			err = os.WriteFile(configFilepath, []byte(configString), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error", func() {
			var err error
			err = LoadFromFile(configFilepath, config)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("when the yaml file does not exist", func() {
		It("returns an error", func() {
			var err error
			err = LoadFromFile("path/doesnot/exist", config)

			Expect(err).To(HaveOccurred())
		})
	})
})
