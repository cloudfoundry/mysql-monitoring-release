package diskstat

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDiskstats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diskstats Suite")
}
