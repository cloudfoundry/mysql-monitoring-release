package templates_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

var _ = Describe("Indicators", func() {
	type Threshold struct {
		Level       string `yaml:"level"`
		GreaterThan int    `yaml:"gt,omitempty"`
		LessThan    int    `yaml:"lt,omitempty"`
		EqualTo     int    `yaml:"eq,omitempty"`
	}

	type Indicator struct {
		Name       string      `yaml:"name"`
		Thresholds []Threshold `yaml:"thresholds"`
	}

	type config struct {
		Indicators []Indicator `yaml:"indicators"`
	}

	var clusterSizeThreshold = func(level string, c config) int {
		for _, indicator := range c.Indicators {
			if indicator.Name == "mysql_galera_cluster_size" {
				for _, threshold := range indicator.Thresholds {
					if threshold.Level == level {
						return threshold.LessThan
					}
				}
			}
		}
		return -1
	}

	var renderTemplate = func(context *TemplateContext) string {
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

	var renderIndicatorTemplateWithOriginHavingValue = func(origin string) string {
		templateContext := &TemplateContext{}

		templateContext.Properties = map[string]interface{}{
			"mysql-metrics": map[string]interface{}{
				"source_id": "source1",
				"origin":    origin,
			},
		}

		templateContext.Links = map[string]interface{}{
			"mysql": map[string]interface{}{
				"properties": map[string]interface{}{},
				"instances": []map[string]interface{}{
					{
						"address": "mysql link address",
					},
				},
			},
		}

		return renderTemplate(templateContext)
	}

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

	Describe("Galera Cluster Size", func() {
		Context("When the node size is 1", func() {
			It("Emits a critical alert when the node size is less than 1", func() {
				templateContext := &TemplateContext{}

				templateContext.Properties = map[string]interface{}{
					"mysql-metrics": map[string]interface{}{
						"source_id": "source",
						"origin":    "origin",
					},
				}
				templateContext.Links = map[string]interface{}{
					"mysql": map[string]interface{}{
						"properties": map[string]interface{}{},
						"instances": []map[string]interface{}{
							{
								"address": "mysql link address",
							},
						},
					},
				}

				a := renderTemplate(templateContext)
				var c config
				err := yaml.Unmarshal([]byte(a), &c)
				Expect(err).NotTo(HaveOccurred())
				criticalThreshold := clusterSizeThreshold("critical", c)

				Expect(criticalThreshold).To(Equal(1))
			})

			It("Does not emit a warning", func() {
				templateContext := &TemplateContext{}

				templateContext.Properties = map[string]interface{}{
					"mysql-metrics": map[string]interface{}{
						"source_id": "source",
						"origin":    "origin",
					},
				}
				templateContext.Links = map[string]interface{}{
					"mysql": map[string]interface{}{
						"properties": map[string]interface{}{},
						"instances": []map[string]interface{}{
							{
								"address": "mysql link address",
							},
						},
					},
				}

				a := renderTemplate(templateContext)
				var c config
				err := yaml.Unmarshal([]byte(a), &c)
				Expect(err).NotTo(HaveOccurred())
				warningThreshold := clusterSizeThreshold("warning", c)

				Expect(warningThreshold).To(Equal(-1))
			})
		})

		Context("When the node size is 3 or greater", func() {

			It("Emits a critical alert when the node size is less than 2", func() {
				templateContext := &TemplateContext{}

				templateContext.Properties = map[string]interface{}{
					"mysql-metrics": map[string]interface{}{
						"source_id": "source",
						"origin":    "origin",
					},
				}
				templateContext.Links = map[string]interface{}{
					"mysql": map[string]interface{}{
						"properties": map[string]interface{}{},
						"instances": []map[string]interface{}{
							{
								"address": "mysql link address",
							},
							{
								"address": "mysql link address",
							},
							{
								"address": "mysql link address",
							},
						},
					},
				}

				a := renderTemplate(templateContext)
				var c config
				err := yaml.Unmarshal([]byte(a), &c)
				Expect(err).NotTo(HaveOccurred())
				criticalThreshold := clusterSizeThreshold("critical", c)
				Expect(criticalThreshold).To(Equal(2))
			})
			It("Emits a warning when the node size is less than was configured", func() {
				templateContext := &TemplateContext{}

				templateContext.Properties = map[string]interface{}{
					"mysql-metrics": map[string]interface{}{
						"source_id": "source",
						"origin":    "origin",
					},
				}
				templateContext.Links = map[string]interface{}{
					"mysql": map[string]interface{}{
						"properties": map[string]interface{}{},
						"instances": []map[string]interface{}{
							{
								"address": "mysql link address",
							},
							{
								"address": "mysql link address",
							},
							{
								"address": "mysql link address",
							},
						},
					},
				}

				a := renderTemplate(templateContext)
				var c config
				err := yaml.Unmarshal([]byte(a), &c)
				Expect(err).NotTo(HaveOccurred())
				warningThreshold := clusterSizeThreshold("warning", c)
				Expect(warningThreshold).To(Equal(3))
			})
		})
	})

	XDescribe("mysql_galera_wsrep_ready", func() {
		It("mysql_galera_wsrep_ready", func() {

			templateContext := &TemplateContext{}

			templateContext.Properties = map[string]interface{}{
				"mysql-metrics": map[string]interface{}{
					"source_id": "source",
					"origin":    "origin",
				},
			}
			templateContext.Links = map[string]interface{}{
				"mysql": map[string]interface{}{
					"properties": map[string]interface{}{},
					"instances": []map[string]interface{}{
						{
							"address": "mysql link address",
						},
					},
				},
			}

			a := renderTemplate(templateContext)

			Expect(a).ToNot(ContainSubstring(`mysql_galera_wsrep_ready`))
		})
	})
})
