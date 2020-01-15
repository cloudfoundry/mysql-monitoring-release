package gather_test

import (
	"fmt"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-metrics/gather"
	"github.com/cloudfoundry/mysql-metrics/gather/gatherfakes"
)

var _ = Describe("Gatherer", func() {
	var (
		databaseClient *gatherfakes.FakeDatabaseClient
		stater         *gatherfakes.FakeStater
		cpustater      *gatherfakes.FakeCpuStater
		gatherer       *gather.Gatherer
	)

	BeforeEach(func() {
		databaseClient = &gatherfakes.FakeDatabaseClient{}
		stater = &gatherfakes.FakeStater{}
		cpustater = &gatherfakes.FakeCpuStater{}
		gatherer = gather.NewGatherer(databaseClient, stater, cpustater)
	})

	Describe("BrokerStats", func() {
		It("returns service plans disk allocated", func() {
			diskAllocatedMap := map[string]string{
				"service_plans_disk_allocated": "200",
			}

			databaseClient.ServicePlansDiskAllocatedReturns(diskAllocatedMap, nil)

			brokerStats, err := gatherer.BrokerStats()
			Expect(err).NotTo(HaveOccurred())

			Expect(brokerStats).To(Equal(diskAllocatedMap))
		})

		Context("error cases", func() {
			It("returns an error when there is an error fetching broker stats", func() {
				databaseClient.ServicePlansDiskAllocatedReturns(nil, errors.New("db error"))

				_, err := gatherer.BrokerStats()
				Expect(err).To(MatchError("db error"))

				Expect(databaseClient.ServicePlansDiskAllocatedCallCount()).To(Equal(1))
			})
		})
	})

	Describe("CpuStats", func() {
		It("returns cpu usage", func() {
			cpustater.GetPercentageReturns(7, nil)

			cpuStats, err := gatherer.CPUStats()
			Expect(err).NotTo(HaveOccurred())

			Expect(cpuStats).To(Equal(map[string]string{"cpu_utilization_percent": "7"}))
		})

		Context("error cases", func() {
			It("returns an error when there is an error fetching cpu stats", func() {
				cpustater.GetPercentageReturns(-1, errors.New("cpu stats error"))

				cpuStats, err := gatherer.CPUStats()
				Expect(err).To(MatchError("cpu stats error"))
				Expect(cpuStats).To(BeEmpty())
			})
		})
	})

	Describe("DiskStats", func() {
		It("returns disk information for ephemeral and persistand disks", func() {
			statsMap := map[string]string{
				"persistent_disk_used":                "2024",
				"persistent_disk_free":                "1024",
				"persistent_disk_used_percent":        "66",
				"persistent_disk_inodes_used":         "50",
				"persistent_disk_inodes_free":         "450",
				"persistent_disk_inodes_used_percent": "10",
				"ephemeral_disk_used":                 "3072",
				"ephemeral_disk_free":                 "4096",
				"ephemeral_disk_used_percent":         "42",
				"ephemeral_disk_inodes_used":          "100",
				"ephemeral_disk_inodes_free":          "200",
				"ephemeral_disk_inodes_used_percent":  "33",
			}

			stater.StatsReturnsOnCall(0, 1048576, 3121152, 450, 500, nil)
			stater.StatsReturnsOnCall(1, 4194304, 7340032, 200, 300, nil)

			stats, err := gatherer.DiskStats()
			Expect(err).NotTo(HaveOccurred())

			Expect(stats).To(Equal(statsMap))

			Expect(stater.StatsCallCount()).To(Equal(2))
			Expect(stater.StatsArgsForCall(0)).To(Equal("/var/vcap/store"))
			Expect(stater.StatsArgsForCall(1)).To(Equal("/var/vcap/data"))
		})

		Context("error cases", func() {
			It("returns an error when the persistent disk fails to be described", func() {
				stater.StatsReturnsOnCall(0, 0, 0, 0, 0, errors.New("failed to inspect persistent disk"))

				_, err := gatherer.DiskStats()
				Expect(err).To(MatchError("failed to inspect persistent disk"))

				Expect(stater.StatsCallCount()).To(Equal(1))
				Expect(stater.StatsArgsForCall(0)).To(Equal("/var/vcap/store"))
			})

			It("returns an error when the ephemeral disk fails to be described", func() {
				stater.StatsReturnsOnCall(1, 0, 0, 0, 0, errors.New("failed to inspect ephemeral disk"))

				_, err := gatherer.DiskStats()
				Expect(err).To(MatchError("failed to inspect ephemeral disk"))

				Expect(stater.StatsCallCount()).To(Equal(2))
				Expect(stater.StatsArgsForCall(1)).To(Equal("/var/vcap/data"))
			})
		})
	})

	Describe("IsDatabaseFollower", func() {
		It("returns true if the database is a follower node", func() {
			databaseClient.IsFollowerReturns(true, nil)

			isFollower, err := gatherer.IsDatabaseFollower()
			Expect(isFollower).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())

			Expect(databaseClient.IsFollowerCallCount()).To(Equal(1))
		})

		Context("the database connection is dead", func() {
			It("returns an error", func() {
				databaseClient.IsFollowerReturns(false, errors.New("could not determine follower state"))

				_, err := gatherer.IsDatabaseFollower()
				Expect(err).To(MatchError("could not determine follower state"))

				Expect(databaseClient.IsFollowerCallCount()).To(Equal(1))
			})
		})

	})

	Describe("IsDatabaseAvailable", func() {
		It("returns true if the database is available", func() {
			databaseClient.IsAvailableReturns(true)

			isAvailable := gatherer.IsDatabaseAvailable()
			Expect(isAvailable).To(BeTrue())

			Expect(databaseClient.IsAvailableCallCount()).To(Equal(1))
		})
	})

	Describe("FollowerMetadata", func() {
		It("returns replication metadata from the database", func() {
			slaveStatusMap := map[string]string{
				"doesnt-matter": "345",
			}

			heartbeatStatusMap := map[string]string{
				"seconds_since_leader_heartbeat": "4",
			}
			databaseClient.ShowSlaveStatusReturns(slaveStatusMap, nil)
			databaseClient.HeartbeatStatusReturns(heartbeatStatusMap, nil)

			slaveStatus, heartbeatStatus, err := gatherer.FollowerMetadata()
			Expect(err).NotTo(HaveOccurred())

			Expect(slaveStatus).To(Equal(slaveStatusMap))
			Expect(heartbeatStatus).To(Equal(heartbeatStatusMap))
		})

		Context("error cases", func() {
			It("returns an errors when ShowSlaveStatus fails", func() {
				databaseClient.ShowSlaveStatusReturns(nil, errors.New("ShowSlaveStatus failed"))

				_, _, err := gatherer.FollowerMetadata()
				Expect(err).To(MatchError("ShowSlaveStatus failed"))
			})

			It("returns ShowSlaveStatus even when HeartbeatStatus fails", func() {
				slaveStatusMap := map[string]string{
					"doesnt-matter": "345",
				}

				databaseClient.ShowSlaveStatusReturns(slaveStatusMap, nil)
				databaseClient.HeartbeatStatusReturns(nil, errors.New("HeartbeatStatus failed"))

				slaveStatus, _, err := gatherer.FollowerMetadata()
				Expect(err).To(MatchError("HeartbeatStatus failed"))
				Expect(slaveStatus).To(Equal(slaveStatusMap))
			})
		})
	})

	Describe("DatabaseMetadata", func() {
		It("returns metadata from the database", func() {
			globalStatusMap := map[string]string{
				"questions":                     "Nope",
				"innodb_buffer_pool_pages_free": "0",
			}
			globalVariablesMap := map[string]string{
				"max_connections": fmt.Sprintf("%v", 1),
			}

			databaseClient.ShowGlobalStatusReturns(globalStatusMap, nil)
			databaseClient.ShowGlobalVariablesReturns(globalVariablesMap, nil)

			globalStatus, globalVariables, err := gatherer.DatabaseMetadata()
			Expect(err).NotTo(HaveOccurred())

			Expect(globalStatus).To(Equal(globalStatusMap))
			Expect(globalVariables).To(Equal(globalVariablesMap))
		})

		Context("error cases", func() {
			It("returns an errors when ShowGlobalStatus fails", func() {
				databaseClient.ShowGlobalStatusReturns(nil, errors.New("ShowGlobalStatus failed"))

				_, _, err := gatherer.DatabaseMetadata()
				Expect(err).To(MatchError("ShowGlobalStatus failed"))
			})

			It("returns an errors when ShowGlobalVariables fails", func() {
				databaseClient.ShowGlobalVariablesReturns(nil, errors.New("ShowGlobalVariables failed"))

				_, _, err := gatherer.DatabaseMetadata()
				Expect(err).To(MatchError("ShowGlobalVariables failed"))
			})
		})

		Context("queries_delta", func() {
			It("returns the queries_delta metric with the difference between queries", func() {
				databaseClient.ShowGlobalStatusReturnsOnCall(0, map[string]string{"queries": "0"}, nil)
				databaseClient.ShowGlobalStatusReturnsOnCall(1, map[string]string{"queries": "5"}, nil)
				databaseClient.ShowGlobalStatusReturnsOnCall(2, map[string]string{"queries": "16"}, nil)

				globalStatus, _, err := gatherer.DatabaseMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(globalStatus).ToNot(HaveKey("queries_delta"))

				globalStatus, _, err = gatherer.DatabaseMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(globalStatus).To(HaveKeyWithValue("queries_delta", "5"))

				globalStatus, _, err = gatherer.DatabaseMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(globalStatus).To(HaveKeyWithValue("queries_delta", "11"))
			})

			Context("when queries metric is not parseable", func() {
				It("returns queries_delta of 0", func() {
					databaseClient.ShowGlobalStatusReturnsOnCall(0, map[string]string{"queries": "%%%%%"}, nil)

					globalStatus, _, err := gatherer.DatabaseMetadata()
					Expect(err).NotTo(HaveOccurred())
					Expect(globalStatus).To(HaveKeyWithValue("queries_delta", "0"))
				})
			})

			Context("when queries metric does not exist", func() {
				It("returns queries_delta of 0", func() {
					databaseClient.ShowGlobalStatusReturnsOnCall(0, map[string]string{}, nil)

					globalStatus, _, err := gatherer.DatabaseMetadata()
					Expect(err).NotTo(HaveOccurred())
					Expect(globalStatus).To(HaveKeyWithValue("queries_delta", "0"))
				})
			})

			Context("when queries metric goes backward due to a mysql restart", func() {
				It("returns queries_delta of 0", func() {
					databaseClient.ShowGlobalStatusReturnsOnCall(0, map[string]string{"queries": "0"}, nil)
					databaseClient.ShowGlobalStatusReturnsOnCall(1, map[string]string{"queries": "1000"}, nil)
					databaseClient.ShowGlobalStatusReturnsOnCall(2, map[string]string{"queries": "0"}, nil)

					globalStatus, _, err := gatherer.DatabaseMetadata()
					Expect(err).NotTo(HaveOccurred())
					Expect(globalStatus).ToNot(HaveKey("queries_delta"))

					globalStatus, _, err = gatherer.DatabaseMetadata()
					Expect(err).NotTo(HaveOccurred())
					Expect(globalStatus).To(HaveKeyWithValue("queries_delta", "1000"))

					globalStatus, _, err = gatherer.DatabaseMetadata()
					Expect(err).NotTo(HaveOccurred())
					Expect(globalStatus).ToNot(HaveKey("queries_delta"))
				})
			})
		})

	})
})
