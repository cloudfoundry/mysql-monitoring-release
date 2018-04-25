package integration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type MetricsConfig struct {
	Host                      string `json:"host"`
	Username                  string `json:"username"`
	Password                  string `json:"password"`
	MetricsFrequency          int    `json:"metrics_frequency"`
	Origin                    string `json:"origin"`
	EmitLeaderFollowerMetrics bool   `json:"emit_leader_follower_metrics"`
	EmitMySQLMetrics          bool   `json:"emit_mysql_metrics"`
	EmitGaleraMetrics         bool   `json:"emit_galera_metrics"`
	EmitDiskMetrics           bool   `json:"emit_disk_metrics"`
	EmitCPUMetrics            bool   `json:"emit_cpu_metrics"`
}

var _ = Describe("mysql-metrics", func() {
	var (
		configFilepath  string
		tempDir         string
		password        string
		username        string
		socket          string
		err             error
		session         *gexec.Session
		metricFrequency int
		config          *MetricsConfig
	)

	BeforeEach(func() {
		var unsetVars []string
		if env, ok := os.LookupEnv("MYSQL_USER"); ok {
			username = env
		} else {
			unsetVars = append(unsetVars, "MYSQL_USER")
		}
		if env, ok := os.LookupEnv("MYSQL_PASSWORD"); ok {
			password = env
		} else {
			unsetVars = append(unsetVars, "MYSQL_PASSWORD")
		}
		if env, ok := os.LookupEnv("MYSQL_SOCKET"); ok {
			socket = env
		} else {
			socket = "/tmp/mysql.sock"
		}

		if len(unsetVars) > 0 {
			panic(fmt.Sprintf("Missing required environment variables: %s", unsetVars))
		}

		metricFrequency = 1
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		configFilepath = filepath.Join(tempDir, "metric-config.yml")

		config = &MetricsConfig{
			Host:             "localhost",
			Username:         username,
			Password:         password,
			MetricsFrequency: metricFrequency,
			Origin:           "my_custom_origin",
			EmitLeaderFollowerMetrics: true,
			EmitMySQLMetrics:          true,
			EmitGaleraMetrics:         true,
			EmitDiskMetrics:           true,
			EmitCPUMetrics:            true,
		}

	})

	JustBeforeEach(func() {
		configBytes, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(configFilepath, configBytes, os.ModePerm)
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
			logFile, err := ioutil.TempFile(tempDir, "metrics_via_tcp_loopback_")
			Expect(err).NotTo(HaveOccurred())
			logFile.Close()
			logFilePath := logFile.Name()

			runMainWithArgs("-l", logFilePath)
			Consistently(session.ExitCode, time.Duration(metricFrequency)*time.Second).Should(Equal(-1))

			contents, err := ioutil.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())

			contentsAsString := fmt.Sprintf("%s", contents)
			Expect(contentsAsString).To(ContainSubstring("innodb/buffer_pool_pages_free"))

			firstLineCount := bytes.Count(contents, []byte("\n"))
			Expect(firstLineCount).Should(BeNumerically(">", 0))

			Consistently(session.ExitCode, time.Duration(metricFrequency)*time.Second).Should(Equal(-1))
			contents, err = ioutil.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())
			newLineCount := bytes.Count(contents, []byte("\n"))
			Expect(newLineCount).Should(BeNumerically(">", firstLineCount))
		})

		Context("when configured with a unix socket", func() {
			BeforeEach(func() {
				config.Host = socket
			})

			It("successfully emits log entries", func() {
				logFile, err := ioutil.TempFile(tempDir, "metrics_via_unix_socket_")
				Expect(err).NotTo(HaveOccurred())
				logFile.Close()

				logFilePath := logFile.Name()
				runMainWithArgs("-l", logFilePath)
				Consistently(session.ExitCode, time.Duration(metricFrequency)*time.Second).Should(Equal(-1))

				contents, err := ioutil.ReadFile(logFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).NotTo(BeEmpty())

				Expect(string(contents)).To(ContainSubstring("innodb/buffer_pool_pages_free"))
			})
		})

		It("logs that it has metrics", func() {
			logFilePath := filepath.Join(tempDir, "metrics.log")
			runMainWithArgs("-l", logFilePath)
			Consistently(session.ExitCode, time.Duration(metricFrequency*5)*time.Second).Should(Equal(-1))

			contents, err := ioutil.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())

			contentsAsString := fmt.Sprintf("%s", contents)
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_used\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_free\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_inodes_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_inodes_used\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/persistent_disk_inodes_free\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_used\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_free\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_inodes_used_percent\",\"value\":\\d+"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_inodes_used\",\"value\":\\d{2,}"))
			Expect(contentsAsString).To(MatchRegexp("system/ephemeral_disk_inodes_free\",\"value\":\\d{2,}"))
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
			contents, err := ioutil.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).NotTo(BeEmpty())
			contentsAsString := fmt.Sprintf("%s", contents)
			Expect(contentsAsString).To(ContainSubstring(`"key":"available","value":0`))
			Expect(contentsAsString).NotTo(ContainSubstring("innodb/buffer_pool_pages_free"))
		})
	})
})
