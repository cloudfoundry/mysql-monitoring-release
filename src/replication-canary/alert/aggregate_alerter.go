package alert

import (
	"github.com/cloudfoundry/multierror"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Alerter
type Alerter interface {
	NotUnhealthy() error
	Unhealthy() error
}

type AggregateAlerter []Alerter

func (a AggregateAlerter) NotUnhealthy() error {
	errors := &multierror.MultiError{}

	for _, alerter := range a {
		e := alerter.NotUnhealthy()
		if e != nil {
			errors.Add(e)
		}
	}

	if errors.Length() > 0 {
		return errors
	}

	return nil
}

func (a AggregateAlerter) Unhealthy() error {
	errors := &multierror.MultiError{}

	for _, alerter := range a {
		e := alerter.Unhealthy()
		if e != nil {
			errors.Add(e)
		}
	}

	if errors.Length() > 0 {
		return errors
	}

	return nil
}
