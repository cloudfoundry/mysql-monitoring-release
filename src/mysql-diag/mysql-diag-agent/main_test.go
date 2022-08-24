package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/mysql-diag-agent/config"
	"github.com/cloudfoundry/mysql-diag/mysql-diag-agent/hattery"
)

var _ = Describe("mysql diag agent", func() {
	var (
		port     uint
		username string
		password string

		tempDir string

		runMainWithArgs func(args ...string) *gexec.Session
	)

	BeforeEach(func() {
		var err error

		tempDir, err = ioutil.TempDir("", "mysql-diag-integration-tests")
		Expect(err).NotTo(HaveOccurred())

		port = 59991
		username = "foo"
		password = "bar"

		configFilepath := filepath.Join(tempDir, "config.yml")

		c := config.Config{
			Port:               port,
			Username:           username,
			Password:           password,
			PersistentDiskPath: "/",
			EphemeralDiskPath:  "/",
		}

		writeAsYamlToFile(c, configFilepath)

		runMainWithArgs = func(args ...string) *gexec.Session {
			allArgs := []string{
				fmt.Sprintf("-c=%s", configFilepath),
			}

			allArgs = append(
				allArgs,
				args...,
			)

			_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", allArgs)
			Expect(err).NotTo(HaveOccurred())

			fmt.Println("binPath is " + binPath)
			command := exec.Command(binPath, allArgs...)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			return session
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Info endpoint", func() {
		var (
			session *gexec.Session
		)

		BeforeEach(func() {
			session = runMainWithArgs()
		})

		AfterEach(func() {
			session = session.Kill().Wait()

			Eventually(session).Should(gexec.Exit())
		})

		It("contains disk information", func() {
			url := fmt.Sprintf(
				"http://localhost:%d/api/v1/info",
				port,
			)

			var info diagagentclient.InfoResponse
			Eventually(func() error {
				return hattery.Url(url).BasicAuth(username, password).Fetch(&info)
			}).ShouldNot(HaveOccurred())

			Expect(info.Persistent.BytesTotal).ToNot(BeZero())
			Expect(info.Persistent.BytesFree).ToNot(BeZero())

			// Concourse containers always report 0 for free/used inodes (!!)
			//Expect(info.Persistent.InodesTotal).ToNot(BeZero())
			//Expect(info.Persistent.InodesFree).ToNot(BeZero())

			Expect(info.Ephemeral.BytesTotal).ToNot(BeZero())
			Expect(info.Ephemeral.BytesFree).ToNot(BeZero())

			// Concourse containers always report 0 for free/used inodes (!!)
			//Expect(info.Ephemeral.InodesTotal).ToNot(BeZero())
			//Expect(info.Ephemeral.InodesFree).ToNot(BeZero())
		})
	})
})

func writeAsYamlToFile(object interface{}, filepath string) {
	b, err := yaml.Marshal(object)
	Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile(filepath, b, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
}
