package metrics

import (
	"github.com/cloudfoundry/mysql-metrics/config"

	"time"

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
}

func NewProcessor(
	gatherer Gatherer,
	metricsComputer MetricsComputer,
	metricsWriter Writer,
	configuration *config.Config,
) Processor {
	return Processor{
		gatherer:        gatherer,
		metricsComputer: metricsComputer,
		metricsWriter:   metricsWriter,
		config:          configuration,
	}
}

func (p Processor) Process() error {
	var collectedMetrics []*Metric
	var collectedErrors error

	if p.config.EmitDiskMetrics {
		diskStatMap, err := p.gatherer.DiskStats()
		if err != nil {
			collectedErrors = multierror.Append(collectedErrors, err)
		}
		collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeDiskMetrics(diskStatMap)...)
	}

	if p.config.EmitBrokerMetrics {
		brokerStatMap, err := p.gatherer.BrokerStats()
		if err != nil {
			collectedErrors = multierror.Append(collectedErrors, err)
		}
		collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeBrokerMetrics(brokerStatMap)...)
	}

	if p.config.EmitCPUMetrics {
		cpuStatMap, err := p.gatherer.CPUStats()
		if err != nil {
			collectedErrors = multierror.Append(collectedErrors, err)
		}
		collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeCPUMetrics(cpuStatMap)...)
	}

	if p.config.EmitBackupMetrics {
		backupTimestamp, err := p.gatherer.FindLastBackupTimestamp()
		if err != nil {
			collectedErrors = multierror.Append(collectedErrors, err)
		} else {
			collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeBackupMetric(backupTimestamp))
		}
	}

	isAvailable := p.gatherer.IsDatabaseAvailable()
	if p.config.EmitMysqlMetrics {
		collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeAvailabilityMetric(isAvailable))

		if isAvailable {
			globalStatus, globalVariables, err := p.gatherer.DatabaseMetadata()
			if err != nil {
				collectedErrors = multierror.Append(collectedErrors, err)
			}

			collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeGlobalMetrics(globalStatus)...)
			collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeGlobalMetrics(globalVariables)...)

			if p.config.EmitGaleraMetrics {
				collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeGaleraMetrics(globalStatus)...)
				collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeGaleraMetrics(globalVariables)...)
			}
		}
	}

	if p.config.EmitLeaderFollowerMetrics {
		isFollower, err := p.gatherer.IsDatabaseFollower()
		if err != nil {
			collectedErrors = multierror.Append(collectedErrors, err)
		}

		collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeIsFollowerMetric(isFollower))

		if isFollower {
			slaveStatus, heartbeatStatus, err := p.gatherer.FollowerMetadata()
			if err != nil {
				collectedErrors = multierror.Append(collectedErrors, err)
			}

			collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeLeaderFollowerMetrics(heartbeatStatus)...)
			collectedMetrics = append(collectedMetrics, p.metricsComputer.ComputeLeaderFollowerMetrics(slaveStatus)...)
		}
	}

	if err := p.metricsWriter.Write(collectedMetrics); err != nil {
		collectedErrors = multierror.Append(collectedErrors, err)
	}

	return collectedErrors
}
