package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const (
	executableTimeout = 5 * time.Second
)

var (
	mysqlDiagBinPath string
)

func TestCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Executable Suite")
}

var _ = BeforeSuite(func() {
	By("Compiling binary")
	var err error
	mysqlDiagBinPath, err = gexec.Build("github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag", "-race")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
