package galera_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGalera(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Galera Suite")
}
