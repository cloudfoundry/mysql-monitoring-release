package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/mysql-diag/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/cloudfoundry/mysql-diag/canaryclient"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/testutil"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("mysql-diag cli", func() {
	var (
		tempDir        string
		configFilepath string

		cfg config.Config

		canaryServer         *ghttp.Server
		canaryUsername       string
		canaryPassword       string
		canaryResponseStatus int
		canaryResponse       interface{}

		agentServer   *ghttp.Server
		agentHost     string
		agentPort     uint
		agentUsername string
		agentPassword string

		diskUsedWarningPercent       uint
		diskInodesUsedWarningPercent uint

		persistentInodesFree uint64
		persistentBytesFree  uint64
		ephemeralInodesFree  uint64
		ephemeralBytesFree   uint64
	)

	BeforeEach(func() {
		canaryServer = ghttp.NewServer()
		_, canaryPort := testutil.ParseURL(canaryServer.URL())
		canaryUsername = "foo"
		canaryPassword = "bar"

		agentServer = ghttp.NewServer()
		agentHost, agentPort = testutil.ParseURL(agentServer.URL())
		agentUsername = "agentfoo"
		agentPassword = "agentbar"

		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		nodeName := "mysql"

		diskUsedWarningPercent = 90
		diskInodesUsedWarningPercent = 90

		persistentInodesFree = 567
		persistentBytesFree = 123
		ephemeralInodesFree = 1567
		ephemeralBytesFree = 1123

		configFilepath = filepath.Join(tempDir, "mysql-diag-config.yml")
		cfg = config.Config{
			Canary: &config.CanaryConfig{
				Username: canaryUsername,
				Password: canaryPassword,
				ApiPort:  canaryPort,
			},
			Mysql: config.MysqlConfig{
				Agent: &config.AgentConfig{
					Port:     agentPort,
					Username: agentUsername,
					Password: agentPassword,
				},
				Threshold: &config.ThresholdConfig{
					DiskUsedWarningPercent:       diskUsedWarningPercent,
					DiskInodesUsedWarningPercent: diskInodesUsedWarningPercent,
				},
				Nodes: []config.MysqlNode{
					{
						Host: agentHost,
						Name: nodeName,
						UUID: "uuid",
					},
				},
			},
		}
		writeAsYamlToFile(cfg, configFilepath)

		canaryResponseStatus = http.StatusOK
		canaryResponse = canaryclient.CanaryStatus{Healthy: false}
	})

	JustBeforeEach(func() {
		canaryServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/v1/status"),
			ghttp.VerifyBasicAuth(canaryUsername, canaryPassword),
			ghttp.RespondWithJSONEncoded(canaryResponseStatus, canaryResponse),
		))

		response := diagagentclient.InfoResponse{
			Persistent: diagagentclient.DiskInfo{
				BytesTotal:  456,
				BytesFree:   persistentBytesFree,
				InodesTotal: 789,
				InodesFree:  persistentInodesFree,
			},
			Ephemeral: diagagentclient.DiskInfo{
				BytesTotal:  1456,
				BytesFree:   ephemeralBytesFree,
				InodesTotal: 1789,
				InodesFree:  ephemeralInodesFree,
			},
		}
		agentServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/v1/info"),
			ghttp.VerifyBasicAuth(agentUsername, agentPassword),
			ghttp.RespondWithJSONEncoded(http.StatusOK, response),
		))
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	runMainWithArgs := func(args ...string) *gexec.Session {
		args = append(
			args,
			fmt.Sprintf("-c=%s", configFilepath),
		)

		_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", args)
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(mysqlDiagBinPath, args...)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	It("prints the date", func() {
		session := runMainWithArgs()

		Eventually(session.Out).Should(gbytes.Say("\\w+ \\w+ +\\d+ \\d+:\\d+:\\d+ UTC \\d{4}"))
	})

	It("renders a wsrep status table without error", func() {
		session := runMainWithArgs()

		Eventually(session.Out).Should(gbytes.Say("HOST"))
		Eventually(session.Out).Should(gbytes.Say("NAME/UUID"))
		Eventually(session.Out).Should(gbytes.Say("WSREP LOCAL STATE"))
		Eventually(session.Out).Should(gbytes.Say("WSREP CLUSTER STATUS"))
		Eventually(session.Out).Should(gbytes.Say("WSREP CLUSTER SIZE"))

		Expect(canaryServer.ReceivedRequests()).Should(HaveLen(1))

		Eventually(session, executableTimeout).Should(gexec.Exit(0))
	})

	It("tells us that we need bootstrap", func() {
		session := runMainWithArgs()
		Eventually(session.Out).Should(gbytes.Say("\\[CRITICAL\\] You must bootstrap the cluster. Follow these instructions: https://docs.pivotal.io/p-mysql/bootstrapping.html"))
	})

	It("tells us that the canary is unhealthy", func() {
		session := runMainWithArgs()
		Eventually(session.Out).Should(gbytes.Say("\\[CRITICAL\\] The replication process is unhealthy. Writes are disabled."))
	})

	It("renders a disk information table", func() {
		session := runMainWithArgs()

		Eventually(session.Out).Should(gbytes.Say("HOST"))
		Eventually(session.Out).Should(gbytes.Say("NAME/UUID"))
		Eventually(session.Out).Should(gbytes.Say("PERSISTENT DISK USED"))
		Eventually(session.Out).Should(gbytes.Say("EPHEMERAL DISK USED"))
		Eventually(session.Out).Should(gbytes.Say("73.0"))

		Expect(agentServer.ReceivedRequests()).Should(HaveLen(1))

		Eventually(session, executableTimeout).Should(gexec.Exit(0))
	})

	It("tells us the download-logs command", func() {
		session := runMainWithArgs()
		Eventually(session.Out).Should(gbytes.Say("\\[CRITICAL\\] Run the download-logs command:"))
		Eventually(session.Out).Should(gbytes.Say("\\$ download-logs -o /tmp/output"))
		Eventually(session.Out).Should(gbytes.Say("For full information about how to download and use the download-logs command see https://discuss.pivotal.io/hc/en-us/articles/221504408"))
	})

	It("warns us to not do silly things", func() {
		session := runMainWithArgs()
		Eventually(session.Out).Should(gbytes.Say("\\[WARNING\\] NOT RECOMMENDED"))
	})

	It("does not render any warnings for disk space", func() {
		session := runMainWithArgs()
		Consistently(session.Out).ShouldNot(gbytes.Say("\\[WARNING\\] Persistent disk usage is very high on node mysql."))
		Consistently(session.Out).ShouldNot(gbytes.Say("\\[WARNING\\] Ephemeral disk usage is very high on node mysql."))
	})

	Context("when the disk usage is above the thresholds", func() {
		BeforeEach(func() {
			persistentInodesFree = 1
			persistentBytesFree = 1
			ephemeralInodesFree = 1
			ephemeralBytesFree = 1
		})

		It("renders a warning for the user", func() {
			session := runMainWithArgs()
			Eventually(session.Out).Should(gbytes.Say("\\[WARNING\\] Ephemeral disk usage is very high on node mysql/uuid"))
			Eventually(session.Out).Should(gbytes.Say("\\[WARNING\\] Persistent disk usage is very high on node mysql/uuid"))
		})
	})

	Context("when agent and replication canary is not present", func() {
		BeforeEach(func() {
			nodeName := "mysql"

			cfg = config.Config{
				Mysql: config.MysqlConfig{
					Nodes: []config.MysqlNode{
						{
							Host: agentHost,
							Name: nodeName,
						},
					},
				},
			}

			writeAsYamlToFile(cfg, configFilepath)
		})

		It("skips rep-canary check and disk check", func() {
			session := runMainWithArgs()
			Eventually(session.Out).Should(gbytes.Say("Canary not configured, skipping health check"))
			Eventually(session.Out).Should(gbytes.Say("WSREP CLUSTER STATUS"))
			Eventually(session.Out).Should(gbytes.Say("Agent not configured, skipping disk check"))
		})
	})
})

func writeAsYamlToFile(object interface{}, filepath string) {
	b, err := yaml.Marshal(object)
	Expect(err).NotTo(HaveOccurred())

	err = ioutil.WriteFile(filepath, b, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
}
