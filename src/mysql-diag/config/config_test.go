package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/mysql-diag/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatabaseConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("config", func() {
	var (
		tempDir string

		configFilepath string

		canaryUsername string
		canaryPassword string
		canaryPort     uint

		host          string
		port          uint
		username      string
		password      string
		nodeName      string
		uuid          string
		agentPort     uint
		agentUsername string
		agentPassword string

		galeraAgentUsername string
		galeraAgentPassword string
		galeraAgentPort     uint

		diskUsedWarningPercent       uint
		diskInodesUsedWarningPercent uint

		node        config.MysqlNode
		mysqlConfig config.MysqlConfig
	)

	BeforeEach(func() {
		host = "hostname"
		port = 1234
		nodeName = "nodeName"
		username = "foo"
		password = "bar"
		canaryUsername = "canary"
		canaryPassword = "canaryPassword"
		canaryPort = 8123
		uuid = "abcd-efgh"
		agentPort = 8124
		agentUsername = "agentfoo"
		agentPassword = "agentPass"
		diskUsedWarningPercent = 88
		diskInodesUsedWarningPercent = 77
		galeraAgentUsername = "galera-agent"
		galeraAgentPassword = "galera-agent-password"
		galeraAgentPort = 9201

		formatString := `{
				"canary": {
					"username": "%s",
					"password": "%s",
					"api_port": %d,
					"tls": {
						"enabled": true,
						"ca": "pem-encoded-authority-for-replication",
						"server_name": "expected-replication-canary-identity"
					}
				},
				"galera_agent": {
					"username": "%s",
					"password": "%s",
					"api_port": %d,
					"tls": {
						"enabled": true,
						"ca": "pem-encoded-authority-for-galera-agent",
						"server_name": "expected-galera-agent-identity"
					}
				},
				"mysql": {
				"username": "%s",
				"password": "%s",
				"port": %d,
				"agent": {
					"port": %d,
				 	"username": "%s",
					"password": "%s",
					"tls": {
						"enabled": true,
						"ca": "pem-encoded-authority",
						"server_name": "expected-mysql-diag-agent-identity"
					}
				},
				"threshold": {
					"disk_used_warning_percent": %d,
					"disk_inodes_used_warning_percent": %d
				},
				"nodes": [
					{
				 	"host": "%s",
				 	"name": "%s",
				 	"uuid": "%s"
				 	}
				]}}`
		formatted := fmt.Sprintf(
			formatString,
			canaryUsername,
			canaryPassword,
			canaryPort,
			galeraAgentUsername,
			galeraAgentPassword,
			galeraAgentPort,
			username,
			password,
			port,
			agentPort,
			agentUsername,
			agentPassword,
			diskUsedWarningPercent,
			diskInodesUsedWarningPercent,
			host,
			nodeName,
			uuid,
		)

		tempDir, err := os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		configFilepath = filepath.Join(tempDir, "mysql-diag-config.yml")

		err = os.WriteFile(configFilepath, []byte(formatted), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		node = config.MysqlNode{
			Host: "nowhere.example.com",
			Name: "somename",
			UUID: "uuid",
		}

		mysqlConfig = config.MysqlConfig{
			Username: "someuser",
			Password: "somepassword",
			Port:     3306,
			Nodes:    []config.MysqlNode{node},
		}

	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("reads the config file correctly", func() {
		c, err := config.LoadFromFile(configFilepath)
		Expect(err).NotTo(HaveOccurred())

		Expect(c.Canary.Username).To(Equal(canaryUsername))
		Expect(c.Canary.Password).To(Equal(canaryPassword))
		Expect(c.Canary.ApiPort).To(Equal(canaryPort))
		Expect(c.Canary.TLS.Enabled).To(BeTrue())
		Expect(c.Canary.TLS.CA).To(Equal(`pem-encoded-authority-for-replication`))
		Expect(c.Canary.TLS.ServerName).To(Equal(`expected-replication-canary-identity`))

		Expect(c.Mysql.Username).To(Equal(username))
		Expect(c.Mysql.Password).To(Equal(password))
		Expect(c.Mysql.Port).To(Equal(port))
		Expect(c.Mysql.Agent.Port).To(Equal(agentPort))
		Expect(c.Mysql.Agent.Username).To(Equal(agentUsername))
		Expect(c.Mysql.Agent.Password).To(Equal(agentPassword))
		Expect(c.Mysql.Agent.TLS.Enabled).To(BeTrue())
		Expect(c.Mysql.Agent.TLS.CA).To(Equal("pem-encoded-authority"))
		Expect(c.Mysql.Agent.TLS.ServerName).To(Equal("expected-mysql-diag-agent-identity"))
		Expect(c.Mysql.Threshold.DiskUsedWarningPercent).To(Equal(diskUsedWarningPercent))
		Expect(c.Mysql.Threshold.DiskInodesUsedWarningPercent).To(Equal(diskInodesUsedWarningPercent))
		Expect(len(c.Mysql.Nodes)).To(Equal(1))
		Expect(c.Mysql.Nodes[0].Host).To(Equal(host))
		Expect(c.Mysql.Nodes[0].Name).To(Equal(nodeName))

		Expect(c.GaleraAgent).To(Not(BeNil()))
		Expect(c.GaleraAgent.Username).To(Equal(galeraAgentUsername))
		Expect(c.GaleraAgent.Password).To(Equal(galeraAgentPassword))
		Expect(c.GaleraAgent.ApiPort).To(Equal(galeraAgentPort))
		Expect(c.GaleraAgent.TLS.Enabled).To(BeTrue())
		Expect(c.GaleraAgent.TLS.CA).To(Equal(`pem-encoded-authority-for-galera-agent`))
		Expect(c.GaleraAgent.TLS.ServerName).To(Equal(`expected-galera-agent-identity`))
	})

	Context("when the file does not exist", func() {
		BeforeEach(func() {
			configFilepath = "not a valid file"
		})

		It("returns an error", func() {
			_, err := config.LoadFromFile(configFilepath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the file contents cannot be unmarshalled", func() {
		BeforeEach(func() {
			err := os.WriteFile(configFilepath, []byte("invalid contents"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error", func() {
			_, err := config.LoadFromFile(configFilepath)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ConnectionString", func() {
		It("builds a mysql connection string", func() {
			Expect(mysqlConfig.ConnectionString(node)).To(Equal("someuser:somepassword@tcp(nowhere.example.com:3306)/?timeout=10s&tls=preferred"))
		})
	})

	It("provides a database connection object", func() {
		conn := mysqlConfig.Connection(node)

		// It's not really connected to the database, it's lazy, so there's not much to assert
		Expect(conn).ToNot(BeNil())
	})

	It("provides a list of hosts with logs", func() {
		c, err := config.LoadFromFile(configFilepath)
		Expect(err).NotTo(HaveOccurred())

		hosts := c.HostsWithLogs()
		Expect(hosts).To(HaveLen(1))
		Expect(hosts[0]).To(Equal(host))
	})
})
