package galera_agent_client

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGaleraAgentClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Galera Agent Client Suite")
}
