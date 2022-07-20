package hattery_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/mysql-diag/hattery"
)

type payload struct {
	Fookey string `json:"foo"`
	Barkey string `json:"bar"`
}

var _ = Describe("Hattery", func() {
	var (
		server   *ghttp.Server
		url      string
		response payload
		username string
		password string
	)

	Context("HTTP (plaintext)", func() {

		BeforeEach(func() {
			server = ghttp.NewServer()
			url = server.URL()
			response = payload{Fookey: "foovalue", Barkey: "barvalue"}
			username = "username"
			password = "password"
		})

		It("fetches stuff", func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/foo/bar"),
				ghttp.VerifyBasicAuth(username, password),
				ghttp.RespondWithJSONEncoded(http.StatusOK, response),
			))

			var fetched payload

			err := hattery.Url(url).Path("/foo").Path("/bar").BasicAuth(username, password).Timeout(5 * time.Second).Fetch(&fetched)

			Expect(err).ToNot(HaveOccurred())
			Expect(fetched).To(Equal(response))

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Context("TLS", func() {
		var (
			client *http.Client
		)
		BeforeEach(func() {
			server = ghttp.NewTLSServer()
			client = server.HTTPTestServer.Client()
			url = server.URL()
		})

		It("fetches stuff ovr TLS", func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/foo/bar"),
				ghttp.VerifyBasicAuth(username, password),
				ghttp.RespondWithJSONEncoded(http.StatusOK, response),
			))

			var fetched payload

			err := hattery.
				Url(url).
				Path("/foo").
				Path("/bar").
				BasicAuth(username, password).
				Timeout(5 * time.Second).
				Client(client).
				Fetch(&fetched)

			Expect(err).ToNot(HaveOccurred())
			Expect(fetched).To(Equal(response))

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
