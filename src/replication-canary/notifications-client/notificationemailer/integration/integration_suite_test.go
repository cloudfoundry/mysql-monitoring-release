package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Notifications Integration Suite")
}

var MailinatorToken string

var _ = BeforeSuite(func() {
	t, ok := os.LookupEnv("NOTIFICATIONS_MAILINATOR_TOKEN")
	Expect(ok).To(BeTrue(), "NOTIFICATIONS_MAILINATOR_TOKEN must be set")

	MailinatorToken = t
})
