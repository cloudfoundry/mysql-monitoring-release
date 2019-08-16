package acceptance_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/pivotal/mysql-test-utils/testhelpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MySql monitoring release Acceptance Suite")
}

var TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
var Config *config.Config

var _ = BeforeSuite(func() {
	requiredEnvs := []string{
		"BOSH_ENVIRONMENT",
		"BOSH_CA_CERT",
		"BOSH_CLIENT",
		"BOSH_CLIENT_SECRET",
		"BOSH_DEPLOYMENT",
		"CONFIG",
	}
	testhelpers.CheckForRequiredEnvVars(requiredEnvs)
	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())

	Config = config.LoadConfig()

	TestSetup = workflowhelpers.NewTestSuiteSetup(Config)
	TestSetup.Setup()

	session := cf.Cf("tail", "--help").Wait(10 * time.Second)
	if session.ExitCode() != 0 {
		session = cf.Cf("install-plugin", "-f", "log-cache", "-r", "CF-Community").Wait(10 * time.Minute)
		Expect(session.ExitCode()).To(BeZero())
	}

})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})

var _ = JustAfterEach(func() {
	if CurrentGinkgoTestDescription().Failed {
		fmt.Fprint(GinkgoWriter, testhelpers.TestFailureMessage)
	}
})
