package canary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCanary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Canary Suite")
}
