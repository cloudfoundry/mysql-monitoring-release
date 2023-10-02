package alert

import (
	"log/slog"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SwitchboardClient
type SwitchboardClient interface {
	EnableClusterTraffic() error
	DisableClusterTraffic() error
}

type SwitchboardAlerter struct {
	Logger            *slog.Logger
	SwitchboardClient SwitchboardClient
	NoOp              bool
}

func (s *SwitchboardAlerter) NotUnhealthy(timestamp time.Time) error {
	if s.NoOp {
		s.Logger.Debug("Switchboard alerter configured to no-op")
		return nil
	}
	s.Logger.Debug("Switchboard alerter enabling traffic")
	return s.SwitchboardClient.EnableClusterTraffic()
}

func (s *SwitchboardAlerter) Unhealthy(timestamp time.Time) error {
	if s.NoOp {
		s.Logger.Debug("Switchboard alerter configured to no-op")
		return nil
	}
	s.Logger.Debug("Switchboard alerter disabling traffic")
	return s.SwitchboardClient.DisableClusterTraffic()
}
