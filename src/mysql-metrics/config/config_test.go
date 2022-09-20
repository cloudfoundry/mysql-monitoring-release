package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/cloudfoundry/mysql-metrics/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		config                    *Config
		configFilepath            string
		tempDir                   string
		host                      string
		port                      int
		password                  string
		username                  string
		instanceId                string
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
			port = 5678
			username = "user"
			password = "secret"
			sourceId = "p-mysql"
			origin = "origin"
			instanceId = "vm-123456"
			metricFrequency = 1
			emitBrokerMetrics = true
			emitMysqlMetrics = true
			emitLeaderFollowerMetrics = true
			emitGaleraMetrics = true
			emitDiskMetrics = true
			emitBackupMetrics = true
			heartbeatDatabase = "someDatabase"
			heartbeatTable = "someTable"

			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			configFilepath = filepath.Join(tempDir, "metric-config.yml")
			configString := fmt.Sprintf(`{
				"instance_id":"%s",
				"host":"%s",
				"port":%d,
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
				"heartbeat_table":"%s"
			}`, instanceId, host, port, username, password, metricFrequency, sourceId, origin, emitBrokerMetrics, emitMysqlMetrics, emitLeaderFollowerMetrics, emitGaleraMetrics, emitDiskMetrics, emitBackupMetrics, heartbeatDatabase, heartbeatTable)

			err = ioutil.WriteFile(configFilepath, []byte(configString), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reads the config file", func() {
			var err error
			err = LoadFromFile(configFilepath, config)

			Expect(err).NotTo(HaveOccurred())
			Expect(config).NotTo(BeNil())
			Expect(config.InstanceID).To(Equal(instanceId))
			Expect(config.Host).To(Equal(host))
			Expect(config.Port).To(Equal(5678))
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
		})
	})

	Describe("when the yaml file is not fully formed", func() {
		BeforeEach(func() {
			configString := `"field1value1"}`
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			configFilepath = filepath.Join(tempDir, "metric-config.yml")

			err = ioutil.WriteFile(configFilepath, []byte(configString), os.ModePerm)
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
