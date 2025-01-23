package metrics

import (
	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewCPUMetricsCalculator(config config.Config, stats func() (map[string]string, error), metricsFunc func(map[string]string) []*Metric) *cpuMetricsCalculator {
	return &cpuMetricsCalculator{
		emit:              config.EmitCPUMetrics,
		cpuStats:          stats,
		computeCPUMetrics: metricsFunc,
	}
}

type cpuMetricsCalculator struct {
	emit              bool
	cpuStats          func() (map[string]string, error)
	computeCPUMetrics func(map[string]string) []*Metric
}

func (t *cpuMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emit {
		return nil, nil
	}

	cpuStatMap, err := t.cpuStats()
	if err != nil {
		return nil, err
	}
	return t.computeCPUMetrics(cpuStatMap), nil
}
