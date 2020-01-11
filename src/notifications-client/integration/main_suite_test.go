package main_test

import (
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const (
	executableTimeout = 5 * time.Second
)

var (
	binPath string

	uaaDomain              string
	notificationsDomain    string
	uaaClientUsername      string
	uaaClientSecret        string
	uaaAdminClientUsername string
	uaaAdminClientSecret   string

	mailinatorToken string
	recipientEmail  string
	testTimeout     int

	skipSSLCertVerify bool
)

func TestUpgrader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executable Suite")
}

var _ = BeforeSuite(func() {
	var err error
	binPath, err = gexec.Build("notifications-client")
	Expect(err).NotTo(HaveOccurred())

	t, ok := os.LookupEnv("NOTIFICATIONS_MAILINATOR_TOKEN")
	Expect(ok).To(BeTrue(), "NOTIFICATIONS_MAILINATOR_TOKEN must be set")
	Expect(t).NotTo(BeEmpty())

	mailinatorToken = t

	uaaDomain, ok = os.LookupEnv("NOTIFICATIONS_UAA_DOMAIN")
	if !ok {
		uaaDomain = "uaa.bosh-lite.com"
	}

	notificationsDomain, ok = os.LookupEnv("NOTIFICATIONS_NOTIFICATIONS_DOMAIN")
	if !ok {
		notificationsDomain = "notifications.bosh-lite.com"
	}

	uaaAdminClientUsername, ok = os.LookupEnv("NOTIFICATIONS_UAA_ADMIN_CLIENT_USERNAME")
	if !ok {
		uaaAdminClientUsername = "admin"
	}

	uaaAdminClientSecret, ok = os.LookupEnv("NOTIFICATIONS_UAA_ADMIN_CLIENT_SECRET")
	if !ok {
		uaaAdminClientSecret = "admin-secret"
	}

	uaaClientUsername, ok = os.LookupEnv("NOTIFICATIONS_CLIENT_USERNAME")
	if !ok {
		uaaClientUsername = "mysql-monitoring"
	}

	uaaClientSecret, ok = os.LookupEnv("NOTIFICATIONS_CLIENT_SECRET")
	if !ok {
		uaaClientSecret = "REPLACE_WITH_CLIENT_SECRET"
	}

	recipientEmail, ok = os.LookupEnv("NOTIFICATIONS_RECIPIENT_EMAIL")
	if !ok {
		recipientEmail = "notifications-integration-test@mailinator.com"
	}

	skipSSLCertVerifyValue, ok := os.LookupEnv("NOTIFICATIONS_SKIP_SSL_CERT_VERIFY")
	if !ok {
		skipSSLCertVerify = true
	} else {
		skipSSLCertVerify, err = strconv.ParseBool(skipSSLCertVerifyValue)
		Expect(err).NotTo(HaveOccurred())
	}

	testTimeoutStr, ok := os.LookupEnv("NOTIFICATIONS_INTEGRATION_TEST_TIMEOUT")
	if !ok {
		testTimeoutStr = "60"
	}

	testTimeout, err = strconv.Atoi(testTimeoutStr)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
