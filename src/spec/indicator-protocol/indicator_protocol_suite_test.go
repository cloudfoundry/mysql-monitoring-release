package indicator_protocol_test

import (
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/mysql-test-utils/testhelpers"
)

func TestIndicatorProtocol(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IndicatorProtocol Suite")
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

})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})
