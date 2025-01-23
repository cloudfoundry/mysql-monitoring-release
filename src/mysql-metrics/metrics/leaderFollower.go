package metrics

import (
	"github.com/cloudfoundry/mysql-metrics/config"
)

func NewLeaderFollowerMetricsCalculator(config config.Config, isDatabaseFollower func() (bool, error), followerMetadata func() (slaveStatus map[string]string, heartbeatStatus map[string]string, err error), isFollowerMetric func(bool) *Metric, metrics func(map[string]string) []*Metric) *leaderFollowerMetricsCalculator {
	return &leaderFollowerMetricsCalculator{
		emit:                         config.EmitLeaderFollowerMetrics,
		isDatabaseFollower:           isDatabaseFollower,
		followerMetadata:             followerMetadata,
		isFollowerMetric:             isFollowerMetric,
		computeLeaderFollowerMetrics: metrics,
	}
}

type leaderFollowerMetricsCalculator struct {
	emit                         bool
	isDatabaseFollower           func() (bool, error)
	followerMetadata             func() (slaveStatus map[string]string, heartbeatStatus map[string]string, err error)
	isFollowerMetric             func(bool) *Metric
	computeLeaderFollowerMetrics func(map[string]string) []*Metric
}

func (t *leaderFollowerMetricsCalculator) Calculate() ([]*Metric, error) {
	if !t.emit {
		return nil, nil
	}

	isFollower, err := t.isDatabaseFollower()
	if err != nil {
		return nil, err
	}

	var metrics []*Metric
	isFollowerMetric := t.isFollowerMetric(isFollower)
	metrics = append(metrics, isFollowerMetric)

	if isFollower {
		slaveStatus, heartbeatStatus, err := t.followerMetadata()
		if err != nil {
			return metrics, err
		}

		metrics = append(metrics, t.computeLeaderFollowerMetrics(heartbeatStatus)...)
		metrics = append(metrics, t.computeLeaderFollowerMetrics(slaveStatus)...)
	}
	return metrics, nil
}
