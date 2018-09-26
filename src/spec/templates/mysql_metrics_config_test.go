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

type TemplateContext struct {
	Links      map[string]interface{} `json:"links"`
	Networks   map[string]interface{} `json:"networks"`
	Properties map[string]interface{} `json:"properties"`
}

var _ = Describe("MysqlMetricsConfig", func() {
	var (
		templateContextFile *os.File
		templateContext     *TemplateContext

		contextPath  string
		templatePath string

		templateOutput string

		renderTemplate              func(context *TemplateContext)
		buildDefaultTemplateContext func()
	)

	BeforeEach(func() {
		var err error

		templateContextFile, err = ioutil.TempFile("", "template-context.json")
		Expect(err).NotTo(HaveOccurred())
		contextPath = templateContextFile.Name()

		buildDefaultTemplateContext()

		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		templateDir := filepath.Join(dir, "../../../jobs/mysql-metrics/templates")
		templatePath = filepath.Join(templateDir, "mysql-metrics-config.yml.erb")
	})

	JustBeforeEach(func() {
		renderTemplate(templateContext)
	})

	Context("when required properties are present", func() {
		It("renders properties into JSON/Yaml", func() {
			Expect(templateOutput).NotTo(ContainSubstring("UnknownProperty"))
			Expect(templateOutput).To(ContainSubstring(`"host": "host"`))
			Expect(templateOutput).To(ContainSubstring(`"username": "username"`))
			Expect(templateOutput).To(ContainSubstring(`"password": "password"`))
			Expect(templateOutput).To(ContainSubstring(`"metrics_frequency": 30`))
			Expect(templateOutput).To(ContainSubstring(`"source_id": "source1"`))
			Expect(templateOutput).To(ContainSubstring(`"origin": "origin1"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_broker_metrics": "false"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_disk_metrics": "false"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_cpu_metrics": "false"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_mysql_metrics": "true"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_leader_follower_metrics": "false"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_galera_metrics": "true"`))
			Expect(templateOutput).To(ContainSubstring(`"heartbeat_database": "heartbeat"`))
			Expect(templateOutput).To(ContainSubstring(`"heartbeat_table": "table"`))

			templateContext.Properties = map[string]interface{}{
				"mysql-metrics": map[string]interface{}{
					"host":                            "host2",
					"password":                        "password2",
					"username":                        "username2",
					"metrics_frequency":               31,
					"broker_metrics_enabled":          "true",
					"disk_metrics_enabled":            "true",
					"cpu_metrics_enabled":             "true",
					"mysql_metrics_enabled":           "false",
					"leader_follower_metrics_enabled": "true",
					"galera_metrics_enabled":          "false",
					"heartbeat_database":              "heartbeat2",
					"heartbeat_table":                 "table2",
					"minimum_metrics_frequency":       11,
					"source_id":                       "source1",
					"origin":                          "origin2",
				},
			}

			var err error
			templateContextFile, err = os.Create(contextPath)
			Expect(err).NotTo(HaveOccurred())

			renderTemplate(templateContext)

			Expect(templateOutput).NotTo(ContainSubstring("UnknownProperty"))
			Expect(templateOutput).To(ContainSubstring(`"host": "host2"`))
			Expect(templateOutput).To(ContainSubstring(`"username": "username2"`))
			Expect(templateOutput).To(ContainSubstring(`"password": "password2"`))
			Expect(templateOutput).To(ContainSubstring(`"metrics_frequency": 31`))
			Expect(templateOutput).To(ContainSubstring(`"source_id": "source1"`))
			Expect(templateOutput).To(ContainSubstring(`"origin": "origin2"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_broker_metrics": "true"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_disk_metrics": "true"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_mysql_metrics": "false"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_leader_follower_metrics": "true"`))
			Expect(templateOutput).To(ContainSubstring(`"emit_galera_metrics": "false"`))
			Expect(templateOutput).To(ContainSubstring(`"heartbeat_database": "heartbeat2"`))
			Expect(templateOutput).To(ContainSubstring(`"heartbeat_table": "table2"`))
		})
	})

	Context("when password is not present as a property", func() {
		BeforeEach(func() {
			metricsMap := templateContext.Properties["mysql-metrics"].(map[string]interface{})
			delete(metricsMap, "password")
			templateContext.Properties["mysql-metrics"] = metricsMap
		})

		Context("when a broker link is available", func() {
			BeforeEach(func() {
				templateContext.Links["broker"] = map[string]interface{}{
					"properties": map[string]interface{}{
						"cf_mysql": map[string]interface{}{
							"broker": map[string]interface{}{
								"db_password": "password from link",
							},
						},
					},
					"instances": []map[string]interface{}{{}},
				}
			})

			It("renders the password from the broker link", func() {
				Expect(templateOutput).To(ContainSubstring(`"password": "password from link"`))
			})
		})

		Context("when there is no broker link available", func() {
			It("raises an exception attempting to render", func() {
				Expect(templateOutput).To(ContainSubstring("Password is required"))
			})
		})
	})

	Context("when host is not present as a property", func() {
		BeforeEach(func() {
			metricsMap := templateContext.Properties["mysql-metrics"].(map[string]interface{})
			delete(metricsMap, "host")
		})

		Context("when proxy is available as a link", func() {
			Context("when there are >0 proxies", func() {
				BeforeEach(func() {
					templateContext.Links["proxy"] = map[string]interface{}{
						"properties": map[string]interface{}{},
						"instances": []map[string]interface{}{
							{
								"address": "proxy link address",
							},
						},
					}
				})

				It("renders the host as the IP of the first proxy", func() {
					Expect(templateOutput).To(ContainSubstring(`"host": "proxy link address"`))
				})
			})

			Context("when there are 0 proxies", func() {
				BeforeEach(func() {
					templateContext.Links["proxy"] = map[string]interface{}{
						"properties": map[string]interface{}{},
						"instances":  []map[string]interface{}{},
					}
				})

				It("renders the host as the IP of the first mysql instance", func() {
					Expect(templateOutput).To(ContainSubstring(`"host": "mysql link address"`))
				})
			})
		})

		Context("when there is no proxy link", func() {
			It("renders the host as the IP of the first mysql instance", func() {
				Expect(templateOutput).To(ContainSubstring(`"host": "mysql link address"`))
			})
		})
	})

	Context("when the metrics frequency is too often", func() {
		BeforeEach(func() {
			metricsMap := templateContext.Properties["mysql-metrics"].(map[string]interface{})
			metricsMap["metrics_frequency"] = 1
		})

		It("raises an exception attempting to render", func() {
			Expect(templateOutput).To(ContainSubstring("collecting metrics at this rate is not advised"))
		})
	})

	buildDefaultTemplateContext = func() {
		templateContext = &TemplateContext{}
		templateContext.Networks = map[string]interface{}{}
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
		templateContext.Properties = map[string]interface{}{
			"mysql-metrics": map[string]interface{}{
				"host":                            "host",
				"password":                        "password",
				"username":                        "username",
				"metrics_frequency":               30,
				"source_id":                       "source1",
				"origin":                          "origin1",
				"broker_metrics_enabled":          "false",
				"disk_metrics_enabled":            "false",
				"cpu_metrics_enabled":             "false",
				"mysql_metrics_enabled":           "true",
				"leader_follower_metrics_enabled": "false",
				"galera_metrics_enabled":          "true",
				"heartbeat_database":              "heartbeat",
				"heartbeat_table":                 "table",
				"minimum_metrics_frequency":       10,
			},
		}
	}

	renderTemplate = func(context *TemplateContext) {
		templateContextJson, err := json.Marshal(context)
		Expect(err).NotTo(HaveOccurred())
		_, err = templateContextFile.Write(templateContextJson)
		Expect(err).NotTo(HaveOccurred())

		defer templateContextFile.Close()

		bytes, err := exec.Command("./template", templatePath, contextPath).CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		templateOutput = string(bytes)
	}
})
