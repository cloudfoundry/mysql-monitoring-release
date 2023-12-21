package database_client_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDatabaseClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DatabaseClient Suite")
}
