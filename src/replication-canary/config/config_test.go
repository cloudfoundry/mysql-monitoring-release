package config_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"code.cloudfoundry.org/tlsconfig/certtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/replication-canary/config"
)

var _ = Describe("Config", func() {
	var (
		rootConfig    *config.Config
		configuration string
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			configuration = `{
				"Canary":{
					"Database": "fake_database",
					"Username": "fake_username",
					"Password": "fake_password",
				},
				"MySQL":{
					"ClusterIPs": ["ip1","ip2","ip3"],
					"Port": 1337,
					"GaleraHealthcheckPort": 1424,
				},
				"Notifications":{
					"AdminClientUsername": "adminsomething",
					"AdminClientSecret": "adminsome-secret",
					"ClientUsername": "something",
					"ClientSecret": "some-secret",
					"NotificationsDomain": "some-notifications-domain.bosh-lite.com",
					"UAADomain": "some-uaa-domain.bosh-lite.com",
					"ToAddress": "to-address@example.com",
					"SystemDomain": "systemdomain",
					"ClusterIdentifier": "test-cluster-identifier",
				},
				"Switchboard":{
					"URLs": ["10.244.7.3","10.244.8.3"],
					"Username": "username",
					"Password": "password",
				},
				"WriteReadDelay": 5,
				"PollFrequency": 525600,
				"SkipSSLValidation": true,
				"APIPort": 8123,
				"TLS": {
					"Enabled": true,
					"Certificate": "pem-encoded-certificate",
					"PrivateKey": "pem-encoded-key",
				},
			}`
		})

		JustBeforeEach(func() {
			osArgs := []string{
				"replication-canary",
				fmt.Sprintf("-config=%s", configuration),
			}

			var err error
			rootConfig, err = config.NewConfig(osArgs)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return error on valid config", func() {
			err := rootConfig.Validate()
			Expect(err).NotTo(HaveOccurred())

			Expect(rootConfig.SkipSSLValidation).To(BeTrue())
		})

		It("contains Canary information", func() {
			Expect(rootConfig.Canary.Database).To(Equal("fake_database"))
			Expect(rootConfig.Canary.Username).To(Equal("fake_username"))
			Expect(rootConfig.Canary.Password).To(Equal("fake_password"))
		})

		It("contains APIPort information", func() {
			Expect(rootConfig.APIPort).To(Equal(uint(8123)))
		})

		It("contains TLS configuration", func() {
			Expect(rootConfig.TLS.Enabled).To(BeTrue())
			Expect(rootConfig.TLS.Certificate).To(Equal(`pem-encoded-certificate`))
			Expect(rootConfig.TLS.PrivateKey).To(Equal(`pem-encoded-key`))
		})

		It("contains Notifications information", func() {
			Expect(rootConfig.Notifications.ClientUsername).To(Equal("something"))
			Expect(rootConfig.Notifications.ClientSecret).To(Equal("some-secret"))
			Expect(rootConfig.Notifications.NotificationsDomain).To(Equal("some-notifications-domain.bosh-lite.com"))
			Expect(rootConfig.Notifications.UAADomain).To(Equal("some-uaa-domain.bosh-lite.com"))
			Expect(rootConfig.Notifications.ToAddress).To(Equal("to-address@example.com"))
			Expect(rootConfig.Notifications.SystemDomain).To(Equal("systemdomain"))
			Expect(rootConfig.Notifications.ClusterIdentifier).To(Equal("test-cluster-identifier"))
		})

		It("contains Switchboard information", func() {
			Expect(rootConfig.Switchboard.URLs).To(ConsistOf(
				"10.244.7.3",
				"10.244.8.3",
			))
			Expect(rootConfig.Switchboard.Username).To(Equal("username"))
			Expect(rootConfig.Switchboard.Password).To(Equal("password"))
		})

		It("does not let you have a WriteReadDelay larger than your PollFrequency", func() {
			err := rootConfig.Validate()
			Expect(err).NotTo(HaveOccurred())

			rootConfig.WriteReadDelay = 10
			rootConfig.PollFrequency = 1

			err = rootConfig.Validate()
			Expect(err).To(MatchError(config.InvalidDelay))

			rootConfig.WriteReadDelay = 10
			rootConfig.PollFrequency = 10

			err = rootConfig.Validate()
			Expect(err).To(MatchError(config.InvalidDelay))

			rootConfig.WriteReadDelay = 1
			rootConfig.PollFrequency = 10

			err = rootConfig.Validate()
			Expect(err).NotTo(HaveOccurred())
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
			rootConfig.APIPort = uint(10000 + GinkgoParallelProcess())
			rootConfig.TLS.Enabled = true
			rootConfig.TLS.Certificate = string(pemCert)
			rootConfig.TLS.PrivateKey = string(pemKey)
		})

		It("can provide a TLS listener", func() {
			l, err := rootConfig.NetworkListener()
			Expect(err).NotTo(HaveOccurred())
			defer l.Close()

			errCh := make(chan error, 1)
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

			address := fmt.Sprintf("%s:%d", rootConfig.BindAddress, rootConfig.APIPort)
			conn, err := tls.Dial("tcp", address, &tls.Config{
				RootCAs:    certPool,
				ServerName: "localhost",
			})
			Expect(err).NotTo(HaveOccurred())

			msg, err := ioutil.ReadAll(conn)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(msg)).To(Equal("foo"))
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
					if err == nil {
						defer conn.Close()
						_, _ = conn.Write([]byte("foo"))
					}
					errCh <- err
				}()

				address := fmt.Sprintf("%s:%d", rootConfig.BindAddress, rootConfig.APIPort)
				conn, err := net.Dial("tcp", address)
				Expect(err).NotTo(HaveOccurred())

				Expect(conn.SetReadDeadline(time.Now().Add(5 * time.Second))).To(Succeed())
				msg, err := ioutil.ReadAll(conn)
				Expect(err).NotTo(HaveOccurred(), `Expected to successfully read data from a plaintext connection, but it failed`)
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
