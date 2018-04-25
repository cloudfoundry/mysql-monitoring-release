package clientcreator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClientcreator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Creator Suite")
}
