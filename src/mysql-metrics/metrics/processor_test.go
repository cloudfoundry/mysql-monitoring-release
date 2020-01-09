package metrics_test

import (
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/config"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/metrics"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/metrics/metricsfakes"
	"github.com/hashicorp/go-multierror"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processor", func() {
	var (
		processor           metrics.Processor
		fakeGatherer        *metricsfakes.FakeGatherer
		fakeMetricsComputer *metricsfakes.FakeMetricsComputer
		fakeMetricsWriter   *metricsfakes.FakeWriter
		configuration       *config.Config
	)

	BeforeEach(func() {
		fakeGatherer = &metricsfakes.FakeGatherer{}
		fakeMetricsComputer = &metricsfakes.FakeMetricsComputer{}
		fakeMetricsWriter = &metricsfakes.FakeWriter{}
		configuration = &config.Config{}

		processor = metrics.NewProcessor(
			fakeGatherer,
			fakeMetricsComputer,
			fakeMetricsWriter,
			configuration,
		)
	})

	Describe("Process()", func() {
		Context("when broker metrics are enabled", func() {
			BeforeEach(func() {
				configuration.EmitBrokerMetrics = true
			})

			It("returns broker metrics", func() {
				brokerStatsReturn := map[string]string{
					"service_plans_disk_allocated": "200",
				}

				servicePlansDiskAllocatedMetric := &metrics.Metric{
					Key:   "ephemeral_disk_free",
					Value: 1024,
				}

				fakeMetricsComputer.ComputeBrokerMetricsReturnsOnCall(0, []*metrics.Metric{servicePlansDiskAllocatedMetric})
				fakeGatherer.BrokerStatsReturns(brokerStatsReturn, nil)
				err := processor.Process()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeMetricsComputer.ComputeBrokerMetricsCallCount()).To(Equal(1))
				computeBrokerMetricsArgs := fakeMetricsComputer.ComputeBrokerMetricsArgsForCall(0)
				Expect(computeBrokerMetricsArgs).To(Equal(brokerStatsReturn))

				Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
				metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
				Expect(len(metricsToEmit)).To(Equal(1))
				Expect(metricsToEmit[0]).To(Equal(servicePlansDiskAllocatedMetric))
			})
		})

		Context("when cpu metrics are enabled", func() {
			BeforeEach(func() {
				configuration.EmitCPUMetrics = true
			})

			It("returns cpu metrics", func() {
				cpuStatsReturn := map[string]string{
					"cpu_utilization_percent": "71",
				}

				cpuUtilizationMetric := &metrics.Metric{
					Key:   "cpu_utilization_percent",
					Value: 71,
				}

				fakeMetricsComputer.ComputeCPUMetricsReturnsOnCall(0, []*metrics.Metric{cpuUtilizationMetric})
				fakeGatherer.CPUStatsReturns(cpuStatsReturn, nil)
				err := processor.Process()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeMetricsComputer.ComputeCPUMetricsCallCount()).To(Equal(1))
				computeCPUMetricsArgs := fakeMetricsComputer.ComputeCPUMetricsArgsForCall(0)
				Expect(computeCPUMetricsArgs).To(Equal(cpuStatsReturn))

				Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
				metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
				Expect(len(metricsToEmit)).To(Equal(1))
				Expect(metricsToEmit[0]).To(Equal(cpuUtilizationMetric))
			})
		})

		Context("when disk metrics are enabled", func() {
			BeforeEach(func() {
				configuration.EmitDiskMetrics = true
			})

			It("returns disk metrics", func() {
				diskStatsReturn := map[string]string{
					"ephemeral_disk_free": "1024",
				}

				diskMetric := &metrics.Metric{
					Key:   "ephemeral_disk_free",
					Value: 1024,
				}

				fakeMetricsComputer.ComputeDiskMetricsReturnsOnCall(0, []*metrics.Metric{diskMetric})
				fakeGatherer.DiskStatsReturns(diskStatsReturn, nil)
				err := processor.Process()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeMetricsComputer.ComputeDiskMetricsCallCount()).To(Equal(1))
				computeDiskMetricsArgs := fakeMetricsComputer.ComputeDiskMetricsArgsForCall(0)
				Expect(computeDiskMetricsArgs).To(Equal(diskStatsReturn))

				Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
				metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
				Expect(len(metricsToEmit)).To(Equal(1))
				Expect(metricsToEmit[0]).To(Equal(diskMetric))
			})
		})

		Context("when the database is available", func() {
			BeforeEach(func() {
				isAvailableReturns := true
				fakeGatherer.IsDatabaseAvailableReturns(isAvailableReturns)
			})

			Context("When mysql metrics are enabled", func() {
				BeforeEach(func() {
					configuration.EmitMysqlMetrics = true
				})

				It("emits mysql metrics", func() {
					isAvailableReturns := true
					globalStatusReturn := map[string]string{
						"a": "b",
					}
					globalVariablesReturn := map[string]string{
						"c": "d",
					}

					availabilityMetric := &metrics.Metric{
						Key:   "Available",
						Value: 1.0,
					}

					globalStatusMetric := &metrics.Metric{
						Key: "GlobalStatus",
					}

					globalVariablesMetric := &metrics.Metric{
						Key: "GlobalVariables",
					}

					fakeGatherer.DatabaseMetadataReturns(globalStatusReturn, globalVariablesReturn, nil)
					fakeGatherer.IsDatabaseAvailableReturns(isAvailableReturns)

					fakeMetricsComputer.ComputeAvailabilityMetricReturns(availabilityMetric)
					fakeMetricsComputer.ComputeGlobalMetricsReturnsOnCall(0, []*metrics.Metric{globalStatusMetric})
					fakeMetricsComputer.ComputeGlobalMetricsReturnsOnCall(1, []*metrics.Metric{globalVariablesMetric})

					err := processor.Process()
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGatherer.DatabaseMetadataCallCount()).To(Equal(1))

					Expect(fakeMetricsComputer.ComputeAvailabilityMetricCallCount()).To(Equal(1))
					computeAvailabilityMetricArgs := fakeMetricsComputer.ComputeAvailabilityMetricArgsForCall(0)
					Expect(computeAvailabilityMetricArgs).To(Equal(isAvailableReturns))

					Expect(fakeMetricsComputer.ComputeGlobalMetricsCallCount()).To(Equal(2))
					computeGlobalStatusMetricArgs := fakeMetricsComputer.ComputeGlobalMetricsArgsForCall(0)
					Expect(computeGlobalStatusMetricArgs).To(Equal(globalStatusReturn))

					computeGlobalVariableMetricArgs := fakeMetricsComputer.ComputeGlobalMetricsArgsForCall(1)
					Expect(computeGlobalVariableMetricArgs).To(Equal(globalVariablesReturn))

					Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
					metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
					Expect(len(metricsToEmit)).To(Equal(3))
					Expect(metricsToEmit).To(ContainElement(availabilityMetric))
					Expect(metricsToEmit).To(ContainElement(globalStatusMetric))
					Expect(metricsToEmit).To(ContainElement(globalVariablesMetric))
				})
			})

			Context("When galera metrics are enabled", func() {
				BeforeEach(func() {
					configuration.EmitMysqlMetrics = true
					configuration.EmitGaleraMetrics = true
				})

				It("emits galera metrics", func() {
					isAvailableReturns := true
					globalStatusReturn := map[string]string{
						"wsrep_status": "good",
					}
					globalVariablesReturn := map[string]string{
						"wsrep_enabled": "on",
					}

					availabilityMetric := &metrics.Metric{
						Key:   "Available",
						Value: 1.0,
					}

					globalGaleraStatusMetric := &metrics.Metric{
						Key: "wsrep_status",
					}

					globalGaleraVariablesMetric := &metrics.Metric{
						Key: "wsrep_enabled",
					}

					globalStatusMetric := &metrics.Metric{
						Key: "GlobalStatus",
					}

					globalVariablesMetric := &metrics.Metric{
						Key: "GlobalVariables",
					}

					fakeGatherer.DatabaseMetadataReturns(globalStatusReturn, globalVariablesReturn, nil)
					fakeGatherer.IsDatabaseAvailableReturns(isAvailableReturns)

					fakeMetricsComputer.ComputeAvailabilityMetricReturns(availabilityMetric)
					fakeMetricsComputer.ComputeGaleraMetricsReturnsOnCall(0, []*metrics.Metric{globalGaleraStatusMetric})
					fakeMetricsComputer.ComputeGaleraMetricsReturnsOnCall(1, []*metrics.Metric{globalGaleraVariablesMetric})
					fakeMetricsComputer.ComputeGlobalMetricsReturnsOnCall(0, []*metrics.Metric{globalStatusMetric})
					fakeMetricsComputer.ComputeGlobalMetricsReturnsOnCall(1, []*metrics.Metric{globalVariablesMetric})

					err := processor.Process()
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGatherer.DatabaseMetadataCallCount()).To(Equal(1))

					Expect(fakeMetricsComputer.ComputeGlobalMetricsCallCount()).To(Equal(2))
					computeGaleraStatusMetricArgs := fakeMetricsComputer.ComputeGaleraMetricsArgsForCall(0)
					Expect(computeGaleraStatusMetricArgs).To(Equal(globalStatusReturn))

					computeGaleraVariableMetricArgs := fakeMetricsComputer.ComputeGaleraMetricsArgsForCall(1)
					Expect(computeGaleraVariableMetricArgs).To(Equal(globalVariablesReturn))

					Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
					metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
					Expect(len(metricsToEmit)).To(Equal(5))
					Expect(metricsToEmit).To(ContainElement(availabilityMetric))
					Expect(metricsToEmit).To(ContainElement(globalStatusMetric))
					Expect(metricsToEmit).To(ContainElement(globalVariablesMetric))
					Expect(metricsToEmit).To(ContainElement(globalGaleraStatusMetric))
					Expect(metricsToEmit).To(ContainElement(globalGaleraVariablesMetric))
				})
			})

			Context("When leader follower metrics are enabled", func() {
				BeforeEach(func() {
					configuration.EmitLeaderFollowerMetrics = true
				})

				It("emits leader follower metrics", func() {
					isFollowerReturns := true
					heartbeatStatusReturn := map[string]string{
						"e": "f",
					}

					slaveStatusReturn := map[string]string{
						"g": "h",
					}

					followerMetric := &metrics.Metric{
						Key:   "is_follower",
						Value: 1.0,
					}

					heartbeatStatusMetric := &metrics.Metric{
						Key: "HeartbeatStatus",
					}

					slaveStatusMetric := &metrics.Metric{
						Key: "SlaveStatus",
					}

					fakeGatherer.IsDatabaseFollowerReturns(isFollowerReturns, nil)
					fakeGatherer.FollowerMetadataReturns(slaveStatusReturn, heartbeatStatusReturn, nil)

					fakeMetricsComputer.ComputeIsFollowerMetricReturns(followerMetric)
					fakeMetricsComputer.ComputeLeaderFollowerMetricsReturnsOnCall(0, []*metrics.Metric{heartbeatStatusMetric})
					fakeMetricsComputer.ComputeLeaderFollowerMetricsReturnsOnCall(1, []*metrics.Metric{slaveStatusMetric})

					err := processor.Process()
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeMetricsComputer.ComputeLeaderFollowerMetricsCallCount()).To(Equal(2))

					computeHeartbeatStatusMetricArgs := fakeMetricsComputer.ComputeLeaderFollowerMetricsArgsForCall(0)
					Expect(computeHeartbeatStatusMetricArgs).To(Equal(heartbeatStatusReturn))

					computeSlaveStatusMetricArgs := fakeMetricsComputer.ComputeLeaderFollowerMetricsArgsForCall(1)
					Expect(computeSlaveStatusMetricArgs).To(Equal(slaveStatusReturn))

					Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
					metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
					Expect(len(metricsToEmit)).To(Equal(3))
					Expect(metricsToEmit).To(ContainElement(heartbeatStatusMetric))
					Expect(metricsToEmit).To(ContainElement(slaveStatusMetric))
					Expect(metricsToEmit).To(ContainElement(followerMetric))
				})
			})
		})

		Context("when database is not available", func() {
			BeforeEach(func() {
				isAvailableReturns := false
				fakeGatherer.IsDatabaseAvailableReturns(isAvailableReturns)
			})

			Context("When mysql metrics are enabled", func() {
				BeforeEach(func() {
					configuration.EmitMysqlMetrics = true
				})

				It("Emits an available=false metric", func() {
					isAvailable := false

					availabilityMetric := &metrics.Metric{
						Key:   "Available",
						Value: 0,
					}

					fakeGatherer.IsDatabaseAvailableReturns(isAvailable)
					fakeMetricsComputer.ComputeAvailabilityMetricReturns(availabilityMetric)

					err := processor.Process()
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGatherer.DatabaseMetadataCallCount()).To(Equal(0))

					Expect(fakeMetricsComputer.ComputeAvailabilityMetricCallCount()).To(Equal(1))
					computeAvailabilityMetricArgs := fakeMetricsComputer.ComputeAvailabilityMetricArgsForCall(0)
					Expect(computeAvailabilityMetricArgs).To(Equal(isAvailable))

					Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
					metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
					Expect(len(metricsToEmit)).To(Equal(1))
					Expect(metricsToEmit[0]).To(Equal(availabilityMetric))
				})
			})

			Context("When leader follower metrics are enabled", func() {
				BeforeEach(func() {
					configuration.EmitLeaderFollowerMetrics = true
					configuration.EmitMysqlMetrics = true
				})

				It("emits is_follower metric", func() {
					isFollowerReturns := false
					isAvailableReturns := false

					followerMetric := &metrics.Metric{
						Key:   "is_follower",
						Value: 0.0,
					}

					availabilityMetric := &metrics.Metric{
						Key:   "Available",
						Value: 0.0,
					}

					fakeGatherer.IsDatabaseFollowerReturns(isFollowerReturns, nil)
					fakeGatherer.IsDatabaseAvailableReturns(isAvailableReturns)

					fakeMetricsComputer.ComputeIsFollowerMetricReturns(followerMetric)
					fakeMetricsComputer.ComputeAvailabilityMetricReturns(availabilityMetric)

					err := processor.Process()
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeMetricsComputer.ComputeLeaderFollowerMetricsCallCount()).To(Equal(0))

					Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
					metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
					Expect(len(metricsToEmit)).To(Equal(2))
					Expect(metricsToEmit).To(ContainElement(availabilityMetric))
					Expect(metricsToEmit).To(ContainElement(followerMetric))
				})
			})
		})

		Context("error cases", func() {
			BeforeEach(func() {
				configuration.EmitGaleraMetrics = true
				configuration.EmitMysqlMetrics = true
				configuration.EmitDiskMetrics = true
				configuration.EmitLeaderFollowerMetrics = true
			})
			It("collects errors but emits whatever metrics it can", func() {
				isAvailable := true
				isFollower := true

				availabilityMetric := &metrics.Metric{
					Key:   "Available",
					Value: 1,
				}

				followerMetric := &metrics.Metric{
					Key:   "is_follower",
					Value: 1,
				}

				fakeGatherer.IsDatabaseAvailableReturns(isAvailable)
				fakeGatherer.IsDatabaseFollowerReturns(isFollower, nil)
				fakeGatherer.DatabaseMetadataReturns(nil, nil, errors.New("DatabaseMetadata failed"))
				fakeGatherer.FollowerMetadataReturns(nil, nil, errors.New("FollowerMetadata failed"))
				fakeGatherer.DiskStatsReturns(nil, errors.New("Disk Stats failed"))
				fakeMetricsComputer.ComputeAvailabilityMetricReturns(availabilityMetric)
				fakeMetricsComputer.ComputeIsFollowerMetricReturns(followerMetric)

				err := processor.Process()
				Expect(err.Error()).To(ContainSubstring("DatabaseMetadata failed"))
				Expect(err.Error()).To(ContainSubstring("FollowerMetadata failed"))
				Expect(err.Error()).To(ContainSubstring("Disk Stats failed"))
				Expect(err).To(BeAssignableToTypeOf(&multierror.Error{}))

				Expect(fakeMetricsWriter.WriteCallCount()).To(Equal(1))
				metricsToEmit := fakeMetricsWriter.WriteArgsForCall(0)
				Expect(len(metricsToEmit)).To(Equal(2))
				Expect(metricsToEmit).To(ContainElement(availabilityMetric))
				Expect(metricsToEmit).To(ContainElement(followerMetric))
			})

			It("returns an error if Write returns an error", func() {
				fakeMetricsWriter.WriteReturns(errors.New("write failed"))

				err := processor.Process()
				Expect(err.Error()).To(ContainSubstring("write failed"))
				Expect(err).To(BeAssignableToTypeOf(&multierror.Error{}))
			})

			It("returns an error if IsDatabaseFollower returns an error", func() {
				fakeGatherer.IsDatabaseAvailableReturns(true)
				fakeGatherer.IsDatabaseFollowerReturns(false, errors.New("failed to determine follower state"))

				err := processor.Process()
				Expect(err.Error()).To(ContainSubstring("failed to determine follower state"))
				Expect(err).To(BeAssignableToTypeOf(&multierror.Error{}))
			})
		})
	})
})
