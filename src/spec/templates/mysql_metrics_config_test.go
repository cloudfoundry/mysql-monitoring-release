package templates_test

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"gopkg.in/yaml.v2"
)

var _ = Describe("MysqlMetricsConfig", func() {
	var (
		templateContext *TemplateContext
		templateOutput  string
		templateErr     error
	)

	renderTemplate := func(context *TemplateContext) (templateOutput string, err error) {
		templateContextJson, err := json.Marshal(context)
		if err != nil {
			return "", err
		}

		var output strings.Builder
		cmd := exec.Command("./template",
			"--job=mysql-metrics", "--template=config/mysql-metrics-config.yml", "--context="+string(templateContextJson),
		)
		cmd.Stdout = &output
		cmd.Stderr = io.MultiWriter(&output, GinkgoWriter)
		err = cmd.Run()
		return output.String(), err
	}

	buildDefaultTemplateContext := func() {
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
		templateContext.Properties = map[string]interface{}{}
	}

	BeforeEach(func() {
		buildDefaultTemplateContext()
	})

	JustBeforeEach(func() {
		templateOutput, templateErr = renderTemplate(templateContext)
	})

	Context("when required properties are present", func() {
		BeforeEach(func() {
			templateContext.Properties["mysql-metrics"] = map[string]interface{}{
				"host":     "required-host",
				"password": "required-password",
			}
		})

		It("renders default properties into JSON/Yaml", func() {
			var cfg map[string]any
			Expect(yaml.Unmarshal([]byte(templateOutput), &cfg)).To(Succeed())
			Expect(cfg).To(gstruct.MatchAllKeys(gstruct.Keys{
				"host":                         Equal("required-host"),
				"port":                         Equal(3306),
				"username":                     Equal("mysql-metrics"),
				"password":                     Equal("required-password"),
				"metrics_frequency":            Equal(30),
				"source_id":                    Equal("p-mysql"),
				"origin":                       Equal("p-mysql"),
				"emit_backup_metrics":          Equal(false),
				"emit_broker_metrics":          Equal(false),
				"emit_disk_metrics":            Equal(true),
				"emit_cpu_metrics":             Equal(true),
				"emit_mysql_metrics":           Equal(true),
				"emit_leader_follower_metrics": Equal(false),
				"emit_galera_metrics":          Equal(true),
				"heartbeat_database":           Equal("replication_monitoring"),
				"heartbeat_table":              Equal("heartbeat"),
				"loggregator_ca_path":          Equal("/var/vcap/jobs/mysql-metrics/certs/loggregator-ca.pem"),
				"loggregator_client_key_path":  Equal("/var/vcap/jobs/mysql-metrics/certs/loggregator-client-key.pem"),
				"loggregator_client_cert_path": Equal("/var/vcap/jobs/mysql-metrics/certs/loggregator-client-cert.pem"),
				"instance_id":                  Equal("xxxxxx-xxxxxxxx-xxxxx"),
			}))
		})

		It("renders user provided properties from the job spec", func() {
			templateContext.Properties = map[string]interface{}{
				"mysql-metrics": map[string]interface{}{
					"host":                            "host2",
					"port":                            6033,
					"password":                        "password2",
					"username":                        "username2",
					"metrics_frequency":               31,
					"broker_metrics_enabled":          true,
					"disk_metrics_enabled":            true,
					"cpu_metrics_enabled":             true,
					"mysql_metrics_enabled":           false,
					"backup_metrics_enabled":          true,
					"leader_follower_metrics_enabled": true,
					"galera_metrics_enabled":          false,
					"heartbeat_database":              "heartbeat2",
					"heartbeat_table":                 "table2",
					"minimum_metrics_frequency":       11,
					"source_id":                       "source1",
					"origin":                          "origin2",
				},
			}

			templateOutput, templateErr = renderTemplate(templateContext)
			Expect(templateErr).NotTo(HaveOccurred())

			var cfg map[string]any
			Expect(yaml.Unmarshal([]byte(templateOutput), &cfg)).To(Succeed())
			Expect(cfg).To(gstruct.MatchAllKeys(gstruct.Keys{
				"host":                         Equal("host2"),
				"port":                         Equal(6033),
				"username":                     Equal("username2"),
				"password":                     Equal("password2"),
				"metrics_frequency":            Equal(31),
				"source_id":                    Equal("source1"),
				"origin":                       Equal("origin2"),
				"emit_backup_metrics":          Equal(true),
				"emit_broker_metrics":          Equal(true),
				"emit_disk_metrics":            Equal(true),
				"emit_cpu_metrics":             Equal(true),
				"emit_mysql_metrics":           Equal(false),
				"emit_leader_follower_metrics": Equal(true),
				"emit_galera_metrics":          Equal(false),
				"heartbeat_database":           Equal("heartbeat2"),
				"heartbeat_table":              Equal("table2"),
				"loggregator_ca_path":          Equal("/var/vcap/jobs/mysql-metrics/certs/loggregator-ca.pem"),
				"loggregator_client_key_path":  Equal("/var/vcap/jobs/mysql-metrics/certs/loggregator-client-key.pem"),
				"loggregator_client_cert_path": Equal("/var/vcap/jobs/mysql-metrics/certs/loggregator-client-cert.pem"),
				"instance_id":                  Equal("xxxxxx-xxxxxxxx-xxxxx"),
			}))
		})
	})

	Context("when password is not present as a property", func() {
		BeforeEach(func() {
			templateContext.Properties["mysql-metrics"] = make(map[string]interface{})
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
			templateContext.Properties["mysql-metrics"] = map[string]interface{}{
				"password": "required-password",
			}
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
			templateContext.Properties["mysql-metrics"] = map[string]interface{}{
				"password":          "required-password",
				"metrics_frequency": 1,
			}
		})

		It("raises an exception attempting to render", func() {
			Expect(templateErr).To(HaveOccurred())
			Expect(templateOutput).To(ContainSubstring("collecting metrics at this rate is not advised"))
		})
	})
})
