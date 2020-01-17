package alert

import (
	"github.com/cloudfoundry/multierror"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Alerter
type Alerter interface {
	NotUnhealthy(time.Time) error
	Unhealthy(time.Time) error
}

type AggregateAlerter []Alerter

func (a AggregateAlerter) NotUnhealthy(now time.Time) error {
	errors := &multierror.MultiError{}

	for _, alerter := range a {
		e := alerter.NotUnhealthy(now)
		if e != nil {
			errors.Add(e)
		}
	}

	if errors.Length() > 0 {
		return errors
	}

	return nil
}

func (a AggregateAlerter) Unhealthy(now time.Time) error {
	errors := &multierror.MultiError{}

	for _, alerter := range a {
		e := alerter.Unhealthy(now)
		if e != nil {
			errors.Add(e)
		}
	}

	if errors.Length() > 0 {
		return errors
	}

	return nil
}
