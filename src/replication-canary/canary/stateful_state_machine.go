package canary

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/lager"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=State
type State int

const (
	NotUnhealthy State = iota
	Unhealthy
)

type StatefulStateMachine struct {
	State                 State
	OnBecomesUnhealthy    func(time.Time)
	OnBecomesNotUnhealthy func(time.Time)
	Logger                lager.Logger
}

func (m *StatefulStateMachine) BecomesUnhealthy(timestamp time.Time) {
	if m.State == NotUnhealthy {
		m.Logger.Debug("StateMachine Transitioning to unhealthy", lager.Data{"timestamp": timestamp})
		m.OnBecomesUnhealthy(timestamp)
	}

	m.Logger.Debug("StateMachine unhealthy", lager.Data{"timestamp": timestamp})

	m.State = Unhealthy
}

func (m *StatefulStateMachine) BecomesNotUnhealthy(timestamp time.Time) {
	if m.State == Unhealthy {
		m.Logger.Debug("StateMachine Transitioning to not unhealthy", lager.Data{"timestamp": timestamp})
		m.OnBecomesNotUnhealthy(timestamp)
	}

	m.Logger.Debug("StateMachine not unhealthy", lager.Data{"timestamp": timestamp})

	m.State = NotUnhealthy
}

func (m *StatefulStateMachine) RemainsInSameState(timestamp time.Time) {
	m.Logger.Debug(
		"StateMachine remaining in the same state",
		lager.Data{"timestamp": timestamp, "state": fmt.Sprint(m.State)},
	)
}

func (m *StatefulStateMachine) GetState() State {
	return m.State
}
