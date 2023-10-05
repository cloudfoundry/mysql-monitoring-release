package alert_test

import (
	"bytes"
	"errors"
	"log/slog"

	"code.cloudfoundry.org/uaa-go-client/schema"
	. "github.com/cloudfoundry/replication-canary/alert"
	"github.com/cloudfoundry/replication-canary/alert/alertfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmailingAlerter", func() {
	var (
		testWriter              *bytes.Buffer
		alerter                 *EmailingAlerter
		fakeUAAClient           *alertfakes.FakeUAAClient
		fakeNotificationsClient *alertfakes.FakeNotificationsClient
		toAddress               string
		systemDomain            string
		clusterIdentifier       string
	)

	BeforeEach(func() {
		testWriter = new(bytes.Buffer)
		// need to set 'LevelDebug' because otherwise it uses the default of Info
		testHandler := slog.NewJSONHandler(testWriter, &slog.HandlerOptions{Level: slog.LevelDebug})
		testLogger := slog.New(testHandler)

		fakeUAAClient = new(alertfakes.FakeUAAClient)
		fakeNotificationsClient = new(alertfakes.FakeNotificationsClient)
		toAddress = "barbaz@example.com"
		systemDomain = "system-domain"
		clusterIdentifier = "test-cluster-identifier"

		token := &schema.Token{
			AccessToken: "foobar",
		}

		fakeUAAClient.FetchTokenReturns(token, nil)

		alerter = &EmailingAlerter{
			Logger:              testLogger,
			UAAClient:           fakeUAAClient,
			NotificationsClient: fakeNotificationsClient,
			ToAddress:           toAddress,
			SystemDomain:        systemDomain,
			ClusterIdentifier:   clusterIdentifier,
		}
	})

	Describe("NotUnhealthy", func() {
		It("no-ops", func() {
			err := alerter.NotUnhealthy()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeUAAClient.FetchTokenCallCount()).To(Equal(0))
			Expect(fakeUAAClient.FetchKeyCallCount()).To(Equal(0))
			Expect(fakeNotificationsClient.EmailCallCount()).To(Equal(0))
		})
	})

	Describe("Unhealthy", func() {
		It("grabs a token from the UAAClient and sends an email with it", func() {
			err := alerter.Unhealthy()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeUAAClient.FetchTokenCallCount()).To(Equal(1))

			Expect(fakeNotificationsClient.EmailCallCount()).To(Equal(1))

			clientToken, _, _, _, _ := fakeNotificationsClient.EmailArgsForCall(0)
			Expect(clientToken).To(Equal("foobar"))
		})

		It("invokes email with correct arguments", func() {
			err := alerter.Unhealthy()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeNotificationsClient.EmailCallCount()).To(Equal(1))

			_, to, subject, html, kindID := fakeNotificationsClient.EmailArgsForCall(0)

			Expect(to).To(Equal(toAddress))
			Expect(subject).To(Equal("[system-domain][test-cluster-identifier] p-mysql Replication Canary, alert 417"))
			Expect(html).To(Equal("{alert-code 417}<br/>This is an e-mail to notify you that the MySQL service's replication canary has detected an unsafe cluster condition in which replication is not performing as expected across all nodes."))
			Expect(kindID).To(Equal("p-mysql"))

		})

		Context("when getting the access token fails", func() {
			var (
				expectedError error
			)

			BeforeEach(func() {
				expectedError = errors.New("some-access-token-error")

				fakeUAAClient.FetchTokenReturns(nil, expectedError)
			})

			It("returns the error", func() {
				err := alerter.Unhealthy()

				Expect(err).To(Equal(expectedError))
			})
		})

		Context("when sending the email fails", func() {
			var (
				expectedError error
			)

			BeforeEach(func() {
				expectedError = errors.New("some-email-error")

				fakeNotificationsClient.EmailReturns(expectedError)
			})

			It("returns the error", func() {
				err := alerter.Unhealthy()

				Expect(err).To(Equal(expectedError))
			})
		})
	})
})
