package config_test

import (
	"fmt"

	"github.com/cloudfoundry/replication-canary/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		rootConfig    *config.Config
		configuration string
	)

	JustBeforeEach(func() {
		osArgs := []string{
			"replication-canary",
			fmt.Sprintf("-config=%s", configuration),
		}

		var err error
		rootConfig, err = config.NewConfig(osArgs)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Validate", func() {
		BeforeEach(func() {
			configuration = `{
				"PidFile": fakePath,
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
			}`
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
})
