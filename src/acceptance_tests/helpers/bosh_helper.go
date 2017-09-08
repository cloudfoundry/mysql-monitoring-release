package helpers

import (
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/cppforlife/go-patch/patch"
)

func GetManifestValue(deploymentManifest string, xPath string) string {
	var in interface{}

	err := yaml.Unmarshal([]byte(deploymentManifest), &in)
	Expect(err).ToNot(HaveOccurred())

	path := patch.MustNewPointerFromString(xPath)

	res, err := patch.FindOp{Path: path}.Apply(in)
	Expect(err).ToNot(HaveOccurred())

	return res.(string)
}
