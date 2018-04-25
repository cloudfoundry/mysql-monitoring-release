package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mysql-diag/config"
)

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

		diskUsedWarningPercent       uint
		diskInodesUsedWarningPercent uint
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

		formatString := `{"canary":{"username":"%s","password":"%s","api_port":%d},
				"mysql": {
				"username": "%s",
				"password": "%s",
				"port": %d,
				"agent": {
					"port": %d,
				 	"username": "%s",
					"password": "%s"
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

		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configFilepath = filepath.Join(tempDir, "mysql-diag-config.yml")

		err = ioutil.WriteFile(configFilepath, []byte(formatted), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
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

		Expect(c.Mysql.Username).To(Equal(username))
		Expect(c.Mysql.Password).To(Equal(password))
		Expect(c.Mysql.Port).To(Equal(port))
		Expect(c.Mysql.Agent.Port).To(Equal(agentPort))
		Expect(c.Mysql.Agent.Username).To(Equal(agentUsername))
		Expect(c.Mysql.Agent.Password).To(Equal(agentPassword))
		Expect(c.Mysql.Threshold.DiskUsedWarningPercent).To(Equal(diskUsedWarningPercent))
		Expect(c.Mysql.Threshold.DiskInodesUsedWarningPercent).To(Equal(diskInodesUsedWarningPercent))
		Expect(len(c.Mysql.Nodes)).To(Equal(1))
		Expect(c.Mysql.Nodes[0].Host).To(Equal(host))
		Expect(c.Mysql.Nodes[0].Name).To(Equal(nodeName))
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
			err := ioutil.WriteFile(configFilepath, []byte("invalid contents"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error", func() {
			_, err := config.LoadFromFile(configFilepath)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ConnectionString", func() {
		It("builds a mysql connection string", func() {
			node := config.MysqlNode{
				"nowhere.example.com",
				"somename",
				"uuid",
			}

			mysqlConfig := config.MysqlConfig{
				"someuser",
				"somepassword",
				3306,
				nil,
				nil,
				[]config.MysqlNode{node},
			}

			Expect(mysqlConfig.ConnectionString(node)).To(Equal("someuser:somepassword@tcp(nowhere.example.com:3306)/?timeout=10s"))
		})
	})

	It("provides a database connection object", func() {
		node := config.MysqlNode{
			"nowhere.example.com",
			"somename",
			"uuid",
		}

		mysqlConfig := config.MysqlConfig{
			"someuser",
			"somepassword",
			3306,
			nil,
			nil,
			[]config.MysqlNode{node},
		}

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
