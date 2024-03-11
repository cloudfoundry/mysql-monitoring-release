package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

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

var _ = Describe("mysql-metrics", func() {
	var (
		configFilepath  string
		tempDir         string
		password        string
		username        string
		err             error
		session         *gexec.Session
		metricFrequency int
		config          *MetricsConfig
	)

	BeforeEach(func() {
		username = "root"
		password = ""

		metricFrequency = 1
		tempDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())
		configFilepath = filepath.Join(tempDir, "metric-config.yml")

		port, err := strconv.Atoi(mysqlPort)
		Expect(err).NotTo(HaveOccurred())

		config = &MetricsConfig{
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

	})

	JustBeforeEach(func() {
		configBytes, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(configFilepath, configBytes, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err = os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session.Interrupt()).Should(gexec.Exit())
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
})
