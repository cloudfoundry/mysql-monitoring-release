package indicator_protocol_test

import (
	"os"
	"strings"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIndicatorProtocol(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IndicatorProtocol Suite")
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

})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})
