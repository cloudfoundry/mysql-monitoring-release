package metrics

import (
	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewDiskMetricsCalculator(config config.Config, stats func() (map[string]string, error), metricsFunc func(map[string]string) []*Metric) *diskMetricsCalculator {
	return &diskMetricsCalculator{
		emit:               config.EmitDiskMetrics,
		diskStats:          stats,
		computeDiskMetrics: metricsFunc,
	}
}

type diskMetricsCalculator struct {
	emit               bool
	diskStats          func() (map[string]string, error)
	computeDiskMetrics func(map[string]string) []*Metric
}

func (t *diskMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emit {
		return nil, nil
	}

	diskStatMap, err := t.diskStats()
	if err != nil {
		return nil, err
	}
	return t.computeDiskMetrics(diskStatMap), nil
}
