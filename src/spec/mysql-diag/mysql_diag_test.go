package mysql_diag_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/mysql-test-utils/testhelpers"
)

var _ = Describe("MySQLDiag", func() {
	It("emits diagnostic output", func() {
		args := []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
		session := testhelpers.ExecuteBosh(args, 10*time.Second)
		Expect(session.ExitCode()).To(BeZero())
		Expect(session).To(SatisfyAll(
			gbytes.Say(`Checking canary status\.\.\. .*healthy.*`),
			gbytes.Say(`Synced\s+\|\s+Primary`),
			gbytes.Say(`PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
		))
	})
})
