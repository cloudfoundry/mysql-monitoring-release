package diagagentclient_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGalerahealthcheckclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diag Agent Client Suite")
}
