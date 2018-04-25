package switchboard_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSwitchboardclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Switchboard Suite")
}
