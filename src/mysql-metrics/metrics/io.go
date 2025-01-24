package metrics

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/procfs/blockdevice"
	"golang.org/x/sys/unix"

	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewIOMetricsCalculator(config config.Config, mappings map[string]MetricDefinition) *ioMetricsCalculator {
	return &ioMetricsCalculator{
		emit:           config.EmitCPUMetrics,
		metricMappings: mappings,
	}
}

type ioMetricsCalculator struct {
	emit           bool
	metricMappings map[string]MetricDefinition
}

func (t *ioMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emit {
		return nil, nil
	}

	return t.computeIOMetrics(), nil
}

// for test setup, see "mysql-metrics/bin/test" to configure docker container
// docker pull ghcr.io/cloudfoundry/ubuntu-jammy-stemcell:1.719
func (t *ioMetricsCalculator) computeIOMetrics() []*Metric {
	var metrics []*Metric

	fs, err := blockdevice.NewFS("/proc", "/sys")
	if err != nil {
		slog.Error("unable to create blockdevice", "error", err)
		return nil
	}
	diskStats, err := fs.ProcDiskstats()

	ephemeral := newDisk("/var/vcap/data")
	slog.Info("ephemeral info: /var/vcap/data", "ephemeral", ephemeral)

	persistent := newDisk("/var/vcap/store")
	slog.Info("persistent info: /var/vcap/store", "persistent", persistent)

	for _, stats := range diskStats {
		if stats.MajorNumber == ephemeral.major && stats.MinorNumber == ephemeral.minor {
			dev := stats.DeviceName
			slog.Info("reporting metrics for ephemeral device", "path", ephemeral.path, "stats.DeviceName", dev, "major", stats.MajorNumber, "minor", stats.MinorNumber)

			readIOs := stats.ReadIOs
			slog.Info("got ReadIOs", "stats.ReadIOS", readIOs)

			key := t.metricMappings["ephemeral_disk_read_ios"].Key
			unit := t.metricMappings["ephemeral_disk_read_ios"].Unit
			metrics = append(metrics, &Metric{Key: key, Value: float64(readIOs), Unit: unit})
		}

		if stats.MajorNumber == persistent.major && stats.MinorNumber == persistent.minor {
			dev := stats.DeviceName
			slog.Info("reporting metrics for persistent device", "path", persistent.path, "stats.DeviceName", dev, "major", stats.MajorNumber, "minor", stats.MinorNumber)

			readIOs := stats.ReadIOs
			slog.Info("got ReadIOs", "stats.ReadIOS", readIOs)

			key := t.metricMappings["persistent_disk_read_ios"].Key
			unit := t.metricMappings["persistent_disk_read_ios"].Unit
			metrics = append(metrics, &Metric{Key: key, Value: float64(readIOs), Unit: unit})
		}
	}

	return metrics
}

type disk struct {
	major uint32
	minor uint32
	path  string
}

func newDisk(path string) disk {
	stat := &unix.Stat_t{}
	err := unix.Stat(path, stat)
	if err != nil {
		slog.Error("unable to read ephemeral stats", "error", err)
	}
	return disk{
		major: unix.Major(uint64(stat.Dev)),
		minor: unix.Minor(uint64(stat.Dev)),
		path:  path,
	}
}

func (d disk) String() string {
	return fmt.Sprintf("major: %d, minor: %d, path: %s", d.major, d.minor, d.path)
}
