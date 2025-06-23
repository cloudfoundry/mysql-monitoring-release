package metrics_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-metrics/metrics"
)

var _ = Describe("DefaultMetricsMappingConfig()", func() {
	It("Returns a struct with all the mappings", func() {
		metricMappingConfig := metrics.DefaultMetricMappingConfig()

		mysqlMetricMappings := metricMappingConfig.MysqlMetricMappings
		galeraMetricMappings := metricMappingConfig.GaleraMetricMappings
		leaderFollowerMetricMappings := metricMappingConfig.LeaderFollowerMetricMappings
		diskMetricMappings := metricMappingConfig.DiskUsageMetricMappings
		brokerMetricMappings := metricMappingConfig.BrokerMetricMappings
		cpuMetricMappings := metricMappingConfig.CPUMetricMappings

		Expect(mysqlMetricMappings).ToNot(BeNil())
		Expect(galeraMetricMappings).ToNot(BeNil())
		Expect(leaderFollowerMetricMappings).ToNot(BeNil())
		Expect(diskMetricMappings).ToNot(BeNil())
		Expect(brokerMetricMappings).ToNot(BeNil())
		Expect(cpuMetricMappings).ToNot(BeNil())

		Expect(len(mysqlMetricMappings)).To(Equal(46))
		Expect(len(galeraMetricMappings)).To(Equal(10))
		Expect(len(leaderFollowerMetricMappings)).To(Equal(6))
		Expect(len(diskMetricMappings)).To(Equal(20))
		Expect(len(brokerMetricMappings)).To(Equal(1))
		Expect(len(cpuMetricMappings)).To(Equal(1))
	})
	Describe("docs", func() {
		var metricsDocString string
		var metricMappingConfig *metrics.MetricMappingConfig

		BeforeEach(func() {
			metricMappingConfig = metrics.DefaultMetricMappingConfig()
			listOfMetricsDoc, err := os.ReadFile("../../../docs/list-of-metrics.md")
			Expect(err).NotTo(HaveOccurred())
			metricsDocString = string(listOfMetricsDoc)

		})

		It("have all Mysql Metrics", func() {

			//MysqlMetricMappings          map[string]MetricDefinition
			//GaleraMetricMappings         map[string]MetricDefinition
			//LeaderFollowerMetricMappings map[string]MetricDefinition
			//DiskUsageMetricMappings           map[string]MetricDefinition
			//BrokerMetricMappings         map[string]MetricDefinition
			//CPUMetricMappings
			for _, emittedMetric := range metricMappingConfig.MysqlMetricMappings {
				Expect(metricsDocString).To(ContainSubstring(emittedMetric.Key))
			}

		})

		It("have all Galera Metrics", func() {
			for _, emittedMetric := range metricMappingConfig.GaleraMetricMappings {
				Expect(metricsDocString).To(ContainSubstring(emittedMetric.Key))
			}
		})

		It("have all Leader Follower Metrics", func() {
			for _, emittedMetric := range metricMappingConfig.LeaderFollowerMetricMappings {
				Expect(metricsDocString).To(ContainSubstring(emittedMetric.Key))
			}
		})

		It("have all Disk Metrics", func() {
			for _, emittedMetric := range metricMappingConfig.DiskUsageMetricMappings {
				Expect(metricsDocString).To(ContainSubstring(emittedMetric.Key))
			}
		})

		It("have all Broker Metrics", func() {
			for _, emittedMetric := range metricMappingConfig.BrokerMetricMappings {
				Expect(metricsDocString).To(ContainSubstring(emittedMetric.Key))
			}
		})

		It("have all CPU Metrics", func() {
			for _, emittedMetric := range metricMappingConfig.CPUMetricMappings {
				Expect(metricsDocString).To(ContainSubstring(emittedMetric.Key))
			}
		})
	})
})
