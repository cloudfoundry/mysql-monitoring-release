package integration_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/mysql-metrics/internal/testing/docker"
)

var (
	metricsBinPath string
	resource       string
	mysqlPort      string
)

func TestMysqlMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	By("Compiling binary")
	var err error
	metricsBinPath, err = gexec.Build("github.com/cloudfoundry/mysql-metrics", "-race")
	if err != nil {
		fmt.Println(err.Error())
	}
	Expect(err).ShouldNot(HaveOccurred())

	resource, err = docker.RunContainer(docker.ContainerSpec{
		Image: "percona/percona-server:8.0",
		Ports: []string{"3306/tcp"},
		Env:   []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1"},
	})
	Expect(err).NotTo(HaveOccurred())

	mysqlPort, err := docker.ContainerPort(resource, "3306/tcp")
	Expect(err).NotTo(HaveOccurred())

	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(localhost:%s)/", mysqlPort))
	Expect(err).NotTo(HaveOccurred())
	Eventually(db.Ping, "5m", "1s").Should(Succeed())
	Expect(db.Exec(`CHANGE REPLICATION SOURCE TO SOURCE_HOST = 'some-host', SOURCE_USER = 'some-user', SOURCE_PASSWORD = 'some-password'`)).
		Error().NotTo(HaveOccurred())
	Expect(db.Exec(`START REPLICA`)).
		Error().NotTo(HaveOccurred())

	return []byte(mysqlPort + "\t" + metricsBinPath)
}, func(data []byte) {
	info := strings.Fields(string(data))

	mysqlPort = info[0]
	metricsBinPath = info[1]
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
	Expect(docker.RemoveContainer(resource)).To(Succeed())
})
