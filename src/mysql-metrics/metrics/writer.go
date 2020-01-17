package metrics

import "fmt"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Logger
type Logger interface {
	Debug(string, map[string]interface{})
	Error(string, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Writer
type Writer interface {
	Write(metric []*Metric) error
}

type MetricWriter struct {
	sender Sender
	logger Logger
	origin string
}

func NewMetricWriter(sender Sender, logger Logger, origin string) *MetricWriter {
	return &MetricWriter{sender, logger, origin}
}

func (writer *MetricWriter) Write(metrics []*Metric) error {
	for i := range metrics {
		metric := metrics[i]

		if metric.Error != nil {
			writer.logger.Debug("Metric had error", map[string]interface{}{"metric": metric})
		} else {
			writer.logger.Debug("Emitted metric", map[string]interface{}{"metric": metric})
			keyWithOrigin := fmt.Sprintf("/%s/%s", writer.origin, metric.Key)
			err := writer.sender.SendValue(keyWithOrigin, metric.Value, metric.Unit)
			if err != nil {
				writer.logger.Error("Error calling metrics sender", err)
			}
		}
	}
	return nil
}
