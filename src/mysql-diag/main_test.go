package main_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/mysql-diag/config"

	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/testutil"
)

var _ = Describe("mysql-diag cli", func() {
	var (
		tempDir        string
		configFilepath string

		cfg config.Config

		agentServer   *ghttp.Server
		agentHost     string
		agentPort     uint
		agentUsername string
		agentPassword string

		galeraAgentUsername       string
		galeraAgentPassword       string
		galeraAgentResponseStatus int
		galeraAgentResponse       interface{}

		diskUsedWarningPercent       uint
		diskInodesUsedWarningPercent uint

		persistentInodesFree uint64
		persistentBytesFree  uint64
		ephemeralInodesFree  uint64
		ephemeralBytesFree   uint64
	)

	BeforeEach(func() {

		agentServer = ghttp.NewServer()
		agentHost, agentPort = testutil.ParseURL(agentServer.URL())
		agentUsername = "agentfoo"
		agentPassword = "agentbar"

		galeraAgentUsername = "agentfoo"
		galeraAgentPassword = "agentbar"

		tempDir, err := os.MkdirTemp("", "")
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
			GaleraAgent: &config.GaleraAgentConfig{
				Username: galeraAgentUsername,
				Password: galeraAgentPassword,
				Host:     agentHost,
				ApiPort:  agentPort,
			},
		}
		writeAsYamlToFile(cfg, configFilepath)

		galeraAgentResponseStatus = http.StatusOK
		galeraAgentResponse = 123
	})

	JustBeforeEach(func() {

		agentServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/sequence_number"),
			ghttp.VerifyBasicAuth(galeraAgentUsername, galeraAgentPassword),
			ghttp.RespondWithJSONEncoded(galeraAgentResponseStatus, galeraAgentResponse),
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

		Eventually(session.Out).Should(gbytes.Say("INSTANCE"))
		Eventually(session.Out).Should(gbytes.Say("STATE"))
		Eventually(session.Out).Should(gbytes.Say("CLUSTER STATUS"))
		Eventually(session.Out).Should(gbytes.Say("N/A - ERROR"))

		Eventually(session, executableTimeout).Should(gexec.Exit(0))
	})

	It("reports the current active writer", func() {
		session := runMainWithArgs()

		Eventually(session.Out).Should(gbytes.Say("INSTANCE"))
		Eventually(session.Out).Should(gbytes.Say(`\[0\] mysql/uuid`))
		//FOR PAIR REVIEW - this doesn't feel great
		Consistently(session.Out).Should(Not(gbytes.Say("NOTE: Proxies will currently attempt to direct traffic to \"mysql/.*\"")))

		Eventually(session, executableTimeout).Should(gexec.Exit(0))
	})

	It("tells us that we need bootstrap", func() {
		session := runMainWithArgs()
		Eventually(session.Out).Should(gbytes.Say(`\[CRITICAL\] You must bootstrap the cluster. Follow these instructions: https://docs\.vmware\.com/en/VMware-SQL-with-MySQL-for-Tanzu-Application-Service/3\.2/mysql-for-tas/bootstrapping\.html`))
	})

	It("renders a disk information table", func() {
		session := runMainWithArgs()

		Eventually(session.Out).Should(gbytes.Say("INSTANCE"))
		Eventually(session.Out).Should(gbytes.Say("PERSISTENT DISK USED"))
		Eventually(session.Out).Should(gbytes.Say("EPHEMERAL DISK USED"))
		Eventually(session.Out).Should(gbytes.Say(`333B / 456B \(73\.0%\)`))

		Expect(agentServer.ReceivedRequests()).Should(HaveLen(2))

		Eventually(session, executableTimeout).Should(gexec.Exit(0))
	})

	It("tells us the download-logs command", func() {
		session := runMainWithArgs()
		Eventually(session.Out).Should(gbytes.Say(`\[CRITICAL\] Run the bosh logs command: targeting each of the VMs in your VMware SQL with MySQL for TAS cluster, proxies, and jumpbox to retrieve the VM logs.`))
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

	Context("when agent, replication canary and galera-agent is not present", func() {
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
		It("skips rep-canary check, disk check and sequence number", func() {
			session := runMainWithArgs()
			Eventually(session).Should(gexec.Exit())
			Eventually(session.Out).Should(gbytes.Say("Galera Agent not configured, skipping sequence number check"))
			Eventually(session.Out).Should(gbytes.Say("Agent not configured, skipping disk check"))
			Eventually(session.Out).Should(gbytes.Say("CLUSTER STATUS"))
			Eventually(session.Out).Should(gbytes.Say("N/A - ERROR"))
		})
	})
})

func writeAsYamlToFile(object interface{}, filepath string) {
	b, err := yaml.Marshal(object)
	Expect(err).NotTo(HaveOccurred())

	err = os.WriteFile(filepath, b, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
}
