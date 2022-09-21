package templates_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Indicators", func() {
	type Threshold struct {
		Level    string `yaml:"level"`
		Operator string `yaml:"operator,omitempty"`
		Value    int    `yaml:"value,omitempty"`
	}

	type Indicator struct {
		Name       string      `yaml:"name"`
		Thresholds []Threshold `yaml:"thresholds"`
	}

	type config struct {
		Spec struct {
			Indicators []Indicator `yaml:"indicators"`
		} `yaml:"spec"`
	}

	var templateContext *TemplateContext

	var getThresholdByNameAndLevel = func(c config, name, level string) Threshold {
		for _, indicator := range c.Spec.Indicators {
			if indicator.Name == name {
				for _, threshold := range indicator.Thresholds {
					if threshold.Level == level {
						return threshold
					}
				}
			}
		}
		return Threshold{}
	}

	var renderTemplate = func(context *TemplateContext) string {
		templateContextJson, err := json.Marshal(context)
		Expect(err).NotTo(HaveOccurred())
		bytes, err := exec.Command("./template",
			"--job=mysql-metrics",
			"--template=config/indicators.yml",
			"--context="+string(templateContextJson),
		).CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		return string(bytes)
	}

	var renderIndicatorTemplateWithOriginHavingValue = func(origin string) string {
		templateContext.Properties["mysql-metrics"].(map[string]interface{})["origin"] = origin

		return renderTemplate(templateContext)
	}

	BeforeEach(func() {
		templateContext = &TemplateContext{}

		templateContext.Properties = map[string]interface{}{
			"mysql-metrics": map[string]interface{}{
				"source_id":                "source",
				"origin":                   "origin",
				"galera_metrics_enabled":   true,
				"backup_schedule_in_hours": 8,
				"backup_metrics_enabled":   false,
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
	})

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

	Describe("backups", func() {
		var (
			indicatorConfig config
		)

		Context("when backup metrics are emitted", func() {
			BeforeEach(func() {
				templateContext.Properties["mysql-metrics"].(map[string]interface{})["backup_metrics_enabled"] = true
			})

			It("renders a backup indicator with the right threshold", func() {
				a := renderTemplate(templateContext)
				err := yaml.Unmarshal([]byte(a), &indicatorConfig)
				Expect(err).NotTo(HaveOccurred())

				var backupIndicator Indicator

				for _, indicator := range indicatorConfig.Spec.Indicators {
					if indicator.Name == "mysql_backups" {
						backupIndicator = indicator
					}
				}

				Expect(backupIndicator).ToNot(Equal(Indicator{}))
				Expect(backupIndicator.Thresholds[0].Value).To(Equal(8))
			})
		})

		Context("when backup metrics are not emitted", func() {
			BeforeEach(func() {
				templateContext.Properties["mysql-metrics"].(map[string]interface{})["backup_metrics_enabled"] = false
			})

			It("does not render a backup indicator", func() {
				a := renderTemplate(templateContext)
				err := yaml.Unmarshal([]byte(a), &indicatorConfig)
				Expect(err).NotTo(HaveOccurred())

				for _, indicator := range indicatorConfig.Spec.Indicators {
					Expect(indicator.Name).ToNot(Equal("mysql_backups"))
				}
			})
		})

	})

	Describe("Non Galera", func() {
		var (
			indicatorConfig config
		)

		BeforeEach(func() {
			templateContext.Properties["mysql-metrics"].(map[string]interface{})["galera_metrics_enabled"] = false

			a := renderTemplate(templateContext)
			err := yaml.Unmarshal([]byte(a), &indicatorConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not have mysql_galera_wsrep_ready indicator", func() {
			for _, indicator := range indicatorConfig.Spec.Indicators {
				Expect(indicator.Name).ToNot(Equal("mysql_galera_wsrep_ready"))
			}
		})

		It("should not have mysql_galera_cluster_size indicator", func() {
			for _, indicator := range indicatorConfig.Spec.Indicators {
				Expect(indicator.Name).ToNot(Equal("mysql_galera_cluster_size"))
			}
		})

		Context("and there are 3 or more nodes", func() {
			BeforeEach(func() {
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
				err := yaml.Unmarshal([]byte(a), &indicatorConfig)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not have mysql_galera_cluster_status indicator", func() {
				for _, indicator := range indicatorConfig.Spec.Indicators {
					Expect(indicator.Name).ToNot(Equal("mysql_galera_cluster_status"))
				}
			})
		})
	})

	Describe("Galera Cluster", func() {
		Describe("Size", func() {
			Context("When the node size is 1", func() {
				It("Emits a critical alert when the node size is less than 1", func() {
					a := renderTemplate(templateContext)
					var c config
					err := yaml.Unmarshal([]byte(a), &c)
					Expect(err).NotTo(HaveOccurred())
					criticalThreshold := getThresholdByNameAndLevel(c, "mysql_galera_cluster_size", "critical")

					Expect(criticalThreshold.Operator).To(Equal("lt"))
					Expect(criticalThreshold.Value).To(Equal(1))
				})

				It("Does not emit a warning", func() {
					a := renderTemplate(templateContext)
					var c config
					err := yaml.Unmarshal([]byte(a), &c)
					Expect(err).NotTo(HaveOccurred())
					warningThreshold := getThresholdByNameAndLevel(c, "mysql_galera_cluster_size", "warning")
					Expect(warningThreshold).To(BeZero())
				})
			})

			Context("When the node size is 3 or greater", func() {
				BeforeEach(func() {
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
				})

				It("Emits a critical alert when the node size is less than 2", func() {
					a := renderTemplate(templateContext)
					var c config
					err := yaml.Unmarshal([]byte(a), &c)
					Expect(err).NotTo(HaveOccurred())
					criticalThreshold := getThresholdByNameAndLevel(c, "mysql_galera_cluster_size", "critical")
					Expect(criticalThreshold.Operator).To(Equal("lt"))
					Expect(criticalThreshold.Value).To(Equal(2))
				})

				It("Emits a warning when the node size is less than was configured", func() {
					a := renderTemplate(templateContext)
					var c config
					err := yaml.Unmarshal([]byte(a), &c)
					Expect(err).NotTo(HaveOccurred())
					warningThreshold := getThresholdByNameAndLevel(c, "mysql_galera_cluster_size", "warning")
					Expect(warningThreshold.Operator).To(Equal("lt"))
					Expect(warningThreshold.Value).To(Equal(3))
				})
			})
		})

		Describe("Status", func() {
			var (
				indicatorConfig config
			)

			Context("the instances count is less than 3", func() {
				BeforeEach(func() {
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
					err := yaml.Unmarshal([]byte(a), &indicatorConfig)
					Expect(err).NotTo(HaveOccurred())
				})

				It("is not provided as an indicator", func() {
					for _, indicator := range indicatorConfig.Spec.Indicators {
						Expect(indicator.Name).ToNot(Equal("mysql_galera_cluster_status"))
					}
				})
			})

			Context("the instances count is greater than or equal to 3", func() {
				BeforeEach(func() {
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
					err := yaml.Unmarshal([]byte(a), &indicatorConfig)
					Expect(err).NotTo(HaveOccurred())
				})

				It("is provided as an indicator", func() {
					var found bool
					for _, indicator := range indicatorConfig.Spec.Indicators {
						if indicator.Name == "mysql_galera_cluster_status" {
							found = true
							break
						}
					}

					Expect(found).To(BeTrue(), "Did not find mysql_galera_cluster_status in the indicators")
				})
			})
		})
	})
})
