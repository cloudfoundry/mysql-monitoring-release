package alert

import (
	"log/slog"
)

type LoggingAlerter struct {
	Logger *slog.Logger
}

func (a *LoggingAlerter) Unhealthy() error {
	a.Logger.Debug("Logging alerter logging unhealthy")
	a.Logger.Error("cluster is unhealthy")

	return nil
}

func (a *LoggingAlerter) NotUnhealthy() error {
	a.Logger.Debug("Logging alerter logging not unhealthy")
	a.Logger.Info("cluster is not unhealthy")

	return nil
}
