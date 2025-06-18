package diskstat

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/procfs/blockdevice"
)

var _ = Describe("SampleDelta", func() {
	var (
		cur  Sample
		prev Sample
	)

	BeforeEach(func() {
		// Set up base time
		baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

		// Create previous sample
		prev = Sample{
			Timestamp: baseTime,
			Stats: blockdevice.Diskstats{
				Info: blockdevice.Info{DeviceName: "sda1"},
				IOStats: blockdevice.IOStats{
					ReadIOs:                1000,
					ReadMerges:             100,
					ReadSectors:            8000,
					ReadTicks:              2000,
					WriteIOs:               500,
					WriteMerges:            50,
					WriteSectors:           4000,
					WriteTicks:             1000,
					IOsInProgress:          5,
					IOsTotalTicks:          3000,
					WeightedIOTicks:        3500,
					DiscardIOs:             10,
					DiscardMerges:          2,
					DiscardSectors:         100,
					DiscardTicks:           50,
					FlushRequestsCompleted: 20,
					TimeSpentFlushing:      100,
				},
			},
		}

		// Create current sample (2 seconds later with increased values)
		cur = Sample{
			Timestamp: baseTime.Add(2 * time.Second),
			Stats: blockdevice.Diskstats{
				Info: blockdevice.Info{DeviceName: "sda1"},
				IOStats: blockdevice.IOStats{
					ReadIOs:                1200,
					ReadMerges:             120,
					ReadSectors:            9600,
					ReadTicks:              2400,
					WriteIOs:               600,
					WriteMerges:            60,
					WriteSectors:           4800,
					WriteTicks:             1200,
					IOsInProgress:          8,
					IOsTotalTicks:          3600,
					WeightedIOTicks:        4200,
					DiscardIOs:             15,
					DiscardMerges:          4,
					DiscardSectors:         150,
					DiscardTicks:           75,
					FlushRequestsCompleted: 25,
					TimeSpentFlushing:      125,
				},
			},
		}
	})

	Context("with valid sample data", func() {
		It("should calculate elapsed time correctly", func() {
			delta := SampleDelta(cur, prev)
			Expect(delta.Elapsed).To(Equal(2 * time.Second))
		})

		It("should calculate read IO deltas correctly", func() {
			delta := SampleDelta(cur, prev)

			Expect(delta.ReadIOs).To(Equal(uint64(200)))      // 1200 - 1000
			Expect(delta.ReadMerges).To(Equal(uint64(20)))    // 120 - 100
			Expect(delta.ReadSectors).To(Equal(uint64(1600))) // 9600 - 8000
			Expect(delta.ReadTicks).To(Equal(uint64(400)))    // 2400 - 2000
		})

		It("should calculate write IO deltas correctly", func() {
			delta := SampleDelta(cur, prev)

			Expect(delta.WriteIOs).To(Equal(uint64(100)))     // 600 - 500
			Expect(delta.WriteMerges).To(Equal(uint64(10)))   // 60 - 50
			Expect(delta.WriteSectors).To(Equal(uint64(800))) // 4800 - 4000
			Expect(delta.WriteTicks).To(Equal(uint64(200)))   // 1200 - 1000
		})

		It("should calculate other IO deltas correctly", func() {
			delta := SampleDelta(cur, prev)

			Expect(delta.IOsTotalTicks).To(Equal(uint64(600)))        // 3600 - 3000
			Expect(delta.WeightedIOTicks).To(Equal(uint64(700)))      // 4200 - 3500
			Expect(delta.DiscardIOs).To(Equal(uint64(5)))             // 15 - 10
			Expect(delta.DiscardMerges).To(Equal(uint64(2)))          // 4 - 2
			Expect(delta.DiscardSectors).To(Equal(uint64(50)))        // 150 - 100
			Expect(delta.DiscardTicks).To(Equal(uint64(25)))          // 75 - 50
			Expect(delta.FlushRequestsCompleted).To(Equal(uint64(5))) // 25 - 20
			Expect(delta.TimeSpentFlushing).To(Equal(uint64(25)))     // 125 - 100
		})

		It("should handle IOsInProgress as a gauge (current value)", func() {
			delta := SampleDelta(cur, prev)
			// IOsInProgress is a gauge, so it should be the current value
			Expect(delta.IOsInProgress).To(Equal(uint64(8))) // Current value: 8
		})

		It("should handle IOsInProgress regardless of previous value", func() {
			// Set current IOsInProgress to be less than previous
			cur.IOsInProgress = 3 // Less than prev.IOsInProgress (5)

			delta := SampleDelta(cur, prev)
			// Should be the current value (3), not a delta
			Expect(delta.IOsInProgress).To(Equal(uint64(3)))
		})

		It("should return a Delta with correct structure", func() {
			delta := SampleDelta(cur, prev)

			// Verify it's a proper Delta struct
			Expect(delta.Elapsed).To(BeNumerically(">", 0))

			// Verify we can call Delta methods
			Expect(delta.ReadsPerSecond()).To(Equal(100.0)) // 200 reads / 2 seconds
			Expect(delta.WritesPerSecond()).To(Equal(50.0)) // 100 writes / 2 seconds
		})
	})

	Context("with zero elapsed time", func() {
		BeforeEach(func() {
			// Make both samples have the same timestamp
			cur.Timestamp = prev.Timestamp
		})

		It("should calculate zero elapsed time", func() {
			delta := SampleDelta(cur, prev)
			Expect(delta.Elapsed).To(Equal(time.Duration(0)))
		})

		It("should still calculate deltas correctly", func() {
			delta := SampleDelta(cur, prev)

			Expect(delta.ReadIOs).To(Equal(uint64(200)))
			Expect(delta.WriteIOs).To(Equal(uint64(100)))
		})
	})

	Context("with identical samples", func() {
		BeforeEach(func() {
			// Make current sample identical to previous (except timestamp)
			cur = prev
			cur.Timestamp = prev.Timestamp.Add(1 * time.Second)
		})

		It("should return zero deltas for all counters", func() {
			delta := SampleDelta(cur, prev)

			Expect(delta.Elapsed).To(Equal(1 * time.Second))
			Expect(delta.ReadIOs).To(Equal(uint64(0)))
			Expect(delta.ReadMerges).To(Equal(uint64(0)))
			Expect(delta.ReadSectors).To(Equal(uint64(0)))
			Expect(delta.ReadTicks).To(Equal(uint64(0)))
			Expect(delta.WriteIOs).To(Equal(uint64(0)))
			Expect(delta.WriteMerges).To(Equal(uint64(0)))
			Expect(delta.WriteSectors).To(Equal(uint64(0)))
			Expect(delta.WriteTicks).To(Equal(uint64(0)))
			Expect(delta.IOsInProgress).To(Equal(uint64(5))) // Gauge: current value, not delta
			Expect(delta.IOsTotalTicks).To(Equal(uint64(0)))
			Expect(delta.WeightedIOTicks).To(Equal(uint64(0)))
			Expect(delta.DiscardIOs).To(Equal(uint64(0)))
			Expect(delta.DiscardMerges).To(Equal(uint64(0)))
			Expect(delta.DiscardSectors).To(Equal(uint64(0)))
			Expect(delta.DiscardTicks).To(Equal(uint64(0)))
			Expect(delta.FlushRequestsCompleted).To(Equal(uint64(0)))
			Expect(delta.TimeSpentFlushing).To(Equal(uint64(0)))
		})
	})

	Context("with negative time difference", func() {
		BeforeEach(func() {
			// Make current timestamp earlier than previous
			cur.Timestamp = prev.Timestamp.Add(-1 * time.Second)
		})

		It("should handle negative elapsed time", func() {
			delta := SampleDelta(cur, prev)
			Expect(delta.Elapsed).To(Equal(-1 * time.Second))
		})
	})

	Context("with very small time differences", func() {
		BeforeEach(func() {
			// 1 microsecond difference
			cur.Timestamp = prev.Timestamp.Add(1 * time.Microsecond)
		})

		It("should handle microsecond precision", func() {
			delta := SampleDelta(cur, prev)
			Expect(delta.Elapsed).To(Equal(1 * time.Microsecond))
		})
	})

	Context("with large counter values", func() {
		BeforeEach(func() {
			// Test with very large counter values
			prev.ReadIOs = 18446744073709551600 // Near uint64 max
			cur.ReadIOs = 18446744073709551615  // uint64 max
		})

		It("should handle large counter differences", func() {
			delta := SampleDelta(cur, prev)
			Expect(delta.ReadIOs).To(Equal(uint64(15)))
		})
	})

	Context("edge case: counter wrap-around", func() {
		BeforeEach(func() {
			// Simulate counter wrap-around (though this would be very rare)
			prev.ReadIOs = 18446744073709551615 // uint64 max
			cur.ReadIOs = 100                   // Wrapped around
		})

		It("should handle counter wrap-around mathematically", func() {
			delta := SampleDelta(cur, prev)
			// This will underflow and wrap around due to uint64 arithmetic
			// In practice, this scenario is extremely unlikely
			Expect(delta.ReadIOs).To(Equal(uint64(18446744073709551615 - 18446744073709551615 + 100 + 1)))
		})
	})

	Context("function purity", func() {
		It("should not modify input samples", func() {
			originalCur := cur
			originalPrev := prev

			_ = SampleDelta(cur, prev)

			// Verify inputs are unchanged
			Expect(cur).To(Equal(originalCur))
			Expect(prev).To(Equal(originalPrev))
		})

		It("should return consistent results for same inputs", func() {
			delta1 := SampleDelta(cur, prev)
			delta2 := SampleDelta(cur, prev)

			Expect(delta1).To(Equal(delta2))
		})

		It("should be commutative when swapping cur and prev with sign", func() {
			delta1 := SampleDelta(cur, prev)
			delta2 := SampleDelta(prev, cur)

			// The deltas should be negatives of each other (for most fields)
			// When we subtract 1200 - 1000 = 200, the reverse is 1000 - 1200 which underflows
			// uint64 underflow: 1000 - 1200 = 18446744073709551416 (due to uint64 wraparound)
			Expect(delta2.ReadIOs).To(Equal(uint64(18446744073709551416))) // Underflow: 1000 - 1200
			Expect(delta2.Elapsed).To(Equal(-delta1.Elapsed))
		})
	})
})
