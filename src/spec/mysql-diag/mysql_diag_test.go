package mysql_diag_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/mysql-monitoring-release/spec/testhelpers"
)

var _ = Describe("MySQLDiag", func() {
	It("emits diagnostic output", func() {
		args := []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
		session := testhelpers.ExecuteBosh(args, 10*time.Second)
		Expect(session.ExitCode()).To(BeZero())
		Expect(session).To(SatisfyAll(
			gbytes.Say(`(Checking canary status\.\.\. .*healthy.*)|(Canary not configured)`),
			gbytes.Say(`SEQNO\s+\|PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
			gbytes.Say(`Synced\s+\|\s+Primary\s+\|\s+[0-9]+\s+|`),
		))
	})
	When("a node is offline", func() {
		BeforeEach(func() {
			args := []string{"ssh", "mysql/1", "--command=\"sudo monit stop galera-init\""}
			session := testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())
		})
		AfterEach(func() {
			args := []string{"ssh", "mysql/1", "--command=\"sudo monit start galera-init\""}
			session := testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())
			Eventually(func() *gbytes.Buffer {
				args = []string{"ssh", "mysql/1", "--command=\"sudo monit summary | grep galera-init\""}
				session = testhelpers.ExecuteBosh(args, 10*time.Second)
				return session.Out
			}, "2m", "1s").Should(gbytes.Say(`Process 'galera-init'\s+running`))
		})
		It("emits diagnostic output", func() {
			args := []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
			session := testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())
			Expect(session).To(SatisfyAll(
				gbytes.Say(`(Checking canary status\.\.\. .*healthy.*)|(Canary not configured)`),
				gbytes.Say(`SEQNO\s+\|PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
				gbytes.Say(` N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|\s+[0-9]+\s+|`),
			))
		})
	})
})
