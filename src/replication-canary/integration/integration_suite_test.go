package integration_test

import (
	"database/sql"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Integration Suite")
}

var (
	databaseUser     = "root"
	databasePassword = "password"
	baseDSN          string

	databaseName string
	databaseDSN  string

	pool     *dockertest.Pool
	resource *dockertest.Resource
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
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

	return []byte(resource.GetPort("3306/tcp"))
}, func(data []byte) {

	baseDSN = fmt.Sprintf("root@tcp(localhost:%s)/", string(data))

	db, err := sql.Open("mysql", baseDSN)
	Expect(err).NotTo(HaveOccurred())
	defer db.Close()

	databaseName = fmt.Sprintf("repcanaryintegration%d", GinkgoParallelNode())

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName)
	Expect(err).NotTo(HaveOccurred())

	databaseDSN = baseDSN + databaseName
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	db, err := sql.Open("mysql", baseDSN)
	Expect(err).NotTo(HaveOccurred())
	defer db.Close()

	_, err = db.Exec("DROP DATABASE " + databaseName)
	Expect(err).NotTo(HaveOccurred())
})
