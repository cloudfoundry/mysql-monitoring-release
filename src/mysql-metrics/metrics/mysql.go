package metrics

import (
	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewMySQLMetricsCalculator(config config.Config, available func() bool, database func() (globalStatus map[string]string, globalVariables map[string]string, err error), availabilityFunc func(bool) *Metric, globalFunc func(map[string]string) []*Metric, metricsFunc func(map[string]string) []*Metric) *mysqlMetricsCalculator {
	return &mysqlMetricsCalculator{
		emitMySQL:        config.EmitMysqlMetrics,
		emitGalera:       config.EmitGaleraMetrics,
		available:        available,
		databaseMetadata: database,
		availabilityFunc: availabilityFunc,
		globalFunc:       globalFunc,
		galeraFunc:       metricsFunc,
	}
}

type mysqlMetricsCalculator struct {
	emitMySQL        bool
	emitGalera       bool
	available        func() bool
	databaseMetadata func() (globalStatus map[string]string, globalVariables map[string]string, err error)
	availabilityFunc func(bool) *Metric
	globalFunc       func(map[string]string) []*Metric
	galeraFunc       func(map[string]string) []*Metric
}

func (t *mysqlMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emitMySQL {
		return nil, nil
	}

	var metrics []*Metric

	isAvailable := t.available()
	availableMetrics := t.availabilityFunc(isAvailable)
	metrics = append(metrics, availableMetrics)

	if isAvailable {
		globalStatus, globalVariables, err := t.databaseMetadata()
		if err != nil {
			return metrics, err
		}

		globalStatusMetrics := t.globalFunc(globalStatus)
		metrics = append(metrics, globalStatusMetrics...)

		globalVariablesMetrics := t.globalFunc(globalVariables)
		metrics = append(metrics, globalVariablesMetrics...)

		if t.emitGalera {
			globalStatusGaleraMetrics := t.galeraFunc(globalStatus)
			metrics = append(metrics, globalStatusGaleraMetrics...)

			globalVariablesGaleraMetrics := t.galeraFunc(globalVariables)
			metrics = append(metrics, globalVariablesGaleraMetrics...)
		}
	}

	return metrics, nil
}
