package helpers

import (
	"database/sql"
	"fmt"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

func GetEnvVar(envVarName string) string {
	value, found := os.LookupEnv(envVarName)
	if !found {
		ginkgo.Fail(fmt.Sprintf("Expected to find environment variable, %s", envVarName))
	}
	return value
}

func GetMysqlConnection(hostIp string, port int, username, password, databaseName string) *sql.DB {
	conn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		username, password, hostIp, port, databaseName))
	Expect(err).NotTo(HaveOccurred())

	return conn
}
