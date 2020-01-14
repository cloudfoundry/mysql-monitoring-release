package switchboard_test

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/cloudfoundry/replication-canary/switchboard"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type backends []backend

type backend struct {
	Host   string `json:"host"`
	Active bool   `json:"active"`
}

var _ = Describe("Client", func() {
	var (
		domain string
		client *Client
		server *ghttp.Server

		testLogger *lagertest.TestLogger
	)

	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("switchboard client test")

		server = ghttp.NewServer()
		domain = server.URL()
	})

	JustBeforeEach(func() {
		skipSSLCertVerify := false

		client = NewClient(
			domain,
			"username",
			"password",
			skipSSLCertVerify,
			testLogger,
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("DisableClusterTraffic", func() {
		It("sends a patch to disable cluster traffic", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/v0/cluster", "trafficEnabled=false&message=Disabling%20cluster%20traffic"),
					ghttp.VerifyBasicAuth("username", "password"),
				),
			)
			err := client.DisableClusterTraffic()

			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		Context("when creating the request fails", func() {
			BeforeEach(func() {
				// induce a failure by providing an invalid URL that fails to parse
				domain = "%%"
			})

			It("returns an error", func() {
				err := client.DisableClusterTraffic()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when executing the request fails", func() {
			BeforeEach(func() {
				// induce a failure by providing a URL that cannot be routed
				domain = "invalid-domain$^&*"
			})

			It("returns an error", func() {
				err := client.DisableClusterTraffic()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response has a non-200 status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(400, nil, nil),
				)
			})

			It("returns error", func() {
				err := client.DisableClusterTraffic()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("EnableClusterTraffic", func() {
		It("sends a patch to enable cluster traffic", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/v0/cluster", "trafficEnabled=true&message=Enabling%20cluster%20traffic"),
					ghttp.VerifyBasicAuth("username", "password"),
				),
			)
			err := client.EnableClusterTraffic()

			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		Context("when creating the request fails", func() {
			BeforeEach(func() {
				// induce a failure by providing an invalid URL that fails to parse
				domain = "%%"
			})

			It("returns an error", func() {
				err := client.EnableClusterTraffic()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when executing the request fails", func() {
			BeforeEach(func() {
				// induce a failure by providing a URL that cannot be routed
				domain = "invalid-domain$^&*"
			})

			It("returns an error", func() {
				err := client.EnableClusterTraffic()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response has a non-200 status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(400, nil, nil),
				)
			})

			It("returns error", func() {
				err := client.EnableClusterTraffic()
				Expect(err).To(HaveOccurred())
			})

		})
	})

	Describe("ActiveBackendHost", func() {
		var (
			returnedBackends backends
			respStatus       int
			respBody         []byte
		)

		BeforeEach(func() {
			returnedBackends = []backend{
				{
					Host:   "host-0",
					Active: false,
				},
				{
					Host:   "host-1",
					Active: true,
				},
			}

			var err error
			respBody, err = json.Marshal(returnedBackends)
			Expect(err).NotTo(HaveOccurred())

			respStatus = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v0/backends"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeaderKV("X-Forwarded-Proto", "https"),
					ghttp.RespondWith(respStatus, respBody),
				),
			)
		})

		It("returns the IP address of the active backend", func() {
			ip, err := client.ActiveBackendHost()
			Expect(err).NotTo(HaveOccurred())

			Expect(ip).To(Equal(returnedBackends[1].Host))
		})

		Context("when creating the request fails", func() {
			BeforeEach(func() {
				// induce a failure by providing an invalid URL that fails to parse
				domain = "%%"
			})

			It("returns an error", func() {
				_, err := client.ActiveBackendHost()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when executing the request fails", func() {
			BeforeEach(func() {
				// induce a failure by providing a URL that cannot be routed
				domain = "invalid-domain$^&*"
			})

			It("returns an error", func() {
				_, err := client.ActiveBackendHost()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response has a non-200 status code", func() {
			BeforeEach(func() {
				respStatus = http.StatusBadRequest
			})

			It("returns error", func() {
				_, err := client.ActiveBackendHost()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response body cannot be unmarshalled from json", func() {
			BeforeEach(func() {
				respBody = []byte("&*(234")
			})

			It("returns error", func() {
				_, err := client.ActiveBackendHost()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when no active backend is found", func() {
			BeforeEach(func() {
				returnedBackends[0].Active = false
				returnedBackends[1].Active = false

				var err error
				respBody, err = json.Marshal(returnedBackends)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns error", func() {
				_, err := client.ActiveBackendHost()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
