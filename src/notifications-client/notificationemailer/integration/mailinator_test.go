package integration_test

import (
	"github.com/nu7hatch/gouuid"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/notifications-client/notificationemailer"

	"fmt"
	"os"

	"regexp"
	"strings"
	"time"

	"strconv"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/notifications-client/notificationemailer/integration/mailinator"
)

type fakeLogger struct{}

func (fakeLogger) Debug(string, map[string]interface{}) {}

var _ = Describe("CF Notifications Mailinator Integration Test", func() {
	var (
		client         *notificationemailer.Client
		recipientEmail string

		mailinatorClient *mailinator.Client
		uaaClient        uaa_go_client.Client
		err              error

		testTimeout int

		testLogger fakeLogger
	)

	BeforeEach(func() {
		uaaDomain, ok := os.LookupEnv("NOTIFICATIONS_UAA_DOMAIN")
		if !ok {
			uaaDomain = "uaa.bosh-lite.com"
		}

		notificationsDomain, ok := os.LookupEnv("NOTIFICATIONS_NOTIFICATIONS_DOMAIN")
		if !ok {
			notificationsDomain = "notifications.bosh-lite.com"
		}

		clientUsername, ok := os.LookupEnv("NOTIFICATIONS_CLIENT_USERNAME")
		if !ok {
			clientUsername = "mysql-monitoring"
		}

		clientSecret, ok := os.LookupEnv("NOTIFICATIONS_CLIENT_SECRET")
		if !ok {
			clientSecret = "REPLACE_WITH_CLIENT_SECRET"
		}

		recipientEmail, ok = os.LookupEnv("NOTIFICATIONS_RECIPIENT_EMAIL")
		if !ok {
			recipientEmail = "notifications-integration-test@mailinator.com"
		}

		testTimeoutStr, ok := os.LookupEnv("NOTIFICATIONS_INTEGRATION_TEST_TIMEOUT")
		if !ok {
			testTimeoutStr = "60"
		}

		testTimeout, err = strconv.Atoi(testTimeoutStr)
		Expect(err).NotTo(HaveOccurred())

		cfg := &config.Config{
			ClientName:       clientUsername,
			ClientSecret:     clientSecret,
			UaaEndpoint:      "https://" + uaaDomain,
			SkipVerification: true,
		}

		logger := lager.NewLogger("UAAClient")
		clock := fakeclock.NewFakeClock(time.Now())

		uaaClient, err = uaa_go_client.NewClient(logger, cfg, clock)
		Expect(err).NotTo(HaveOccurred())

		skipSSLCertVerify := true
		testLogger = fakeLogger{}
		client = notificationemailer.NewClient(
			fmt.Sprintf("https://%s", notificationsDomain),
			skipSSLCertVerify,
			testLogger,
		)

		mailinatorClient = mailinator.NewClient(MailinatorToken)
	})

	It("grabs client secret and sends emails", func() {
		uuid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())
		subject := "p-mysql VM restart, alert 111" + uuid.String()
		recipientEmailFragment := strings.Split(recipientEmail, "@")[0]
		html := "{alert-code 111}<br /> This is an e-mail proving that we can automatically send e-mail from a BOSH-deployed job."
		forceUpdate := true

		token, err := uaaClient.FetchToken(forceUpdate)
		Expect(err).NotTo(HaveOccurred())

		err = client.Email(token.AccessToken, recipientEmail, subject, html, "p-mysql")
		Expect(err).NotTo(HaveOccurred())

		hasMailinatorReceivedEmail := func() bool {
			res, err := mailinatorClient.GetMessageList(recipientEmail)
			Expect(err).NotTo(HaveOccurred())

			found := false
			for _, message := range res.Messages {
				matchedSubject, err := regexp.MatchString(".*"+subject, message.Subject)
				Expect(err).NotTo(HaveOccurred())

				if matchedSubject && (message.To == recipientEmailFragment) {
					found = true
					break
				}
			}

			return found
		}

		Eventually(
			hasMailinatorReceivedEmail,
			time.Duration(testTimeout)*time.Second,
			time.Duration(testTimeout/20)*time.Second,
		).Should(BeTrue(), fmt.Sprintf("Did not find message with Subject: %s, To: %s", subject, recipientEmail))
	})
})
