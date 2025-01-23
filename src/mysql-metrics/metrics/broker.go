package metrics

import (
	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewBrokerMetricsCalculator(config config.Config, brokerStats func() (map[string]string, error), metricsFunc func(map[string]string) []*Metric) *brokerMetricsCalculator {
	return &brokerMetricsCalculator{
		emit:                 config.EmitBrokerMetrics,
		brokerStats:          brokerStats,
		computeBrokerMetrics: metricsFunc,
	}
}

type brokerMetricsCalculator struct {
	emit                 bool
	brokerStats          func() (map[string]string, error)
	computeBrokerMetrics func(map[string]string) []*Metric
}

func (t *brokerMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emit {
		return nil, nil
	}

	brokerStatMap, err := t.brokerStats()
	if err != nil {
		return nil, err
	}
	return t.computeBrokerMetrics(brokerStatMap), nil
}

type Calculator interface {
	Calculate() ([]*Metric, error)
}
