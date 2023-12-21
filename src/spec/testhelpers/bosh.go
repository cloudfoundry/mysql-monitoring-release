package testhelpers

import (
	"os/exec"
	"time"

	"github.com/cloudfoundry/cf-test-helpers/v2/commandreporter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func ExecuteBosh(args []string, timeout time.Duration) *gexec.Session {
	command := exec.Command("bosh", args...)
	reporter := commandreporter.NewCommandReporter(GinkgoWriter)
	reporter.Report(time.Now(), command)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	session.Wait(timeout)
	return session
}
