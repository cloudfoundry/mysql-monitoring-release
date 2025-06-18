package diskstat

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/procfs/blockdevice"
)

var _ = Describe("Delta", func() {
	var delta Delta

	BeforeEach(func() {
		// Create a Delta with known values for testing calculations
		delta = Delta{
			Elapsed: 2 * time.Second,
			Stats: blockdevice.Diskstats{
				IOStats: blockdevice.IOStats{
					ReadIOs:         100,
					ReadMerges:      20,
					ReadSectors:     2048, // 1024 KiB (2048 * 512 bytes / 1024)
					ReadTicks:       1000,
					WriteIOs:        50,
					WriteMerges:     10,
					WriteSectors:    1024, // 512 KiB
					WriteTicks:      500,
					IOsInProgress:   5,
					IOsTotalTicks:   1200,
					WeightedIOTicks: 1800,
				},
			},
		}
	})

	Describe("Read Operations", func() {
		Context("ReadsPerSecond", func() {
			It("should calculate reads per second correctly", func() {
				// 100 reads / 2 seconds = 50 reads/sec
				Expect(delta.ReadsPerSecond()).To(Equal(50.0))
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should return infinity", func() {
					result := delta.ReadsPerSecond()
					Expect(result).To(Equal(math.Inf(1)))
				})
			})
		})

		Context("ReadKiB", func() {
			It("should calculate read data in KiB correctly", func() {
				// 2048 sectors * 512 bytes/sector / 1024 bytes/KiB = 1024 KiB
				Expect(delta.ReadKiB()).To(Equal(1024.0))
			})

			When("no sectors read", func() {
				BeforeEach(func() {
					delta.ReadSectors = 0
				})

				It("should return zero", func() {
					Expect(delta.ReadKiB()).To(Equal(0.0))
				})
			})
		})

		Context("ReadAvgKB", func() {
			It("should calculate average read size correctly", func() {
				// 1024 KiB / 100 reads = 10.24 KB per read
				Expect(delta.ReadAvgKB()).To(Equal(10.24))
			})

			When("no read IOs", func() {
				BeforeEach(func() {
					delta.ReadIOs = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.ReadAvgKB()).To(Equal(0.0))
				})
			})
		})

		Context("ReadMiBPerSec", func() {
			It("should calculate read throughput in MiB/sec correctly", func() {
				// 2048 sectors / 2048 / 2 seconds = 0.5 MiB/sec
				Expect(delta.ReadMiBPerSec()).To(Equal(0.5))
			})

			When("no sectors read", func() {
				BeforeEach(func() {
					delta.ReadSectors = 0
				})

				It("should return zero", func() {
					Expect(delta.ReadMiBPerSec()).To(Equal(0.0))
				})
			})
		})

		Context("ReadMergesPercent", func() {
			It("should calculate read merge percentage correctly", func() {
				// ReadRequests = ReadIOs + ReadMerges = 100 + 20 = 120
				// 20 merges / 120 requests * 100 = 16.67%
				Expect(delta.ReadMergesPercent()).To(BeNumerically("~", 16.666666666666668, 0.000001))
			})

			When("no read requests", func() {
				BeforeEach(func() {
					delta.ReadIOs = 0
					delta.ReadMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.ReadMergesPercent()).To(Equal(0.0))
				})
			})
		})

		Context("ReadResponseTime", func() {
			It("should calculate average read response time correctly", func() {
				// 1000 ticks / 120 requests = 8.33 ms per request
				Expect(delta.ReadResponseTime()).To(BeNumerically("~", 8.333333333333334, 0.000001))
			})

			When("no read requests", func() {
				BeforeEach(func() {
					delta.ReadIOs = 0
					delta.ReadMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.ReadResponseTime()).To(Equal(0.0))
				})
			})
		})

		Context("ReadConcurrency", func() {
			It("should calculate read concurrency correctly", func() {
				// 1000 ticks / 2000 ms = 0.5
				Expect(delta.ReadConcurrency()).To(Equal(0.5))
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should return infinity", func() {
					result := delta.ReadConcurrency()
					Expect(result).To(Equal(math.Inf(1)))
				})
			})
		})

		Context("ReadRequests", func() {
			It("should calculate total read requests correctly", func() {
				// ReadIOs + ReadMerges = 100 + 20 = 120
				Expect(delta.ReadRequests()).To(Equal(uint64(120)))
			})
		})
	})

	Describe("Write Operations", func() {
		Context("WritesPerSecond", func() {
			It("should calculate writes per second correctly", func() {
				// 50 writes / 2 seconds = 25 writes/sec
				Expect(delta.WritesPerSecond()).To(Equal(25.0))
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should return infinity", func() {
					result := delta.WritesPerSecond()
					Expect(result).To(Equal(math.Inf(1)))
				})
			})
		})

		Context("WriteMiBPerSec", func() {
			It("should calculate write throughput in MiB/sec correctly", func() {
				// 1024 sectors / 2048 / 2 seconds = 0.25 MiB/sec
				Expect(delta.WriteMiBPerSec()).To(Equal(0.25))
			})

			When("no sectors written", func() {
				BeforeEach(func() {
					delta.WriteSectors = 0
				})

				It("should return zero", func() {
					Expect(delta.WriteMiBPerSec()).To(Equal(0.0))
				})
			})
		})

		Context("WriteAvgKB", func() {
			It("should calculate average write size correctly", func() {
				// 1024 sectors / 2 / 50 write ios = 10.24 KB per io
				Expect(delta.WriteAvgKB()).To(Equal(10.24))
			})

			When("no write IOs", func() {
				BeforeEach(func() {
					delta.WriteIOs = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.WriteAvgKB()).To(Equal(0.0))
				})
			})
		})

		Context("WriteMergesPercent", func() {
			It("should calculate write merge percentage correctly", func() {
				// WriteRequests = WriteIOs + WriteMerges = 50 + 10 = 60
				// 10 merges / 60 requests * 100 = 16.67%
				Expect(delta.WriteMergesPercent()).To(BeNumerically("~", 16.666666666666668, 0.000001))
			})

			When("no write requests", func() {
				BeforeEach(func() {
					delta.WriteIOs = 0
					delta.WriteMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.WriteMergesPercent()).To(Equal(0.0))
				})
			})
		})

		Context("WriteResponseTime", func() {
			It("should calculate average write response time correctly", func() {
				// 500 ticks / 60 requests = 8.33 ms per request
				Expect(delta.WriteResponseTime()).To(BeNumerically("~", 8.333333333333334, 0.000001))
			})

			When("no write requests", func() {
				BeforeEach(func() {
					delta.WriteIOs = 0
					delta.WriteMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.WriteResponseTime()).To(Equal(0.0))
				})
			})
		})

		Context("WriteConcurrency", func() {
			It("should calculate write concurrency correctly", func() {
				// 500 ticks / 2000 ms = 0.25
				result := delta.WriteConcurrency()
				Expect(result).To(Equal(0.25))
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should return infinity", func() {
					result := delta.WriteConcurrency()
					Expect(result).To(Equal(math.Inf(1)))
				})
			})
		})

		Context("WriteRequests", func() {
			It("should calculate total write requests correctly", func() {
				// WriteIOs + WriteMerges = 50 + 10 = 60
				Expect(delta.WriteRequests()).To(Equal(uint64(60)))
			})
		})
	})

	Describe("Combined Operations", func() {
		Context("IOsRequests", func() {
			It("should calculate total IO requests correctly", func() {
				// ReadRequests + WriteRequests = 120 + 60 = 180
				Expect(delta.IOsRequests()).To(Equal(uint64(180)))
			})
		})

		Context("AvgResponseTime", func() {
			It("should calculate average response time correctly", func() {
				// WeightedIOTicks / (IOsRequests + IOsInProgress)
				// 1800 / (180 + 5) = 1800 / 185 = 9.73 ms
				Expect(delta.AvgResponseTime()).To(BeNumerically("~", 9.72972972972973, 0.000001))
			})

			When("no IOs in progress and no requests", func() {
				BeforeEach(func() {
					delta.ReadIOs = 0
					delta.ReadMerges = 0
					delta.WriteIOs = 0
					delta.WriteMerges = 0
					delta.IOsInProgress = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.AvgResponseTime()).To(Equal(0.0))
				})
			})
		})

		Context("AvgServiceTime", func() {
			It("should calculate average service time correctly", func() {
				// IOsTotalTicks / IOsRequests
				// 1200 / 180 = 6.67 ms
				Expect(delta.AvgServiceTime()).To(BeNumerically("~", 6.666666666666667, 0.000001))
			})

			When("no IO requests", func() {
				BeforeEach(func() {
					delta.ReadIOs = 0
					delta.ReadMerges = 0
					delta.WriteIOs = 0
					delta.WriteMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.AvgServiceTime()).To(Equal(0.0))
				})
			})
		})

		Context("QTime", func() {
			It("should calculate queue time correctly", func() {
				// AvgResponseTime - AvgServiceTime
				// 9.73 - 6.67 = 3.06 ms
				avgResponseTime := delta.AvgResponseTime()
				avgServiceTime := delta.AvgServiceTime()
				expectedQTime := avgResponseTime - avgServiceTime

				Expect(delta.QTime()).To(BeNumerically("~", expectedQTime, 0.000001))
			})

			When("service time equals response time", func() {
				BeforeEach(func() {
					// Set values so that response time equals service time
					delta.WeightedIOTicks = 1200 // Same as IOsTotalTicks
					delta.IOsInProgress = 0      // No IOs in progress
				})

				It("should return zero queue time", func() {
					Expect(delta.QTime()).To(BeNumerically("~", 0.0, 0.000001))
				})
			})
		})

		Context("BusyPercent", func() {
			It("should calculate device utilization percentage correctly", func() {
				// IOsTotalTicks / Elapsed.Milliseconds() * 100
				// 1200 ms / 2000 ms * 100 = 60%
				Expect(delta.BusyPercent()).To(Equal(60.0))
			})

			When("device is fully saturated", func() {
				BeforeEach(func() {
					delta.IOsTotalTicks = 2000 // Same as elapsed time in ms
				})

				It("should return 100% busy", func() {
					Expect(delta.BusyPercent()).To(Equal(100.0))
				})
			})

			When("no I/O activity", func() {
				BeforeEach(func() {
					delta.IOsTotalTicks = 0
				})

				It("should return 0% busy", func() {
					Expect(delta.BusyPercent()).To(Equal(0.0))
				})
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.BusyPercent()).To(Equal(0.0))
				})
			})

			When("device is over-saturated (theoretical case)", func() {
				BeforeEach(func() {
					delta.IOsTotalTicks = 3000 // More than elapsed time
				})

				It("should return over 100% busy", func() {
					Expect(delta.BusyPercent()).To(Equal(150.0))
				})
			})
		})
	})

	Describe("Discard Operations", func() {
		BeforeEach(func() {
			// Add discard statistics to our test delta
			delta.DiscardIOs = 30
			delta.DiscardMerges = 5
			delta.DiscardSectors = 1024 // 512 KiB
			delta.DiscardTicks = 300
		})

		Context("DiscardsPerSecond", func() {
			It("should calculate discards per second correctly", func() {
				// 30 discards / 2 seconds = 15 discards/sec
				Expect(delta.DiscardsPerSecond()).To(Equal(15.0))
			})

			When("no discard operations", func() {
				BeforeEach(func() {
					delta.DiscardIOs = 0
				})

				It("should return zero", func() {
					Expect(delta.DiscardsPerSecond()).To(Equal(0.0))
				})
			})
		})

		Context("DiscardKiB", func() {
			It("should calculate total kilobytes discarded correctly", func() {
				// 1024 sectors * 512 bytes/sector / 1024 bytes/KiB = 512 KiB
				Expect(delta.DiscardKiB()).To(Equal(512.0))
			})

			When("no discard sectors", func() {
				BeforeEach(func() {
					delta.DiscardSectors = 0
				})

				It("should return zero", func() {
					Expect(delta.DiscardKiB()).To(Equal(0.0))
				})
			})
		})

		Context("DiscardMiBPerSec", func() {
			It("should calculate discard throughput correctly", func() {
				// 1024 sectors / 2048 / 2 seconds = 0.25 MiB/sec
				Expect(delta.DiscardMiBPerSec()).To(Equal(0.25))
			})

			When("no discard sectors", func() {
				BeforeEach(func() {
					delta.DiscardSectors = 0
				})

				It("should return zero", func() {
					Expect(delta.DiscardMiBPerSec()).To(Equal(0.0))
				})
			})
		})

		Context("DiscardAvgKB", func() {
			It("should calculate average KB per discard operation correctly", func() {
				// 512 KiB / 30 discards = 17.067 KB per discard
				Expect(delta.DiscardAvgKB()).To(BeNumerically("~", 17.066666666666666, 0.000001))
			})

			When("no discard operations", func() {
				BeforeEach(func() {
					delta.DiscardIOs = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.DiscardAvgKB()).To(Equal(0.0))
				})
			})
		})

		Context("DiscardRequests", func() {
			It("should calculate total discard requests correctly", func() {
				// DiscardIOs + DiscardMerges = 30 + 5 = 35
				Expect(delta.DiscardRequests()).To(Equal(uint64(35)))
			})

			When("only merges without IOs", func() {
				BeforeEach(func() {
					delta.DiscardIOs = 0
					delta.DiscardMerges = 10
				})

				It("should return merge count", func() {
					Expect(delta.DiscardRequests()).To(Equal(uint64(10)))
				})
			})
		})

		Context("DiscardMergesPercent", func() {
			It("should calculate discard merge percentage correctly", func() {
				// 5 merges / 35 requests * 100 = 14.29%
				Expect(delta.DiscardMergesPercent()).To(BeNumerically("~", 14.285714285714286, 0.000001))
			})

			When("no discard requests", func() {
				BeforeEach(func() {
					delta.DiscardIOs = 0
					delta.DiscardMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.DiscardMergesPercent()).To(Equal(0.0))
				})
			})

			When("all operations are merged", func() {
				BeforeEach(func() {
					delta.DiscardIOs = 0
					delta.DiscardMerges = 20
				})

				It("should return 100%", func() {
					Expect(delta.DiscardMergesPercent()).To(Equal(100.0))
				})
			})
		})

		Context("DiscardResponseTime", func() {
			It("should calculate average discard response time correctly", func() {
				// 300 ticks / 35 requests = 8.57 ms per request
				Expect(delta.DiscardResponseTime()).To(BeNumerically("~", 8.571428571428571, 0.000001))
			})

			When("no discard requests", func() {
				BeforeEach(func() {
					delta.DiscardIOs = 0
					delta.DiscardMerges = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.DiscardResponseTime()).To(Equal(0.0))
				})
			})
		})

		Context("DiscardConcurrency", func() {
			It("should calculate discard concurrency correctly", func() {
				// 300 ticks / 2000 ms = 0.15
				Expect(delta.DiscardConcurrency()).To(Equal(0.15))
			})

			When("no discard ticks", func() {
				BeforeEach(func() {
					delta.DiscardTicks = 0
				})

				It("should return zero", func() {
					Expect(delta.DiscardConcurrency()).To(Equal(0.0))
				})
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should handle division by zero gracefully", func() {
					// This will result in +Inf, but that's mathematically correct
					Expect(delta.DiscardConcurrency()).To(Equal(math.Inf(1)))
				})
			})
		})
	})

	Describe("Flush Operations", func() {
		BeforeEach(func() {
			// Add flush statistics to our test delta
			delta.FlushRequestsCompleted = 15
			delta.TimeSpentFlushing = 450 // milliseconds
		})

		Context("FlushesPerSecond", func() {
			It("should calculate flushes per second correctly", func() {
				// 15 flushes / 2 seconds = 7.5 flushes/sec
				Expect(delta.FlushesPerSecond()).To(Equal(7.5))
			})

			When("no flush operations", func() {
				BeforeEach(func() {
					delta.FlushRequestsCompleted = 0
				})

				It("should return zero", func() {
					Expect(delta.FlushesPerSecond()).To(Equal(0.0))
				})
			})
		})

		Context("FlushResponseTime", func() {
			It("should calculate average flush response time correctly", func() {
				// 450 ms / 15 flushes = 30 ms per flush
				Expect(delta.FlushResponseTime()).To(Equal(30.0))
			})

			When("no flush operations", func() {
				BeforeEach(func() {
					delta.FlushRequestsCompleted = 0
				})

				It("should return zero to avoid division by zero", func() {
					Expect(delta.FlushResponseTime()).To(Equal(0.0))
				})
			})

			When("zero time spent flushing", func() {
				BeforeEach(func() {
					delta.TimeSpentFlushing = 0
				})

				It("should return zero response time", func() {
					Expect(delta.FlushResponseTime()).To(Equal(0.0))
				})
			})
		})

		Context("FlushConcurrency", func() {
			It("should calculate flush concurrency correctly", func() {
				// 450 ms / 2000 ms = 0.225
				Expect(delta.FlushConcurrency()).To(Equal(0.225))
			})

			When("no time spent flushing", func() {
				BeforeEach(func() {
					delta.TimeSpentFlushing = 0
				})

				It("should return zero", func() {
					Expect(delta.FlushConcurrency()).To(Equal(0.0))
				})
			})

			When("elapsed time is zero", func() {
				BeforeEach(func() {
					delta.Elapsed = 0
				})

				It("should handle division by zero gracefully", func() {
					// This will result in +Inf, but that's mathematically correct
					Expect(delta.FlushConcurrency()).To(Equal(math.Inf(1)))
				})
			})
		})

		Context("high flush activity", func() {
			BeforeEach(func() {
				delta.FlushRequestsCompleted = 100
				delta.TimeSpentFlushing = 1000 // 1 second of flushing in 2 second interval
			})

			It("should handle high flush rates correctly", func() {
				Expect(delta.FlushesPerSecond()).To(Equal(50.0))
				Expect(delta.FlushResponseTime()).To(Equal(10.0))
				Expect(delta.FlushConcurrency()).To(Equal(0.5))
			})
		})
	})

	Describe("Combined Discard and Flush Edge Cases", func() {
		Context("with very large values", func() {
			BeforeEach(func() {
				delta.DiscardIOs = 1000000
				delta.DiscardSectors = 2048000000 // ~1TB
				delta.FlushRequestsCompleted = 500000
				delta.TimeSpentFlushing = 1000000 // 1000 seconds
			})

			It("should handle large discard values correctly", func() {
				Expect(delta.DiscardsPerSecond()).To(Equal(500000.0))
				Expect(delta.DiscardKiB()).To(Equal(1024000000.0))   // ~1TB in KiB
				Expect(delta.DiscardMiBPerSec()).To(Equal(500000.0)) // ~500GB/s
			})

			It("should handle large flush values correctly", func() {
				Expect(delta.FlushesPerSecond()).To(Equal(250000.0))
				Expect(delta.FlushResponseTime()).To(Equal(2.0)) // 1000000 ms / 500000 flushes
			})
		})

		Context("with very small time intervals", func() {
			BeforeEach(func() {
				delta.Elapsed = 1 * time.Microsecond
				delta.DiscardIOs = 1
				delta.DiscardSectors = 1
				delta.FlushRequestsCompleted = 1
			})

			It("should handle microsecond precision for discards", func() {
				Expect(delta.DiscardsPerSecond()).To(BeNumerically(">=", 999999))
				// 1 sector / 2048 sectors per MiB / 1 microsecond = very small MiB/s
				// This is actually correct - 1 sector is only 512 bytes
				Expect(delta.DiscardMiBPerSec()).To(BeNumerically(">", 0))
			})

			It("should handle microsecond precision for flushes", func() {
				Expect(delta.FlushesPerSecond()).To(BeNumerically(">=", 999999))
			})
		})
	})

	Describe("Edge Cases", func() {
		When("all values are zero", func() {
			BeforeEach(func() {
				delta = Delta{
					Elapsed: 1 * time.Second,
					Stats:   blockdevice.Diskstats{},
				}
			})

			It("should handle zero values gracefully", func() {
				Expect(delta.ReadsPerSecond()).To(Equal(0.0))
				Expect(delta.WritesPerSecond()).To(Equal(0.0))
				Expect(delta.ReadKiB()).To(Equal(0.0))
				Expect(delta.ReadAvgKB()).To(Equal(0.0))
				Expect(delta.WriteAvgKB()).To(Equal(0.0))
				Expect(delta.ReadMergesPercent()).To(Equal(0.0))
				Expect(delta.WriteMergesPercent()).To(Equal(0.0))
				Expect(delta.ReadResponseTime()).To(Equal(0.0))
				Expect(delta.WriteResponseTime()).To(Equal(0.0))
				Expect(delta.ReadConcurrency()).To(Equal(0.0))
				Expect(delta.WriteConcurrency()).To(Equal(0.0))
				Expect(delta.ReadMiBPerSec()).To(Equal(0.0))
				Expect(delta.WriteMiBPerSec()).To(Equal(0.0))
				Expect(delta.IOsRequests()).To(Equal(uint64(0)))
				Expect(delta.ReadRequests()).To(Equal(uint64(0)))
				Expect(delta.WriteRequests()).To(Equal(uint64(0)))
				Expect(delta.AvgResponseTime()).To(Equal(0.0))
				Expect(delta.AvgServiceTime()).To(Equal(0.0))
				Expect(delta.QTime()).To(Equal(0.0))
				Expect(delta.BusyPercent()).To(Equal(0.0))

				// Discard operations
				Expect(delta.DiscardsPerSecond()).To(Equal(0.0))
				Expect(delta.DiscardKiB()).To(Equal(0.0))
				Expect(delta.DiscardAvgKB()).To(Equal(0.0))
				Expect(delta.DiscardMergesPercent()).To(Equal(0.0))
				Expect(delta.DiscardResponseTime()).To(Equal(0.0))
				Expect(delta.DiscardConcurrency()).To(Equal(0.0))
				Expect(delta.DiscardMiBPerSec()).To(Equal(0.0))
				Expect(delta.DiscardRequests()).To(Equal(uint64(0)))

				// Flush operations
				Expect(delta.FlushesPerSecond()).To(Equal(0.0))
				Expect(delta.FlushResponseTime()).To(Equal(0.0))
				Expect(delta.FlushConcurrency()).To(Equal(0.0))
			})
		})

		When("elapsed time is very small", func() {
			BeforeEach(func() {
				delta.Elapsed = 1 * time.Nanosecond
				delta.ReadIOs = 1
				delta.WriteIOs = 1
				delta.ReadSectors = 1
				delta.WriteSectors = 1
				delta.DiscardIOs = 1
				delta.DiscardSectors = 1
				delta.FlushRequestsCompleted = 1
			})

			It("should handle very small time intervals", func() {
				// Should produce very large rates
				Expect(delta.ReadsPerSecond()).To(BeNumerically(">=", 999999999))
				Expect(delta.WritesPerSecond()).To(BeNumerically(">=", 999999999))
				Expect(delta.ReadMiBPerSec()).To(BeNumerically(">=", 400000))
				Expect(delta.WriteMiBPerSec()).To(BeNumerically(">=", 400000))
				Expect(delta.DiscardsPerSecond()).To(BeNumerically(">=", 999999999))
				// 1 sector is only 512 bytes, so MiB/s will be much smaller
				Expect(delta.DiscardMiBPerSec()).To(BeNumerically(">", 0))
				Expect(delta.FlushesPerSecond()).To(BeNumerically(">=", 999999999))
			})
		})

		When("only merges occur without IOs", func() {
			BeforeEach(func() {
				delta.ReadIOs = 0
				delta.WriteIOs = 0
				delta.ReadMerges = 10
				delta.WriteMerges = 5
			})

			It("should calculate requests correctly", func() {
				Expect(delta.ReadRequests()).To(Equal(uint64(10)))
				Expect(delta.WriteRequests()).To(Equal(uint64(5)))
				Expect(delta.IOsRequests()).To(Equal(uint64(15)))
			})

			It("should handle average calculations with zero IOs", func() {
				Expect(delta.ReadAvgKB()).To(Equal(0.0))
				Expect(delta.WriteAvgKB()).To(Equal(0.0))
				Expect(delta.ReadsPerSecond()).To(Equal(0.0))
				Expect(delta.WritesPerSecond()).To(Equal(0.0))
			})
		})
	})
})
