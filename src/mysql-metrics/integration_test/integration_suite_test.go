package integration_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	metricsBinPath    string
	executableTimeout = 5 * time.Second
)

func TestMysqlMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	By("Compiling binary")
	var err error
	metricsBinPath, err = gexec.Build("github.com/cloudfoundry/mysql-metrics", "-race")
	if err != nil {
		fmt.Println(err.Error())
	}
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
