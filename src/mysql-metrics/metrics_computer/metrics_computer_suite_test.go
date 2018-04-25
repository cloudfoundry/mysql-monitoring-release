package metrics_computer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMetricsComputer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetricsComputer Suite")
}
