package diagagentclient_test

import (
	"encoding/pem"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/mysql-diag/config"
	. "github.com/cloudfoundry/mysql-diag/diagagentclient"
)

var _ = Describe("DiagAgentClient", func() {
	var (
		server          *ghttp.Server
		address         string
		diagAgentClient *DiagAgentClient

		responseStatus int
		response       interface{}
		cfg            config.AgentConfig
	)

	BeforeEach(func() {
		cfg = config.AgentConfig{
			Username: "foo",
			Password: "bar",
		}
	})

	Context("HTTP (plaintext", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			serverURL, err := url.Parse(server.URL())
			Expect(err).NotTo(HaveOccurred())
			address = serverURL.Host
			diagAgentClient = NewDiagAgentClient(cfg)

			responseStatus = http.StatusOK
			response = InfoResponse{
				Persistent: DiskInfo{
					BytesTotal:  456,
					BytesFree:   123,
					InodesTotal: 789,
					InodesFree:  567,
				},
				Ephemeral: DiskInfo{
					BytesTotal:  1456,
					BytesFree:   1123,
					InodesTotal: 1789,
					InodesFree:  1567,
				},
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/info"),
				ghttp.VerifyBasicAuth(cfg.Username, cfg.Password),
				ghttp.RespondWithJSONEncoded(responseStatus, response),
			))
		})

		It("returns info without error", func() {
			info, err := diagAgentClient.Info(address, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(*info).To(Equal(response))
		})

		Context("when making the request returns an error", func() {
			client := NewDiagAgentClient(cfg)

			It("returns an error", func() {
				info, err := client.Info("notvalid:12345", false)
				Expect(err).To(HaveOccurred())
				Expect(info).To(BeNil())
			})
		})

		Context("when the response status is not 200", func() {
			BeforeEach(func() {
				responseStatus = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := diagAgentClient.Info(address, false)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response fails to deserialize", func() {
			BeforeEach(func() {
				response = "not what we expected"
			})

			It("returns an error", func() {
				_, err := diagAgentClient.Info(address, false)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("TLS", func() {
		BeforeEach(func() {
			server = ghttp.NewTLSServer()
			serverURL, err := url.Parse(server.URL())
			Expect(err).NotTo(HaveOccurred())
			address = serverURL.Host

			cfg.TLS.Enabled = true
			cfg.TLS.CA = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.HTTPTestServer.Certificate().Raw}))
			cfg.TLS.ServerName = "example.com"

			diagAgentClient = NewDiagAgentClient(cfg)

			responseStatus = http.StatusOK
			response = InfoResponse{
				Persistent: DiskInfo{
					BytesTotal:  456,
					BytesFree:   123,
					InodesTotal: 789,
					InodesFree:  567,
				},
				Ephemeral: DiskInfo{
					BytesTotal:  1456,
					BytesFree:   1123,
					InodesTotal: 1789,
					InodesFree:  1567,
				},
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/info"),
				ghttp.VerifyBasicAuth(cfg.Username, cfg.Password),
				ghttp.RespondWithJSONEncoded(responseStatus, response),
			))
		})

		It("returns info without error", func() {
			info, err := diagAgentClient.Info(address, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(*info).To(Equal(response))
		})

		Context("when making the request returns an error", func() {
			It("returns an error", func() {
				client := NewDiagAgentClient(cfg)

				info, err := client.Info("notvalid:12345", true)
				Expect(err).To(HaveOccurred())
				Expect(info).To(BeNil())
			})
		})

		Context("when the response status is not 200", func() {
			BeforeEach(func() {
				responseStatus = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := diagAgentClient.Info(address, true)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response fails to deserialize", func() {
			BeforeEach(func() {
				response = "not what we expected"
			})

			It("returns an error", func() {
				_, err := diagAgentClient.Info(address, true)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
