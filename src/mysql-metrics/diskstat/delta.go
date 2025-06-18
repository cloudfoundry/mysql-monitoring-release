package diskstat

import (
	"time"
)

const (
	sectorSizeBytes   = 512
	bytesPerKiB       = 1024
	sectorsPerMiB     = 2048 // 1 MiB = 2048 sectors (2048 * 512 bytes = 1 MiB)
	percentMultiplier = 100.0
)

// Delta represents the change in disk I/O statistics over a time period.
// It contains the elapsed time and the delta values for all disk metrics.
type Delta struct {
	// Elapsed represents the time interval over which this delta was computed
	Elapsed time.Duration

	// Diskstats provide a delta of disk statistics
	Stats
}

// ReadsPerSecond returns the number of read operations per second.
func (d Delta) ReadsPerSecond() float64 {
	return float64(d.ReadIOs) / d.Elapsed.Seconds()
}

// ReadKiB returns the total kilobytes read.
func (d Delta) ReadKiB() float64 {
	return float64(d.ReadSectors*sectorSizeBytes) / bytesPerKiB
}

// ReadAvgKB returns the average kilobytes per read operation.
func (d Delta) ReadAvgKB() float64 {
	if d.ReadIOs == 0 {
		return 0.0
	}
	return d.ReadKiB() / float64(d.ReadIOs)
}

// ReadMiBPerSec returns the read throughput in MiB per second.
func (d Delta) ReadMiBPerSec() float64 {
	return float64(d.ReadSectors) / sectorsPerMiB / d.Elapsed.Seconds()
}

// ReadMergesPercent returns the percentage of read operations that were merged.
func (d Delta) ReadMergesPercent() float64 {
	if d.ReadRequests() == 0 {
		return 0.0
	}
	return percentMultiplier * float64(d.ReadMerges) / float64(d.ReadRequests())
}

// ReadResponseTime returns the average response time for read operations in milliseconds.
func (d Delta) ReadResponseTime() float64 {
	if d.ReadRequests() == 0 {
		return 0.0
	}
	return float64(d.ReadTicks) / float64(d.ReadRequests())
}

// ReadConcurrency returns the read concurrency level.
func (d Delta) ReadConcurrency() float64 {
	return float64(d.ReadTicks) / float64(d.Elapsed.Milliseconds())
}

// WritesPerSecond returns the number of write operations per second.
func (d Delta) WritesPerSecond() float64 {
	return float64(d.WriteIOs) / d.Elapsed.Seconds()
}

// WriteKiB returns the total kilobytes written.
func (d Delta) WriteKiB() float64 {
	return float64(d.WriteSectors*sectorSizeBytes) / bytesPerKiB
}

// WriteAvgKB returns the average kilobytes per write operation.
func (d Delta) WriteAvgKB() float64 {
	if d.WriteIOs == 0 {
		return 0.0
	}
	return d.WriteKiB() / float64(d.WriteIOs)
}

// WriteMiBPerSec returns the write throughput in MiB per second.
func (d Delta) WriteMiBPerSec() float64 {
	return float64(d.WriteSectors) / sectorsPerMiB / d.Elapsed.Seconds()
}

// WriteConcurrency returns the write concurrency level.
func (d Delta) WriteConcurrency() float64 {
	return float64(d.WriteTicks) / float64(d.Elapsed.Milliseconds())
}

// WriteMergesPercent returns the percentage of write operations that were merged.
func (d Delta) WriteMergesPercent() float64 {
	if d.WriteRequests() == 0 {
		return 0.0
	}
	return percentMultiplier * float64(d.WriteMerges) / float64(d.WriteRequests())
}

// WriteResponseTime returns the average response time for write operations in milliseconds.
func (d Delta) WriteResponseTime() float64 {
	if d.WriteRequests() == 0 {
		return 0.0
	}
	return float64(d.WriteTicks) / float64(d.WriteRequests())
}

// IOsRequests returns the total number of I/O requests (reads + writes).
func (d Delta) IOsRequests() uint64 {
	return d.ReadRequests() + d.WriteRequests()
}

// ReadRequests returns the total number of read requests (including merges).
func (d Delta) ReadRequests() uint64 {
	return d.ReadIOs + d.ReadMerges
}

// WriteRequests returns the total number of write requests (including merges).
func (d Delta) WriteRequests() uint64 {
	return d.WriteIOs + d.WriteMerges
}

// AvgResponseTime returns the average response time in milliseconds.
// This represents the total time from request submission to completion.
func (d Delta) AvgResponseTime() float64 {
	avgIOs := d.IOsRequests() + d.IOsInProgress
	if avgIOs == 0 {
		return 0.0
	}
	return float64(d.WeightedIOTicks) / float64(avgIOs)
}

// AvgServiceTime returns the average service time in milliseconds.
// This represents the time spent actively servicing requests.
func (d Delta) AvgServiceTime() float64 {
	totalRequests := d.IOsRequests()
	if totalRequests == 0 {
		return 0.0
	}
	return float64(d.IOsTotalTicks) / float64(totalRequests)
}

// QTime returns the average queue time in milliseconds.
// This is the difference between response time and service time.
func (d Delta) QTime() float64 {
	return d.AvgResponseTime() - d.AvgServiceTime()
}

// BusyPercent returns the device utilization percentage.
func (d Delta) BusyPercent() float64 {
	if d.Elapsed.Milliseconds() == 0 {
		return 0.0
	}
	// Convert IOsTotalTicks (in milliseconds) to percentage of elapsed time
	return (float64(d.IOsTotalTicks) / float64(d.Elapsed.Milliseconds())) * percentMultiplier
}

// DiscardsPerSecond returns the number of discard operations per second.
func (d Delta) DiscardsPerSecond() float64 {
	return float64(d.DiscardIOs) / d.Elapsed.Seconds()
}

// DiscardKiB returns the total kilobytes discarded.
func (d Delta) DiscardKiB() float64 {
	return float64(d.DiscardSectors*sectorSizeBytes) / bytesPerKiB
}

// DiscardMiBPerSec returns the discard throughput in MiB per second.
func (d Delta) DiscardMiBPerSec() float64 {
	return float64(d.DiscardSectors) / sectorsPerMiB / d.Elapsed.Seconds()
}

// DiscardAvgKB returns the average kilobytes per discard operation.
func (d Delta) DiscardAvgKB() float64 {
	if d.DiscardIOs == 0 {
		return 0.0
	}
	return d.DiscardKiB() / float64(d.DiscardIOs)
}

// DiscardRequests returns the total number of discard requests (including merges).
func (d Delta) DiscardRequests() uint64 {
	return d.DiscardIOs + d.DiscardMerges
}

// DiscardMergesPercent returns the percentage of discard operations that were merged.
func (d Delta) DiscardMergesPercent() float64 {
	if d.DiscardRequests() == 0 {
		return 0.0
	}
	return percentMultiplier * float64(d.DiscardMerges) / float64(d.DiscardRequests())
}

// DiscardResponseTime returns the average response time for discard operations in milliseconds.
func (d Delta) DiscardResponseTime() float64 {
	if d.DiscardRequests() == 0 {
		return 0.0
	}
	return float64(d.DiscardTicks) / float64(d.DiscardRequests())
}

// DiscardConcurrency returns the discard concurrency level.
func (d Delta) DiscardConcurrency() float64 {
	return float64(d.DiscardTicks) / float64(d.Elapsed.Milliseconds())
}

// FlushesPerSecond returns the number of flush operations per second.
func (d Delta) FlushesPerSecond() float64 {
	return float64(d.FlushRequestsCompleted) / d.Elapsed.Seconds()
}

// FlushResponseTime returns the average response time for flush operations in milliseconds.
func (d Delta) FlushResponseTime() float64 {
	if d.FlushRequestsCompleted == 0 {
		return 0.0
	}
	return float64(d.TimeSpentFlushing) / float64(d.FlushRequestsCompleted)
}

// FlushConcurrency returns the flush concurrency level.
func (d Delta) FlushConcurrency() float64 {
	return float64(d.TimeSpentFlushing) / float64(d.Elapsed.Milliseconds())
}
