package hattery_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHattery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hattery Suite")
}
