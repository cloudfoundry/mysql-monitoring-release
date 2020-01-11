package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mysql-diag-agent/config"
)

var _ = Describe("config", func() {
	var (
		tempDir string

		configFilepath string
		configContents []byte

		port               uint
		username           string
		password           string
		persistentDiskPath string
		ephemeralDiskPath  string
	)

	BeforeEach(func() {
		port = 1234
		username = "foo"
		password = "bar"
		persistentDiskPath = "/persistent"
		ephemeralDiskPath = "/ephemeral"

		configContents = []byte(fmt.Sprintf(
			`{Port: %d, Username: %s, Password: %s, PersistentDiskPath: %s, EphemeralDiskPath: %s}`,
			port,
			username,
			password,
			persistentDiskPath,
			ephemeralDiskPath,
		))

		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configFilepath = filepath.Join(tempDir, "mysql-diag-agent-config.yml")

		err = ioutil.WriteFile(configFilepath, configContents, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("reads the config file correctly", func() {
		c, err := config.LoadFromFile(configFilepath)
		Expect(err).NotTo(HaveOccurred())

		Expect(c.Port).To(Equal(port))
		Expect(c.Username).To(Equal(username))
		Expect(c.Password).To(Equal(password))
		Expect(c.PersistentDiskPath).To(Equal(persistentDiskPath))
		Expect(c.EphemeralDiskPath).To(Equal(ephemeralDiskPath))
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
})
