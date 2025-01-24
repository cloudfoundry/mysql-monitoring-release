package metrics

import (
	"log/slog"

	"github.com/prometheus/procfs/blockdevice"

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

func (t *ioMetricsCalculator) computeIOMetrics() []*Metric {
	fs, err := blockdevice.NewFS("/proc", "/sys")
	if err != nil {
		slog.Error("unable to create blockdevice", "error", err)
		return nil
	}
	diskStats, err := fs.ProcDiskstats()
	for _, stats := range diskStats {
		//loop0-loop7,nvme0n1,nvme0n1p1,nvme0n1p2,sda,sda1,sdb,sdb1

		//# df -h
		//Filesystem      Size  Used Avail Use% Mounted on
		//overlay         303G  3.7G  299G   2% /
		//tmpfs            15G     0   15G   0% /dev
		//tmpfs            15G     0   15G   0% /dev/shm
		///dev/nvme0n1p2  340G  8.3G  314G   3% /tmp/garden-init
		///dev/sdb1       192G   28G  155G  16% /var/vcap/data
		//tmpfs            16M  332K   16M   3% /var/vcap/data/sys/run
		///dev/loop6      9.6G  2.6G  6.5G  29% /var/vcap/store
		//tmpfs            15G     0   15G   0% /sys/fs/cgroup

		//persistent = /var/vcap/store
		//ephemeral = /var/vcap/data

		//TODO: try 'syscal.Stat /var/vcap/store'
		// => { "ephemeral": "dev/sdb1", "persistent": "dev/loop6"]

		//unix.Stat("/var/vcap/store") =>Stat.dev => Major(Stat.dev), Minor(Stat.dev)

		dev := stats.DeviceName
		slog.Info("got device name", "stats.DeviceName", dev)
		//	MajorNumber uint32
		//	MinorNumber uint32
		// these map to stat's device number
		// for test setup, see "mysql-metrics/bin/test" to configure docker container
		// docker pull ghcr.io/cloudfoundry/ubuntu-jammy-stemcell:1.719

		readIOS := stats.ReadIOs
		slog.Info("got ReadIOs", "stats.ReadIOS", readIOS)
	}

	readIOs := 100
	key := t.metricMappings["persistent_disk_read_ios"].Key
	unit := t.metricMappings["persistent_disk_read_ios"].Unit

	return []*Metric{
		{Key: key, Value: float64(readIOs), Unit: unit},
	}
}
