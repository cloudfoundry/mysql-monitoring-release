package acceptance_test

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/pivotal/mysql-test-utils/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"time"
)

var _ = Describe("Metrics are received", func() {
	It("correct metrics are emitted within 40s", func() {
		workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
			session := cf.Cf("tail", "-f", "p-mysql")
			Eventually(session.Out, 40 * time.Second).Should(gbytes.Say("/p-mysql/performance/questions:"))
			Eventually(session.Out, 40 * time.Second).Should(gbytes.Say("/p-mysql/galera/wsrep_cluster_status:1"))

			session.Terminate()
		})
	})

	Context("when ephemeral disk (/var/vcap/data) usage increases", func() {
		FIt("the corresponding metric increases", func() {
			//bosh -d pxc ssh mysql/0 -c="sudo stat -f /var/vcap/data"
			args := []string{"-d", "pxc", "ssh", "mysql/0", "-c='sudo stat -f /var/vcap/data'"}
			session := testhelpers.ExecuteBosh(args, 10 * time.Second)
			//Inspect the space available so we can have an idea of how much we can add, decide on some amount to add
			session.Out
			args = []string{"sudo", "fallocate", "-l100000M", "/var/vcap/data/pxc-mysql/tmp/file"}
		})
	})
})
