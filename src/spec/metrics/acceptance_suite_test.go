package acceptance_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MySql monitoring release Acceptance Suite")
}

var TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
var Config *config.Config

var _ = BeforeSuite(func() {
	var missingEnvs []string
	for _, v := range []string{
		"BOSH_ENVIRONMENT",
		"BOSH_CA_CERT",
		"BOSH_CLIENT",
		"BOSH_CLIENT_SECRET",
		"BOSH_DEPLOYMENT",
		"CONFIG",
	} {
		if os.Getenv(v) == "" {
			missingEnvs = append(missingEnvs, v)
		}
	}
	Expect(missingEnvs).To(BeEmpty(), "Missing environment variables: %s", strings.Join(missingEnvs, ", "))

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
