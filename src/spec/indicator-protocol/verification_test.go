package indicator_protocol_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal/mysql-test-utils/testhelpers"
)

var _ = Describe("Verification", func() {
	Context("Indicator Protocol", func() {
		It("Passes the verification CLI", func() {
			var token string
			By("Fetching cf oauth-token")
			workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
				session := cf.Cf("oauth-token")
				Eventually(session.Out, 20*time.Second).Should(gbytes.Say("bearer"))
				token = string(session.Out.Contents())
				session.Terminate()
			})

			By("Fetching the real rendered template off the MySQL VM")
			tmpFile, err := ioutil.TempFile(os.TempDir(), "indicators-*.yml")
			defer os.Remove(tmpFile.Name())
			Expect(err).To(Succeed())
			fmt.Println(tmpFile.Name())
			args := []string{"scp", "mysql/0:/var/vcap/jobs/mysql-metrics/config/indicators.yml", tmpFile.Name()}
			session := testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())

			By("Running the verification CLI with the downloaded template")
			// TODO: This should work on CI and locally, and not assume that the CLI lives in the ~/Downloads folder. We should change this
			verifierPath := "/Users/pivotal/Downloads/indicator-verification-macosx64-0.8.5"
			// TODO: Extract the hardcoded clever-sloth to a system domain env var (or something)
			command := exec.Command(
				verifierPath,
				"-k",
				"-authorization",
				token,
				"-indicators",
				tmpFile.Name(),
				"-query-endpoint",
				"https://log-cache.clever-sloth.pmysql.farm",
				"-metadata",
				"deployment=pxc,origin=p_mysql,source_id=p-mysql",
			)
			// TODO: change the indicators.yml on the bosh vm in an invalid way and observe that this fails
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

	})
})
