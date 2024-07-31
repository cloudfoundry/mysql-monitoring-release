package galera_agent_client

import (
	"encoding/pem"
	"net/http"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/testutil"
)

var _ = Describe("GaleraAgentClient", func() {
	var (
		server            *ghttp.Server
		host              string
		port              uint
		galeraAgentClient *GaleraAgentClient

		responseStatus int
		response       interface{}
	)
	var galeraAgentCfg config.GaleraAgentConfig

	BeforeEach(func() {
		galeraAgentCfg = config.GaleraAgentConfig{Username: "username", Password: "fake-password", ApiPort: 1234}
	})

	Context("HTTP (plaintext)", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			host, port = testutil.ParseURL(server.URL())
			galeraAgentCfg.ApiPort = port
			galeraAgentClient = NewGaleraAgentClient(host, galeraAgentCfg)

			responseStatus = http.StatusOK
			response = 123
		})

		JustBeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/sequence_number"),
				ghttp.VerifyBasicAuth(galeraAgentCfg.Username, galeraAgentCfg.Password),
				ghttp.RespondWithJSONEncoded(responseStatus, response),
			))
		})

		It("returns sequence numbers without error", func() {
			status, err := galeraAgentClient.SequenceNumber()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(123))
		})

		Context("when making the request returns an error", func() {
			client := NewGaleraAgentClient("notvalid", galeraAgentCfg)

			It("returns an error", func() {
				_, err := client.SequenceNumber()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response status is not 200", func() {
			BeforeEach(func() {
				responseStatus = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := galeraAgentClient.SequenceNumber()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response fails to deserialize", func() {
			BeforeEach(func() {
				response = "not what we expected"
			})

			It("returns an error", func() {
				_, err := galeraAgentClient.SequenceNumber()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("HTTPS (TLS)", func() {
		pemEncode := func(raw []byte) string {
			return string(pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: raw,
			}))
		}

		BeforeEach(func() {
			server = ghttp.NewTLSServer()

			galeraAgentCfg.TLS.Enabled = true
			galeraAgentCfg.TLS.CA = pemEncode(server.HTTPTestServer.Certificate().Raw)
			galeraAgentCfg.TLS.ServerName = "example.com"

			host, port = testutil.ParseURL(server.URL())
			galeraAgentCfg.ApiPort = port
			galeraAgentClient = NewGaleraAgentClient(host, galeraAgentCfg)

			responseStatus = http.StatusOK
			response = 123
		})

		JustBeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/sequence_number"),
				ghttp.VerifyBasicAuth(galeraAgentCfg.Username, galeraAgentCfg.Password),
				ghttp.RespondWithJSONEncoded(responseStatus, response),
			))
		})

		It("returns status without error", func() {
			status, err := galeraAgentClient.SequenceNumber()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(123))
		})

		Context("when making the request returns an error", func() {
			client := NewGaleraAgentClient("notvalid", galeraAgentCfg)

			It("returns an error", func() {
				_, err := client.SequenceNumber()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response status is not 200", func() {
			BeforeEach(func() {
				responseStatus = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := galeraAgentClient.SequenceNumber()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response fails to deserialize", func() {
			BeforeEach(func() {
				response = "not what we expected"
			})

			It("returns an error", func() {
				_, err := galeraAgentClient.SequenceNumber()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
