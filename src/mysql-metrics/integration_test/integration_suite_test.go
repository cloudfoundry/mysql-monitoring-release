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
	"github.com/ory/dockertest/v3"
)

var (
	metricsBinPath string
	pool           *dockertest.Pool
	resource       *dockertest.Resource
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

	pool, err = dockertest.NewPool("")
	Expect(err).NotTo(HaveOccurred())

	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "percona/percona-server",
		Tag:        "8.0",
		Env:        []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1"},
	})
	Expect(err).NotTo(HaveOccurred())

	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(localhost:%s)/", resource.GetPort("3306/tcp")))
	Expect(err).NotTo(HaveOccurred())
	Expect(pool.Retry(db.Ping)).To(Succeed())
	Expect(db.Exec(`CHANGE REPLICATION SOURCE TO SOURCE_HOST = 'some-host', SOURCE_USER = 'some-user', SOURCE_PASSWORD = 'some-password'`)).
		Error().NotTo(HaveOccurred())
	Expect(db.Exec(`START REPLICA`)).
		Error().NotTo(HaveOccurred())
	return []byte(resource.GetPort("3306/tcp") + "\t" + metricsBinPath)
}, func(data []byte) {
	info := strings.Fields(string(data))

	mysqlPort = info[0]
	metricsBinPath = info[1]
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
	Expect(pool.Purge(resource)).To(Succeed())
})
