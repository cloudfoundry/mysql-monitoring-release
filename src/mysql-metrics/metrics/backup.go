package metrics

import (
	"time"

	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewBackupMetricsCalculator(config config.Config, stats func() (time.Time, error), metricsFunc func(time.Time) *Metric) *backupMetricsCalculator {
	return &backupMetricsCalculator{
		emit:                 config.EmitBackupMetrics,
		backupStats:          stats,
		computeBackupMetrics: metricsFunc,
	}
}

type backupMetricsCalculator struct {
	emit                 bool
	backupStats          func() (time.Time, error)
	computeBackupMetrics func(time.Time) *Metric
}

func (t *backupMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emit {
		return nil, nil
	}

	backupStatMap, err := t.backupStats()
	if err != nil {
		return nil, err
	}
	return []*Metric{t.computeBackupMetrics(backupStatMap)}, nil
}
