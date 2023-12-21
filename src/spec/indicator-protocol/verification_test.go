package indicator_protocol_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/mysql-monitoring-release/spec/testhelpers"
)

var _ = Describe("Verification", func() {
	Context("Indicator Protocol", func() {
		var fetchIndicatorYaml = func() string {
			tmpFile, err := ioutil.TempFile(os.TempDir(), "indicators-*.yml")
			Expect(err).ToNot(HaveOccurred())
			args := []string{"scp", "mysql/0:/var/vcap/jobs/mysql-metrics/config/indicators.yml", tmpFile.Name()}
			session := testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())
			return tmpFile.Name()
		}

		var fetchCfOauthToken = func() string {
			var token string
			workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
				session := cf.Cf("oauth-token")
				Eventually(session, 10*time.Second).Should(gexec.Exit(0))
				token = string(session.Out.Contents())
				Expect(token).To(ContainSubstring("bearer"))
				session.Terminate()
			})
			token = strings.Replace(token, "\n", "", 1)
			return token
		}

		It("has valid indicator yaml", func() {
			token := fetchCfOauthToken()
			tempFile := fetchIndicatorYaml()
			defer os.Remove(tempFile)

			command := exec.Command(
				"indicator-verification",
				"-k",
				"-authorization",
				token,
				"-indicators",
				tempFile,
				"-query-endpoint",
				// log-cache endpoint is different to the regular cf api endpoint, and is not part of the /v2/info endpoint
				// guess the log-cache url from the CC API url
				fmt.Sprintf("https://%s", strings.Replace(Config.ApiEndpoint, "api", "log-cache", 1)),
				"-metadata",
				fmt.Sprintf("deployment=%s,origin=p_mysql,source_id=p-mysql", os.Getenv("BOSH_DEPLOYMENT")),
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		})

	})
})
