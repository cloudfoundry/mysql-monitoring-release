package integration_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/replication-canary/internal/testing/docker"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Integration Suite")
}

var (
	baseDSN      string
	databaseName string
	databaseDSN  string
	resource     string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
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

	return []byte(mysqlPort)
}, func(data []byte) {

	baseDSN = fmt.Sprintf("root@tcp(localhost:%s)/", string(data))

	db, err := sql.Open("mysql", baseDSN)
	Expect(err).NotTo(HaveOccurred())
	defer func() { _ = db.Close() }()

	databaseName = fmt.Sprintf("repcanaryintegration%d", GinkgoParallelProcess())

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName)
	Expect(err).NotTo(HaveOccurred())

	databaseDSN = baseDSN + databaseName
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	db, err := sql.Open("mysql", baseDSN)
	Expect(err).NotTo(HaveOccurred())
	defer func() { _ = db.Close() }()

	_, err = db.Exec("DROP DATABASE " + databaseName)
	Expect(err).NotTo(HaveOccurred())
})
