package canary

import "time"

type StatelessStateMachine struct {
	OnBecomesUnhealthy    func(time.Time)
	OnBecomesNotUnhealthy func(time.Time)
}

func (m StatelessStateMachine) BecomesUnhealthy(now time.Time) {
	m.OnBecomesUnhealthy(now)

}

func (m StatelessStateMachine) BecomesNotUnhealthy(now time.Time) {
	m.OnBecomesNotUnhealthy(now)
}
