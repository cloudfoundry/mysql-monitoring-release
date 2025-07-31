package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/mysql-metrics/internal/testing/docker"
)

func TestMysqlMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite", Label("integration"))
}

type MetricsConfig struct {
	Host                      string `json:"host"`
	Port                      int    `json:"port"`
	Username                  string `json:"username"`
	Password                  string `json:"password"`
	MetricsFrequency          int    `json:"metrics_frequency"`
	SourceID                  string `json:"source_id"`
	EmitLeaderFollowerMetrics bool   `json:"emit_leader_follower_metrics"`
	EmitMySQLMetrics          bool   `json:"emit_mysql_metrics"`
	EmitGaleraMetrics         bool   `json:"emit_galera_metrics"`
	EmitDiskMetrics           bool   `json:"emit_disk_metrics"`
	EmitCPUMetrics            bool   `json:"emit_cpu_metrics"`
	LoggregatorCAPath         string `json:"loggregator_ca_path"`
	LoggregatorClientCertPath string `json:"loggregator_client_cert_path"`
	LoggregatorClientKeyPath  string `json:"loggregator_client_key_path"`
}

var _ = DescribeTableSubtree("mysql-metrics", Ordered, func(mysqlVersion string) {
	var (
		metricsBinPath  string
		resource        string
		mysqlPort       string
		configFilepath  string
		tempDir         string
		password        string
		username        string
		session         *gexec.Session
		metricFrequency int
	)

	BeforeAll(func() {
		var err error
		metricsBinPath, err = gexec.Build("github.com/cloudfoundry/mysql-metrics", "-race")
		Expect(err).ShouldNot(HaveOccurred())

		DeferCleanup(func() {
			gexec.Kill()
			gexec.CleanupBuildArtifacts()
		})

		resource, err = docker.RunContainer(docker.ContainerSpec{
			Image: "percona/percona-server:" + mysqlVersion,
			Ports: []string{"3306/tcp"},
			Env:   []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1"},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(docker.RemoveContainer(resource)).To(Succeed())
		})

		mysqlPort, err = docker.ContainerPort(resource, "3306/tcp")
		Expect(err).NotTo(HaveOccurred())

		db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(localhost:%s)/", mysqlPort))
		Expect(err).NotTo(HaveOccurred())
		Eventually(db.Ping, "5m", "1s").Should(Succeed())
		Expect(db.Exec(`CHANGE REPLICATION SOURCE TO SOURCE_HOST = 'some-host', SOURCE_USER = 'some-user', SOURCE_PASSWORD = 'some-password'`)).
			Error().NotTo(HaveOccurred())
		Expect(db.Exec(`START REPLICA`)).
			Error().NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		username = "root"
		password = ""

		metricFrequency = 1
		var err error
		tempDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})
		configFilepath = filepath.Join(tempDir, "metric-config.yml")

		port, err := strconv.Atoi(mysqlPort)
		Expect(err).NotTo(HaveOccurred())

		config := &MetricsConfig{
			Host:                      "localhost",
			Port:                      port,
			Username:                  username,
			Password:                  password,
			MetricsFrequency:          metricFrequency,
			SourceID:                  "my_custom_sourceid",
			EmitLeaderFollowerMetrics: true,
			EmitMySQLMetrics:          true,
			EmitGaleraMetrics:         true,
			EmitDiskMetrics:           true,
			EmitCPUMetrics:            true,
			LoggregatorCAPath:         "../fixtures/certs/loggregator.crt",
			LoggregatorClientCertPath: "../fixtures/certs/loggregator-agent.crt",
			LoggregatorClientKeyPath:  "../fixtures/certs/loggregator-agent.key",
		}

		configBytes, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(configFilepath, configBytes, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	runMainWithArgs := func(args ...string) {
		args = append(
			args,
			"-c", configFilepath,
		)

		_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", args)
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(metricsBinPath, args...)
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	}

	It("is able to run the binary without exit", func() {
		runMainWithArgs()

		Consistently(session.ExitCode, 2*time.Second).Should(Equal(-1))
	})

	Describe("when logging is enabled", func() {
		It("emits multiple log entries", func() {
			logFile, err := os.CreateTemp(tempDir, "metrics_via_tcp_loopback_")
			Expect(err).NotTo(HaveOccurred())
			_ = logFile.Close()
			logFilePath := logFile.Name()

			runMainWithArgs("-l", logFilePath)
			Consistently(session.ExitCode, time.Duration(metricFrequency)*time.Second).Should(Equal(-1))

			contents, err := os.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())

			contentsAsString := fmt.Sprintf("%s", contents)
			Expect(contentsAsString).To(ContainSubstring("innodb/buffer_pool_pages_free"))

			firstLineCount := bytes.Count(contents, []byte("\n"))
			Expect(firstLineCount).Should(BeNumerically(">", 0))

			Consistently(session.ExitCode, time.Duration(metricFrequency)*time.Second).Should(Equal(-1))
			contents, err = os.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())
			newLineCount := bytes.Count(contents, []byte("\n"))
			Expect(newLineCount).Should(BeNumerically(">", firstLineCount))
		})

		It("logs that it has metrics", func() {
			logFilePath := filepath.Join(tempDir, "metrics.log")
			runMainWithArgs("-l", logFilePath)
			Consistently(session.ExitCode, time.Duration(metricFrequency*5)*time.Second).Should(Equal(-1))

			contents, err := os.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())

			contentsAsString := fmt.Sprintf("%s", contents)
			Expect(contentsAsString).To(MatchRegexp(`follower/seconds_behind_master","value":\d+`), `metrics output:\n%s`, contentsAsString)
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_used\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_free\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_inodes_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_inodes_used\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_inodes_free\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_used\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_free\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_inodes_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_inodes_used\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_inodes_free\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("performance/queries\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("performance/queries_delta\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("performance/cpu_utilization_percent"))
			Expect(contentsAsString).To(MatchRegexp("variables/read_only\",\"value\":[0,1],"))
		})
	})

	Describe("when the database is unreachable", func() {
		It("still logs that it has emitted the available=0 metric", func() {
			configFilepath = "../fixtures/bad-credentials-fixtures.yml"
			logFilePath := filepath.Join(tempDir, "metrics.log")
			runMainWithArgs("-l", logFilePath)

			Consistently(session.ExitCode, time.Duration(metricFrequency)*time.Second).Should(Equal(-1))
			contents, err := os.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())
			contentsAsString := fmt.Sprintf("%s", contents)
			Expect(contentsAsString).To(ContainSubstring(`"key":"available","value":0`))
			Expect(contentsAsString).NotTo(ContainSubstring("innodb/buffer_pool_pages_free"))
		})
	})
},
	Entry("MySQL 8.0", "8.0"),
	Entry("MySQL 8.4", "8.4"),
)
