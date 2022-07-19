package config_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/tlsconfig/certtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/cloudfoundry/mysql-diag/mysql-diag-agent/config"
)

var _ = Describe("config", func() {
	var (
		tempDir string

		configFilepath string

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

		configContents := []byte(fmt.Sprintf(
			`{Port: %d, Username: %s, Password: %s, PersistentDiskPath: %s, EphemeralDiskPath: %s}`,
			port,
			username,
			password,
			persistentDiskPath,
			ephemeralDiskPath,
		))

		var err error
		tempDir, err = ioutil.TempDir("", "")
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
		Expect(c.TLS.Enabled).To(BeFalse())
		Expect(c.TLS.Certificate).To(BeEmpty())
		Expect(c.TLS.PrivateKey).To(BeEmpty())
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

	Context("when TLS is enabled", func() {
		BeforeEach(func() {
			m := map[string]any{
				"Port":               port,
				"Username":           username,
				"Password":           password,
				"PersistentDiskPath": persistentDiskPath,
				"EphemeralDiskPath":  ephemeralDiskPath,
				"TLS": map[string]any{
					"Enabled":     true,
					"Certificate": "some-pem-certificate",
					"PrivateKey":  "some-pem-private-key",
				},
			}
			configContents, err := yaml.Marshal(m)
			Expect(err).NotTo(HaveOccurred())

			configFilepath = filepath.Join(tempDir, "mysql-diag-agent-config.yml")

			err = ioutil.WriteFile(configFilepath, configContents, os.ModePerm)
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
			Expect(c.TLS.Enabled).To(BeTrue())
			Expect(c.TLS.Certificate).To(Equal(`some-pem-certificate`))
			Expect(c.TLS.PrivateKey).To(Equal(`some-pem-private-key`))
		})
	})

	Context("NetworkListener", func() {
		var rootConfig config.Config
		BeforeEach(func() {
			serverAuthority, err := certtest.BuildCA("test")
			Expect(err).ToNot(HaveOccurred())

			cert, err := serverAuthority.BuildSignedCertificate("localhost")
			Expect(err).NotTo(HaveOccurred())

			pemCert, pemKey, err := cert.CertificatePEMAndPrivateKey()
			Expect(err).NotTo(HaveOccurred())

			rootConfig.BindAddress = "127.0.0.1"
			rootConfig.Port = uint(10000 + GinkgoParallelProcess())
			rootConfig.TLS.Enabled = true
			rootConfig.TLS.Certificate = string(pemCert)
			rootConfig.TLS.PrivateKey = string(pemKey)
		})

		It("can provide a TLS listener", func() {
			l, err := rootConfig.NetworkListener()
			Expect(err).NotTo(HaveOccurred())
			defer l.Close()

			errCh := make(chan error)
			go func() {
				conn, err := l.Accept()
				if err != nil {
					errCh <- err
					return
				}
				defer conn.Close()
				conn.Write([]byte("foo"))
				errCh <- err
			}()

			block, _ := pem.Decode([]byte(rootConfig.TLS.Certificate))
			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).NotTo(HaveOccurred())
			certPool := x509.NewCertPool()
			certPool.AddCert(cert)

			address := fmt.Sprintf("%s:%d", rootConfig.BindAddress, rootConfig.Port)
			conn, err := tls.Dial("tcp", address, &tls.Config{
				RootCAs:    certPool,
				ServerName: "localhost",
			})
			Expect(err).NotTo(HaveOccurred())

			var buf [3]byte
			_, err = io.ReadFull(conn, buf[:])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buf[:])).To(Equal("foo"))

			Expect(conn.Close()).To(Succeed())

			Eventually(errCh).Should(Receive(nil))
		})

		When("tls is disabled", func() {
			It("provides a plaintext TCP listener", func() {
				rootConfig.TLS.Enabled = false

				l, err := rootConfig.NetworkListener()
				Expect(err).NotTo(HaveOccurred())
				defer l.Close()

				errCh := make(chan error, 1)
				go func() {
					conn, err := l.Accept()
					defer conn.Close()
					_, _ = conn.Write([]byte("foo"))
					errCh <- err
				}()

				address := fmt.Sprintf("%s:%d", rootConfig.BindAddress, rootConfig.Port)
				conn, err := net.Dial("tcp", address)
				Expect(err).NotTo(HaveOccurred())

				msg, err := ioutil.ReadAll(conn)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(msg)).To(Equal("foo"))
				Expect(conn.Close()).To(Succeed())
				Eventually(errCh).Should(Receive(nil))
			})
		})

		When("tls is misconfigured", func() {
			It("returns an error", func() {
				rootConfig.TLS.Enabled = true
				rootConfig.TLS.Certificate = "not proper PEM content"

				_, err := rootConfig.NetworkListener()
				Expect(err).To(MatchError(`tls: failed to find any PEM data in certificate input`))
			})
		})
	})
})
