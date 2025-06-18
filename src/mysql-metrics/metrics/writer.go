package metrics

import (
	"fmt"
	"log/slog"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Writer
type Writer interface {
	Write(metric []*Metric) error
}

type MetricWriter struct {
	sender Sender
	origin string
}

func NewMetricWriter(sender Sender, origin string) *MetricWriter {
	return &MetricWriter{sender, origin}
}

func (writer *MetricWriter) Write(metrics []*Metric) error {
	for i := range metrics {
		metric := metrics[i]

		if metric.Error != nil {
			slog.Debug("Metric had an error", "metric", metric, "error", metric.Error)
		} else {
			slog.Debug("Emitted metric", slog.Group("data", "metric", metric))
			keyWithOrigin := fmt.Sprintf("/%s/%s", writer.origin, metric.Key)
			err := writer.sender.SendValue(keyWithOrigin, metric.Value, metric.Unit)
			if err != nil {
				slog.Error("Error calling metrics sender", "error", err)
			}
		}
	}
	return nil
}
