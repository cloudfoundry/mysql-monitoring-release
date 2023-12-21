package cpu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCPU(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CPU Suite")
}
