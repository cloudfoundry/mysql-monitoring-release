package metrics_computer

// TODO move into metrics package

import (
	"errors"
	"strconv"
	"time"

	"strings"

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

	return &metrics.Metric{}
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
