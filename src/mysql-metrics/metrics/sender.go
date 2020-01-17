package metrics

import (
	"code.cloudfoundry.org/go-loggregator"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Sender
type Sender interface {
	SendValue(name string, value float64, unit string) error
}

type LoggregatorSender struct {
	client   *loggregator.IngressClient
	sourceID string
}

func NewLoggregatorSender(client *loggregator.IngressClient, sourceID string) *LoggregatorSender {
	return &LoggregatorSender{
		client:   client,
		sourceID: sourceID,
	}
}

func (sender *LoggregatorSender) SendValue(name string, value float64, unit string) error {
	sender.client.EmitGauge(
		loggregator.WithGaugeSourceInfo(sender.sourceID, ""),
		loggregator.WithGaugeValue(name, value, unit),
	)
	return nil
}
