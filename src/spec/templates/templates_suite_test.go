package templates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

type TemplateContext struct {
	Links      map[string]interface{} `json:"links"`
	Networks   map[string]interface{} `json:"networks"`
	Properties map[string]interface{} `json:"properties"`
}

func TestTemplates(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Templates Suite")
}
