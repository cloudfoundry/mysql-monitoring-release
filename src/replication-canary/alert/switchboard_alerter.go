package alert

import (
	"time"

	"code.cloudfoundry.org/lager"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SwitchboardClient
type SwitchboardClient interface {
	EnableClusterTraffic() error
	DisableClusterTraffic() error
}

type SwitchboardAlerter struct {
	Logger            lager.Logger
	SwitchboardClient SwitchboardClient
	NoOp              bool
}

func (s *SwitchboardAlerter) NotUnhealthy(timestamp time.Time) error {
	if s.NoOp {
		s.Logger.Debug("Switchboard alerter configured to no-op", lager.Data{"timestamp": timestamp})
		return nil
	}
	s.Logger.Debug("Switchboard alerter enabling traffic", lager.Data{"timestamp": timestamp})
	return s.SwitchboardClient.EnableClusterTraffic()
}

func (s *SwitchboardAlerter) Unhealthy(timestamp time.Time) error {
	if s.NoOp {
		s.Logger.Debug("Switchboard alerter configured to no-op", lager.Data{"timestamp": timestamp})
		return nil
	}
	s.Logger.Debug("Switchboard alerter disabling traffic", lager.Data{"timestamp": timestamp})
	return s.SwitchboardClient.DisableClusterTraffic()
}
