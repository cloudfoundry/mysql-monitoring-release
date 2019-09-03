package templates_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indicators", func() {
	Describe("origin is sanitized for any non-alphanumeric character into an underscore", func() {
		It("sanitizes -", func() {
			Expect(renderIndicatorTemplateWithOriginHavingValue("origin-1")).To(ContainSubstring(`origin: origin_1`))
		})

		It("sanitizes %", func() {
			Expect(renderIndicatorTemplateWithOriginHavingValue("origin%1")).To(ContainSubstring(`origin: origin_1`))
		})

		It("Allows camel case %", func() {
			Expect(renderIndicatorTemplateWithOriginHavingValue("Origin1")).To(ContainSubstring(`origin: Origin1`))
		})
	})
})

func renderTemplate(context *TemplateContext) string {
	templateContextFile, err := ioutil.TempFile("", "template-context.json")
	Expect(err).NotTo(HaveOccurred())
	contextPath := templateContextFile.Name()

	templateContextJson, err := json.Marshal(context)
	Expect(err).NotTo(HaveOccurred())
	_, err = templateContextFile.Write(templateContextJson)
	Expect(err).NotTo(HaveOccurred())

	defer templateContextFile.Close()

	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	templateDir := filepath.Join(dir, "../../../jobs/mysql-metrics/templates")
	templatePath := filepath.Join(templateDir, "indicators.yml.erb")

	bytes, err := exec.Command("./template", templatePath, contextPath).CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
	return string(bytes)
}

func renderIndicatorTemplateWithOriginHavingValue(origin string) string {
	templateContext := &TemplateContext{}
	templateContext.Properties = map[string]interface{}{
		"mysql-metrics": map[string]interface{}{
			"source_id": "source1",
			"origin":    "origin1",
		},
	}

	templateContext.Properties = map[string]interface{}{
		"mysql-metrics": map[string]interface{}{
			"source_id": "source1",
			"origin":    origin,
		},
	}

	return renderTemplate(templateContext)
}
