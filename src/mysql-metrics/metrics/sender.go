package metrics

import "github.com/cloudfoundry/dropsonde/metrics"

//go:generate counterfeiter . Sender
type Sender interface {
	SendValue(name string, value float64, unit string) error
}

type DropsondeSender struct{}

func (sender *DropsondeSender) SendValue(name string, value float64, unit string) error {
	return metrics.SendValue(name, value, unit)
}
