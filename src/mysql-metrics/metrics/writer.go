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
	batch := make([]MetricDatum, 0, len(metrics))

	for i := range metrics {
		metric := metrics[i]

		if metric.Error != nil {
			writer.logger.Debug("Metric had error", map[string]interface{}{"metric": metric})
		} else {
			writer.logger.Debug("Emitted metric", map[string]interface{}{"metric": metric})
			keyWithOrigin := fmt.Sprintf("/%s/%s", writer.origin, metric.Key)
			batch = append(batch, MetricDatum{
				Name:  keyWithOrigin,
				Value: metric.Value,
				Unit:  metric.Unit,
			})
		}
	}

	if len(batch) > 0 {
		err := writer.sender.SendBatch(batch)
		if err != nil {
			writer.logger.Error("Error calling metrics sender", err)
		}
	}

	return nil
}
