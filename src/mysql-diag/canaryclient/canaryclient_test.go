package canaryclient_test

import (
	"net/http"

	. "github.com/cloudfoundry/mysql-diag/canaryclient"
	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var goodCanaryConfig = config.CanaryConfig{Username: "username", Password: "fake-password", ApiPort: 1234}

var _ = Describe("Canaryclient", func() {
	var (
		server       *ghttp.Server
		host         string
		port         uint
		canaryclient *CanaryClient

		responseStatus int
		response       interface{}
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		host, port = testutil.ParseURL(server.URL())
		canaryclient = NewCanaryClient(host, port, goodCanaryConfig)

		responseStatus = http.StatusOK
		response = CanaryStatus{Healthy: true}
	})

	JustBeforeEach(func() {
		server.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/v1/status"),
			ghttp.VerifyBasicAuth(goodCanaryConfig.Username, goodCanaryConfig.Password),
			ghttp.RespondWithJSONEncoded(responseStatus, response),
		))
	})

	It("returns status without error", func() {
		status, err := canaryclient.Status()
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(BeTrue())
	})

	Context("when making the request returns an error", func() {
		client := NewCanaryClient("notvalid", port, goodCanaryConfig)

		It("returns an error", func() {
			_, err := client.Status()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the response status is not 200", func() {
		BeforeEach(func() {
			responseStatus = http.StatusTeapot
		})

		It("returns an error", func() {
			_, err := canaryclient.Status()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the response fails to deserialize", func() {
		BeforeEach(func() {
			response = "not what we expected"
		})

		It("returns an error", func() {
			_, err := canaryclient.Status()
			Expect(err).To(HaveOccurred())
		})
	})
})
