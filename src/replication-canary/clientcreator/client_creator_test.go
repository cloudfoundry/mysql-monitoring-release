package clientcreator_test

import (
	"errors"

	"github.com/cloudfoundry/replication-canary/alert/alertfakes"
	"github.com/cloudfoundry/replication-canary/clientcreator"
	"github.com/cloudfoundry/replication-canary/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Creator", func() {

	var (
		fakeUAAClient *alertfakes.FakeUAAClient
		rootConfig    *config.Config
	)

	BeforeEach(func() {
		fakeUAAClient = new(alertfakes.FakeUAAClient)

		rootConfig = &config.Config{
			Notifications: config.Notifications{
				AdminClientUsername: "adminsomething",
				AdminClientSecret:   "adminsome-secret",
				ClientUsername:      "something",
				ClientSecret:        "some-secret",
				NotificationsDomain: "some-notifications-domain.bosh-lite.com",
				UAADomain:           "some-uaa-domain.bosh-lite.com",
				ToAddress:           "to-address@example.com",
			},
		}
	})

	Describe("CreateClient", func() {
		It("registers a uaa client from the config", func() {
			err := clientcreator.CreateClient(fakeUAAClient, rootConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeUAAClient.RegisterOauthClientCallCount()).To(Equal(1))
			uaaOauthClient := fakeUAAClient.RegisterOauthClientArgsForCall(0)
			Expect(uaaOauthClient.ClientId).To(Equal(rootConfig.Notifications.ClientUsername))
			Expect(uaaOauthClient.ClientSecret).To(Equal(rootConfig.Notifications.ClientSecret))
			Expect(uaaOauthClient.Authorities).To(Equal([]string{"notifications.write", "critical_notifications.write", "emails.write"}))
		})

		It("returns an error when the registration fails", func() {
			fakeUAAClient.RegisterOauthClientReturns(nil, errors.New("broken"))
			err := clientcreator.CreateClient(fakeUAAClient, rootConfig)
			Expect(err).To(HaveOccurred())
		})
	})
})
