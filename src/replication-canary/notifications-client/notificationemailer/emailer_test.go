package notificationemailer_test

import (
	. "github.com/cloudfoundry/replication-canary/notifications-client/notificationemailer"

	"net/http"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type fakeLogger struct{}

func (fakeLogger) Debug(string, map[string]interface{}) {}

var _ = Describe("Emailer", func() {
	var (
		notificationsDomain string

		server *ghttp.Server
		client *Client

		testLogger fakeLogger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		notificationsDomain = server.URL()
		testLogger = fakeLogger{}
	})

	JustBeforeEach(func() {
		skipSSLCertVerify := true
		client = NewClient(
			notificationsDomain,
			skipSSLCertVerify,
			testLogger,
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Email", func() {
		var (
			handler     http.HandlerFunc
			clientToken string
			to          string
			subject     string
			body        string
			kindID      string
		)

		BeforeEach(func() {
			clientToken = "foobar"
			to = "to@example.com"
			subject = "some email"
			body = "I am HTML"
			kindID = "notifications"

			handler = ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/emails"),
				ghttp.VerifyHeader(http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
				}),
				ghttp.VerifyHeader(http.Header{
					"Authorization": []string{"Bearer " + clientToken},
				}),
				ghttp.VerifyHeader(http.Header{
					"X-NOTIFICATIONS-VERSION": []string{"1"},
				}),
			)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(handler)
		})

		It("hits the notification domain's /email endpoint", func() {
			err := client.Email(
				clientToken,
				to,
				subject,
				body,
				kindID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when you misconfigure the request", func() {
			BeforeEach(func() {
				notificationsDomain = "not-a-real-domain"
			})

			It("returns an error", func() {
				err := client.Email(
					clientToken,
					to,
					subject,
					body,
					kindID,
				)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the server returns a 400 or above", func() {
			var (
				responseBody string
			)
			BeforeEach(func() {
				responseBody = "some-error"
				handler = ghttp.CombineHandlers(
					ghttp.RespondWith(400, responseBody),
				)
			})

			It("returns an error", func() {
				err := client.Email(
					clientToken,
					to,
					subject,
					body,
					kindID,
				)
				Expect(err).To(MatchError(errors.New("bad response sending email (400) - some-error")))
			})
		})
	})
})
