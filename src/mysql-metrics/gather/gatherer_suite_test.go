package gather_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGather(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gather Suite")
}
