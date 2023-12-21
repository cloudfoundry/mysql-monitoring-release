package hattery_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHattery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hattery Suite")
}
