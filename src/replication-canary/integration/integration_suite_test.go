package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"os"
	"testing"
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
)

var _ = SynchronizedBeforeSuite(func() []byte { return nil }, func(data []byte) {
	if env, ok := os.LookupEnv("MYSQL_USER"); ok {
		databaseUser = env
	}
	if env, ok := os.LookupEnv("MYSQL_PASSWORD"); ok {
		databasePassword = env
	}

	baseDSN = databaseUser + ":" + databasePassword + "@/"

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
