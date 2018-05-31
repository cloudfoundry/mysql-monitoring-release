package metrics_test

import (
	metrics "mysql-metrics/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultMetricsMappingConfig()", func() {
	It("Returns a struct with all the mappings", func() {
		metricMappingConfig := metrics.DefaultMetricMappingConfig()

		mysqlMetricMappings := metricMappingConfig.MysqlMetricMappings
		galeraMetricMappings := metricMappingConfig.GaleraMetricMappings
		leaderFollowerMetricMappings := metricMappingConfig.LeaderFollowerMetricMappings
		diskMetricMappings := metricMappingConfig.DiskMetricMappings
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
		Expect(len(diskMetricMappings)).To(Equal(12))
		Expect(len(brokerMetricMappings)).To(Equal(1))
		Expect(len(cpuMetricMappings)).To(Equal(1))
	})
})
