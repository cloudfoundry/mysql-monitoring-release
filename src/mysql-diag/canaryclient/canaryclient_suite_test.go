package canaryclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCanaryClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Replication Canary Client Suite")
}
