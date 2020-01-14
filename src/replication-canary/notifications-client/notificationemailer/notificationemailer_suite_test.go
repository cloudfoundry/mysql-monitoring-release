package notificationemailer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNotificationemailer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Notificationemailer Suite")
}
