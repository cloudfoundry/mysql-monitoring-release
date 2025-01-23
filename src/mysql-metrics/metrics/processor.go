package metrics

import (
	"time"

	"github.com/cloudfoundry/mysql-metrics/config"

	"github.com/hashicorp/go-multierror"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Gatherer
type Gatherer interface {
	DatabaseMetadata() (globalStatus map[string]string, globalVariables map[string]string, err error)
	FollowerMetadata() (slaveStatus map[string]string, heartbeatStatus map[string]string, err error)
	IsDatabaseFollower() (bool, error)
	IsDatabaseAvailable() bool
	DiskStats() (map[string]string, error)
	BrokerStats() (map[string]string, error)
	CPUStats() (map[string]string, error)
	FindLastBackupTimestamp() (time.Time, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . MetricsComputer
type MetricsComputer interface {
	ComputeAvailabilityMetric(bool) *Metric
	ComputeIsFollowerMetric(bool) *Metric
	ComputeGlobalMetrics(map[string]string) []*Metric
	ComputeLeaderFollowerMetrics(map[string]string) []*Metric
	ComputeDiskMetrics(map[string]string) []*Metric
	ComputeBrokerMetrics(map[string]string) []*Metric
	ComputeGaleraMetrics(map[string]string) []*Metric
	ComputeCPUMetrics(map[string]string) []*Metric
	ComputeBackupMetric(time.Time) *Metric
}

type Processor struct {
	metricsWriter Writer
	pipeline      []Calculator
}

func NewProcessor(
	gatherer Gatherer,
	metricsComputer MetricsComputer,
	metricsWriter Writer,
	configuration *config.Config,
	metricMappingConfig MetricMappingConfig,
) Processor {
	processor := Processor{
		metricsWriter: metricsWriter,
	}
	processor.pipeline = []Calculator{
		NewDiskMetricsCalculator(*configuration, gatherer.DiskStats, metricsComputer.ComputeDiskMetrics),
		NewBrokerMetricsCalculator(*configuration, gatherer.BrokerStats, metricsComputer.ComputeBrokerMetrics),
		NewCPUMetricsCalculator(*configuration, gatherer.CPUStats, metricsComputer.ComputeCPUMetrics),
		NewBackupMetricsCalculator(*configuration, gatherer.FindLastBackupTimestamp, metricsComputer.ComputeBackupMetric),
		NewMySQLMetricsCalculator(*configuration, gatherer.IsDatabaseAvailable, gatherer.DatabaseMetadata, metricsComputer.ComputeAvailabilityMetric, metricsComputer.ComputeGlobalMetrics, metricsComputer.ComputeGaleraMetrics),
		NewLeaderFollowerMetricsCalculator(*configuration, gatherer.IsDatabaseFollower, gatherer.FollowerMetadata, metricsComputer.ComputeIsFollowerMetric, metricsComputer.ComputeLeaderFollowerMetrics),
		NewIOMetricsCalculator(*configuration, metricMappingConfig.IOMetricMappings),
	}
	return processor
}

func (p Processor) Process() error {
	var collectedMetrics []*Metric
	var collectedErrors error
	for _, calculator := range p.pipeline {
		metrics, err := calculator.Calculate()
		if err != nil {
			collectedErrors = multierror.Append(collectedErrors, err)
		}
		collectedMetrics = append(collectedMetrics, metrics...)
	}

	if err := p.metricsWriter.Write(collectedMetrics); err != nil {
		collectedErrors = multierror.Append(collectedErrors, err)
	}

	return collectedErrors
}
