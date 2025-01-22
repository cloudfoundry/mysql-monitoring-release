package metrics_computer

// TODO move into metrics package

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/procfs/blockdevice"

	"github.com/cloudfoundry/mysql-metrics/metrics"
)

type MetricsComputer struct {
	metricMappingConfig metrics.MetricMappingConfig
}

func NewMetricsComputer(metricMappingConfig metrics.MetricMappingConfig) *MetricsComputer {
	return &MetricsComputer{
		metricMappingConfig: metricMappingConfig,
	}
}

func (mc *MetricsComputer) ComputeIsFollowerMetric(isFollower bool) *metrics.Metric {
	isFollowerAsFloat := 0.0
	if isFollower {
		isFollowerAsFloat = 1.0
	}
	key := mc.metricMappingConfig.LeaderFollowerMetricMappings["is_follower"].Key
	unit := mc.metricMappingConfig.LeaderFollowerMetricMappings["is_follower"].Unit

	return &metrics.Metric{Key: key, Value: isFollowerAsFloat, Unit: unit}
}

func (mc *MetricsComputer) ComputeAvailabilityMetric(isAvailable bool) *metrics.Metric {
	availableAsFloat := 0.0
	if isAvailable {
		availableAsFloat = 1.0
	}
	key := mc.metricMappingConfig.MysqlMetricMappings["available"].Key
	unit := mc.metricMappingConfig.MysqlMetricMappings["available"].Unit

	return &metrics.Metric{Key: key, Value: availableAsFloat, Unit: unit}
}

func (mc *MetricsComputer) ComputeLeaderFollowerMetrics(values map[string]string) []*metrics.Metric {
	return mc.ComputeMetricsFromMapping(values, mc.metricMappingConfig.LeaderFollowerMetricMappings)
}

func (mc *MetricsComputer) ComputeGlobalMetrics(values map[string]string) []*metrics.Metric {
	return mc.ComputeMetricsFromMapping(values, mc.metricMappingConfig.MysqlMetricMappings)
}

func (mc *MetricsComputer) ComputeDiskMetrics(values map[string]string) []*metrics.Metric {
	return mc.ComputeMetricsFromMapping(values, mc.metricMappingConfig.DiskMetricMappings)
}

func (mc *MetricsComputer) ComputeGaleraMetrics(values map[string]string) []*metrics.Metric {
	return mc.ComputeMetricsFromMapping(values, mc.metricMappingConfig.GaleraMetricMappings)
}

func (mc *MetricsComputer) ComputeBrokerMetrics(values map[string]string) []*metrics.Metric {
	return mc.ComputeMetricsFromMapping(values, mc.metricMappingConfig.BrokerMetricMappings)
}

func (mc *MetricsComputer) ComputeCPUMetrics(values map[string]string) []*metrics.Metric {
	return mc.ComputeMetricsFromMapping(values, mc.metricMappingConfig.CPUMetricMappings)
}

func (mc *MetricsComputer) ComputeMetricsFromMapping(metricValues map[string]string, mappingConfig map[string]metrics.MetricDefinition) []*metrics.Metric {
	var gatheredMetrics []*metrics.Metric
	for metricName, mapping := range mappingConfig {
		rawValue, found := metricValues[metricName]
		if found {
			floatValue, err := mc.parseMetricValue(rawValue)

			metric := metrics.Metric{
				Key:      mapping.Key,
				Unit:     mapping.Unit,
				RawValue: rawValue,
				Value:    floatValue,
				Error:    err,
			}
			gatheredMetrics = append(gatheredMetrics, &metric)
		}
	}
	return gatheredMetrics
}

func (mc *MetricsComputer) ComputeBackupMetric(backupTimestamp time.Time) *metrics.Metric {
	backupTimestampSeconds := float64(backupTimestamp.Unix())
	key := mc.metricMappingConfig.BackupMetricMappings["last_successful_backup"].Key
	unit := mc.metricMappingConfig.BackupMetricMappings["last_successful_backup"].Unit
	return &metrics.Metric{Key: key, Value: float64(backupTimestampSeconds), Unit: unit}
}

func (mc *MetricsComputer) ComputeIOMetrics() []*metrics.Metric {

	fs, err := blockdevice.NewFS("/proc", "/sys")
	if err != nil {
		slog.Error("unable to create blockdevice", err)
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
		dev := stats.DeviceName
		slog.Info("got device name", "stats.DeviceName", dev)

		readIOS := stats.ReadIOs
		slog.Info("got ReadIOs", "stats.ReadIOS", readIOS)
	}

	readIOs := 100
	key := mc.metricMappingConfig.IOMetricMappings["persistent_disk_read_ios"].Key
	unit := mc.metricMappingConfig.IOMetricMappings["persistent_disk_read_ios"].Unit

	return []*metrics.Metric{
		{Key: key, Value: float64(readIOs), Unit: unit},
	}
}

func (mc *MetricsComputer) parseMetricValue(rawValue string) (float64, error) {
	floatValue, err := mc.parseFloat(rawValue)
	if err != nil {
		floatValue, err = mc.parseStringValue(rawValue)
	}
	if err != nil {
		floatValue, err = mc.parseClusterStatus(rawValue)
	}

	return floatValue, err
}

func (mc *MetricsComputer) parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func (mc *MetricsComputer) parseStringValue(value string) (float64, error) {
	var floatValue float64
	var err error

	switch strings.ToLower(value) {
	case "on":
		floatValue = 1.0
	case "off":
		floatValue = 0.0
	case "yes":
		floatValue = 1.0
	case "no":
		floatValue = 0.0
	case "null":
		floatValue = -1.0
	default:
		err = errors.New("could not convert")
	}

	return floatValue, err
}

func (mc *MetricsComputer) parseClusterStatus(value string) (float64, error) {
	var floatValue float64
	var err error

	switch strings.ToLower(value) {
	case "primary":
		floatValue = 1
	case "non-primary":
		floatValue = 0
	case "disconnected":
		floatValue = -1
	default:
		err = errors.New("could not convert raw value")
	}

	return floatValue, err
}
