package main_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	executableTimeout = 5 * time.Second
)

var (
	binPath string
)

func TestCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Executable Suite")
}

var _ = BeforeSuite(func() {
	By("Compiling binary")
	var err error
	binPath, err = gexec.Build(
		"github.com/cloudfoundry/mysql-diag/mysql-diag-agent",
		"-race",
	)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
