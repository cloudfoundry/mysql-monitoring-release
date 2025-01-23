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
	gatherer        Gatherer
	metricsComputer MetricsComputer
	metricsWriter   Writer
	config          *config.Config
	mapping         MetricMappingConfig
}

func NewProcessor(
	gatherer Gatherer,
	metricsComputer MetricsComputer,
	metricsWriter Writer,
	configuration *config.Config,
	metricMappingConfig MetricMappingConfig,
) Processor {
	return Processor{
		gatherer:        gatherer,
		metricsComputer: metricsComputer,
		metricsWriter:   metricsWriter,
		config:          configuration,
		mapping:         metricMappingConfig,
	}
}

func (p Processor) Process() error {
	calculatorPipeline := []Calculator{
		NewDiskMetricsCalculator(*p.config, p.gatherer.DiskStats, p.metricsComputer.ComputeDiskMetrics),
		NewBrokerMetricsCalculator(*p.config, p.gatherer.BrokerStats, p.metricsComputer.ComputeBrokerMetrics),
		NewCPUMetricsCalculator(*p.config, p.gatherer.CPUStats, p.metricsComputer.ComputeCPUMetrics),
		NewBackupMetricsCalculator(*p.config, p.gatherer.FindLastBackupTimestamp, p.metricsComputer.ComputeBackupMetric),
		NewMySQLMetricsCalculator(*p.config, p.gatherer.IsDatabaseAvailable, p.gatherer.DatabaseMetadata, p.metricsComputer.ComputeAvailabilityMetric, p.metricsComputer.ComputeGlobalMetrics, p.metricsComputer.ComputeGaleraMetrics),
		NewLeaderFollowerMetricsCalculator(*p.config, p.gatherer.IsDatabaseFollower, p.gatherer.FollowerMetadata, p.metricsComputer.ComputeIsFollowerMetric, p.metricsComputer.ComputeLeaderFollowerMetrics),
		NewIOMetricsCalculator(*p.config, p.mapping.IOMetricMappings),
	}

	var collectedMetrics []*Metric
	var collectedErrors error
	for _, calculator := range calculatorPipeline {
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
