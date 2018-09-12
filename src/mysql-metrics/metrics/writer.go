package metrics

import "fmt"

//go:generate counterfeiter . Logger
type Logger interface {
	Debug(string, map[string]interface{})
	Error(string, error)
}

//go:generate counterfeiter . Writer
type Writer interface {
	Write(metric []*Metric) error
}

type MetricWriter struct {
	sender Sender
	logger Logger
	sourceId string
}

func NewMetricWriter(sender Sender, logger Logger, sourceId string) *MetricWriter {
	return &MetricWriter{sender, logger, sourceId}
}

func (writer *MetricWriter) Write(metrics []*Metric) error {
	for i := range metrics {
		metric := metrics[i]

		if metric.Error != nil {
			writer.logger.Debug("Metric had error", map[string]interface{}{"metric": metric})
		} else {
			writer.logger.Debug("Emitted metric", map[string]interface{}{"metric": metric})
			keyWithOrigin := fmt.Sprintf("/%s/%s", writer.sourceId, metric.Key)
			err := writer.sender.SendValue(keyWithOrigin, metric.Value, metric.Unit)
			if err != nil {
				writer.logger.Error("Error calling metrics sender", err)
			}
		}
	}
	return nil
}
