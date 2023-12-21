package emit_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEmit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Emit Suite")
}
