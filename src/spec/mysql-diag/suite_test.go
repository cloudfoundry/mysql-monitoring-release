package mysql_diag_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/mysql-test-utils/testhelpers"
)

func TestMySQLDiag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MySQLDiag Suite")
}

var _ = BeforeSuite(func() {
	requiredEnvs := []string{
		"BOSH_ENVIRONMENT",
		"BOSH_CA_CERT",
		"BOSH_CLIENT",
		"BOSH_CLIENT_SECRET",
		"BOSH_DEPLOYMENT",
	}
	testhelpers.CheckForRequiredEnvVars(requiredEnvs)
})
