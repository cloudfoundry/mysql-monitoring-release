package alert

import (
	"time"

	"code.cloudfoundry.org/lager"
)

type LoggingAlerter struct {
	Logger lager.Logger
}

func (a *LoggingAlerter) Unhealthy(timestamp time.Time) error {
	a.Logger.Debug("Logging alerter logging unhealthy", lager.Data{"timestamp": timestamp})
	a.Logger.Error("cluster is unhealthy", nil, lager.Data{
		"timestamp": timestamp,
	})

	return nil
}

func (a *LoggingAlerter) NotUnhealthy(timestamp time.Time) error {
	a.Logger.Debug("Logging alerter logging not unhealthy", lager.Data{"timestamp": timestamp})
	a.Logger.Info("cluster is not unhealthy", lager.Data{
		"timestamp": timestamp,
	})

	return nil
}
