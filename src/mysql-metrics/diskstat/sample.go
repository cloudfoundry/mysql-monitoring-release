package diskstat

import (
	"time"

	"github.com/prometheus/procfs/blockdevice"
)

type Stats = blockdevice.Diskstats
type Sample struct {
	Timestamp time.Time
	Stats
}

func SampleDelta(cur, prev Sample) (delta Delta) {
	delta.Elapsed = cur.Timestamp.Sub(prev.Timestamp)
	delta.ReadIOs = cur.ReadIOs - prev.ReadIOs
	delta.ReadMerges = cur.ReadMerges - prev.ReadMerges
	delta.ReadSectors = cur.ReadSectors - prev.ReadSectors
	delta.ReadTicks = cur.ReadTicks - prev.ReadTicks
	delta.WriteIOs = cur.WriteIOs - prev.WriteIOs
	delta.WriteMerges = cur.WriteMerges - prev.WriteMerges
	delta.WriteSectors = cur.WriteSectors - prev.WriteSectors
	delta.WriteTicks = cur.WriteTicks - prev.WriteTicks

	// IOsInProgress is a gauge (current snapshot), not a counter
	delta.IOsInProgress = cur.IOsInProgress

	delta.IOsTotalTicks = cur.IOsTotalTicks - prev.IOsTotalTicks
	delta.WeightedIOTicks = cur.WeightedIOTicks - prev.WeightedIOTicks
	delta.DiscardIOs = cur.DiscardIOs - prev.DiscardIOs
	delta.DiscardMerges = cur.DiscardMerges - prev.DiscardMerges
	delta.DiscardSectors = cur.DiscardSectors - prev.DiscardSectors
	delta.DiscardTicks = cur.DiscardTicks - prev.DiscardTicks
	delta.FlushRequestsCompleted = cur.FlushRequestsCompleted - prev.FlushRequestsCompleted
	delta.TimeSpentFlushing = cur.TimeSpentFlushing - prev.TimeSpentFlushing

	return delta
}
