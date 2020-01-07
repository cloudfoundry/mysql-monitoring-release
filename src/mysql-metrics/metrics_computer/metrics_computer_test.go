package metrics_computer_test

import (
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/metrics"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/metrics_computer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricsComputer", func() {
	var (
		metricsComputer     *metrics_computer.MetricsComputer
		metricMappingConfig metrics.MetricMappingConfig
		metricMappings      map[string]metrics.MetricDefinition
		values              map[string]string
		computedMetrics     []*metrics.Metric
	)

	Describe("ComputeMetricsFromMapping", func() {
		BeforeEach(func() {
			metricMappings = map[string]metrics.MetricDefinition{
				"matchingValue": {Key: "/test_metric", Unit: "testUnit"},
			}

			metricMappingConfig = metrics.MetricMappingConfig{}

			metricsComputer = metrics_computer.NewMetricsComputer(metricMappingConfig)
		})

		Context("When the value is a float", func() {
			BeforeEach(func() {
				values = map[string]string{"matchingValue": "1.0"}
				computedMetrics = metricsComputer.ComputeMetricsFromMapping(values, metricMappings)
			})

			It("returns a metric whose value is the given float", func() {
				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/test_metric"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(1.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})

		Context("When the value is a string", func() {
			var expectedValues map[string]float64

			BeforeEach(func() {
				expectedValues = map[string]float64{
					"ON":   1.0,
					"on":   1.0,
					"YES":  1.0,
					"yes":  1.0,
					"OFF":  0.0,
					"off":  0.0,
					"NO":   0.0,
					"no":   0.0,
					"NULL": -1.0,
					"null": -1.0,
				}
			})

			It("properly parses known string values", func() {
				for rawValue, expectedValue := range expectedValues {
					values = map[string]string{"matchingValue": rawValue}
					computedMetrics = metricsComputer.ComputeMetricsFromMapping(values, metricMappings)

					Expect(len(computedMetrics)).To(Equal(1))
					Expect(computedMetrics[0].Key).To(Equal("/test_metric"))
					Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
					Expect(computedMetrics[0].Value).To(Equal(expectedValue))
					Expect(computedMetrics[0].Error).To(BeNil())
				}
			})
		})

		Context("When the value is a string that has meaning as a cluster status", func() {
			var expectedValues map[string]float64

			BeforeEach(func() {
				expectedValues = map[string]float64{
					"primary":      1.0,
					"PRIMARY":      1.0,
					"non-primary":  0.0,
					"NON-PRIMARY":  0.0,
					"disconnected": -1.0,
					"DISCONNECTED": -1.0,
				}
			})

			It("properly parses known string values", func() {
				for rawValue, expectedValue := range expectedValues {
					values = map[string]string{"matchingValue": rawValue}
					computedMetrics = metricsComputer.ComputeMetricsFromMapping(values, metricMappings)

					Expect(len(computedMetrics)).To(Equal(1))
					Expect(computedMetrics[0].Key).To(Equal("/test_metric"))
					Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
					Expect(computedMetrics[0].Value).To(Equal(expectedValue))
					Expect(computedMetrics[0].Error).To(BeNil())
				}
			})
		})

		Context("When the value an unknown string", func() {
			BeforeEach(func() {
				values = map[string]string{"matchingValue": "somethingUnrecognizable"}
				computedMetrics = metricsComputer.ComputeMetricsFromMapping(values, metricMappings)
			})

			It("returns a metric with a value of nil and with an error", func() {
				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/test_metric"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(0.0))
				Expect(computedMetrics[0].Error).To(MatchError("could not convert raw value"))
			})
		})

		Describe("multiple metrics", func() {
			BeforeEach(func() {
				values = map[string]string{
					"matchingValue":      "cannot-parse-this",
					"otherMatchingValue": "cannot-parse-this",
				}
				metricMappings = map[string]metrics.MetricDefinition{
					"matchingValue":      {Key: "/test_metric", Unit: "testUnit"},
					"otherMatchingValue": {Key: "/test_metric2", Unit: "testUnit"},
				}
				computedMetrics = metricsComputer.ComputeMetricsFromMapping(values, metricMappings)
			})

			It("produces a metric for each value even if prior metrics have an error", func() {
				Expect(len(computedMetrics)).To(Equal(2))
				Expect(computedMetrics[0].Error).To(MatchError("could not convert raw value"))
				Expect(computedMetrics[1].Error).To(MatchError("could not convert raw value"))
			})
		})
	})

	Describe("ComputeAvailabilityMetric", func() {
		var (
			mysqlMetricMappings map[string]metrics.MetricDefinition
		)

		BeforeEach(func() {
			mysqlMetricMappings = map[string]metrics.MetricDefinition{
				"available": {Key: "/p.mysql/available", Unit: "testUnit"},
			}
			metricMappingConfig = metrics.MetricMappingConfig{
				MysqlMetricMappings: mysqlMetricMappings,
			}
			metricsComputer = metrics_computer.NewMetricsComputer(metricMappingConfig)
		})

		It("returns an availability metric with the given boolean", func() {
			availabilityMetric := metricsComputer.ComputeAvailabilityMetric(true)
			Expect(availabilityMetric.Key).To(Equal("/p.mysql/available"))
			Expect(availabilityMetric.Unit).To(Equal("testUnit"))
			Expect(availabilityMetric.Value).To(Equal(1.0))

			availabilityMetric = metricsComputer.ComputeAvailabilityMetric(false)
			Expect(availabilityMetric.Key).To(Equal("/p.mysql/available"))
			Expect(availabilityMetric.Unit).To(Equal("testUnit"))
			Expect(availabilityMetric.Value).To(Equal(0.0))
		})
	})

	Describe("ComputeIsFollowerMetric", func() {
		var (
			leaderFollowerMetricMappings map[string]metrics.MetricDefinition
		)

		BeforeEach(func() {
			leaderFollowerMetricMappings = map[string]metrics.MetricDefinition{
				"is_follower": {Key: "/p.mysql/follower/is_follower", Unit: "testUnit"},
			}
			metricMappingConfig = metrics.MetricMappingConfig{
				LeaderFollowerMetricMappings: leaderFollowerMetricMappings,
			}
			metricsComputer = metrics_computer.NewMetricsComputer(metricMappingConfig)
		})

		It("returns an is_follower metric with the given boolean", func() {
			isFollowerMetric := metricsComputer.ComputeIsFollowerMetric(true)
			Expect(isFollowerMetric.Key).To(Equal("/p.mysql/follower/is_follower"))
			Expect(isFollowerMetric.Unit).To(Equal("testUnit"))
			Expect(isFollowerMetric.Value).To(Equal(1.0))

			isFollowerMetric = metricsComputer.ComputeIsFollowerMetric(false)
			Expect(isFollowerMetric.Key).To(Equal("/p.mysql/follower/is_follower"))
			Expect(isFollowerMetric.Unit).To(Equal("testUnit"))
			Expect(isFollowerMetric.Value).To(Equal(0.0))
		})
	})

	Describe("All other metrics", func() {
		var (
			mysqlMetricMappings          map[string]metrics.MetricDefinition
			galeraMetricMappings         map[string]metrics.MetricDefinition
			leaderFollowerMetricMappings map[string]metrics.MetricDefinition
			diskMetricMappings           map[string]metrics.MetricDefinition
			brokerMetricMappings         map[string]metrics.MetricDefinition
			cpuMetricMappings            map[string]metrics.MetricDefinition
		)

		BeforeEach(func() {
			mysqlMetricMappings = map[string]metrics.MetricDefinition{
				"mysql_metric_key": {Key: "/p.mysql/mysql_metric_name", Unit: "testUnit"},
			}

			galeraMetricMappings = map[string]metrics.MetricDefinition{
				"galera_metric_key": {Key: "/p.mysql/galera_metric_name", Unit: "testUnit"},
			}

			leaderFollowerMetricMappings = map[string]metrics.MetricDefinition{
				"leader_follower_metric_key": {Key: "/p.mysql/leader_follower_metric_name", Unit: "testUnit"},
			}

			diskMetricMappings = map[string]metrics.MetricDefinition{
				"disk_metric_key": {Key: "/p.mysql/disk_metric_name", Unit: "testUnit"},
			}

			brokerMetricMappings = map[string]metrics.MetricDefinition{
				"broker_metric_key": {Key: "/p.mysql/broker_metric_name", Unit: "testUnit"},
			}

			cpuMetricMappings = map[string]metrics.MetricDefinition{
				"cpu_metric_key": {Key: "/p.mysql/cpu_metric_name", Unit: "testUnit"},
			}

			metricMappingConfig = metrics.MetricMappingConfig{
				MysqlMetricMappings:          mysqlMetricMappings,
				GaleraMetricMappings:         galeraMetricMappings,
				LeaderFollowerMetricMappings: leaderFollowerMetricMappings,
				DiskMetricMappings:           diskMetricMappings,
				BrokerMetricMappings:         brokerMetricMappings,
				CPUMetricMappings:            cpuMetricMappings,
			}
			metricsComputer = metrics_computer.NewMetricsComputer(metricMappingConfig)
		})

		Describe("ComputeLeaderFollowerMetrics", func() {
			It("Calls through to ComputeMetricsFromMapping", func() {
				values = map[string]string{"leader_follower_metric_key": "123.0"}
				computedMetrics = metricsComputer.ComputeLeaderFollowerMetrics(values)

				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/p.mysql/leader_follower_metric_name"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(123.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})

		Describe("ComputeGlobalMetrics", func() {
			It("Calls through to ComputeMetricsFromMapping", func() {
				values = map[string]string{"mysql_metric_key": "123.0"}
				computedMetrics = metricsComputer.ComputeGlobalMetrics(values)

				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/p.mysql/mysql_metric_name"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(123.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})

		Describe("ComputeDiskMetrics", func() {
			It("Calls through to ComputeMetricsFromMapping", func() {
				values = map[string]string{"disk_metric_key": "123.0"}
				computedMetrics = metricsComputer.ComputeDiskMetrics(values)

				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/p.mysql/disk_metric_name"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(123.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})

		Describe("ComputeGaleraMetrics", func() {
			It("Calls through to ComputeMetricsFromMapping", func() {
				values = map[string]string{"galera_metric_key": "123.0"}
				computedMetrics = metricsComputer.ComputeGaleraMetrics(values)

				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/p.mysql/galera_metric_name"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(123.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})

		Describe("ComputeBrokerMetrics", func() {
			It("Calls through to ComputeMetricsFromMapping", func() {
				values = map[string]string{"broker_metric_key": "123.0"}
				computedMetrics = metricsComputer.ComputeBrokerMetrics(values)

				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/p.mysql/broker_metric_name"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(123.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})

		Describe("ComputeCPUMetrics", func() {
			It("Calls through to ComputeMetricsFromMapping", func() {
				values = map[string]string{"cpu_metric_key": "123.0"}
				computedMetrics = metricsComputer.ComputeCPUMetrics(values)

				Expect(len(computedMetrics)).To(Equal(1))
				Expect(computedMetrics[0].Key).To(Equal("/p.mysql/cpu_metric_name"))
				Expect(computedMetrics[0].Unit).To(Equal("testUnit"))
				Expect(computedMetrics[0].Value).To(Equal(123.0))
				Expect(computedMetrics[0].Error).To(BeNil())
			})
		})
	})
})
