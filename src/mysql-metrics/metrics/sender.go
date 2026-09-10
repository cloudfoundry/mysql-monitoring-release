package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"code.cloudfoundry.org/go-loggregator/v9"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metricNameSanitizer = strings.NewReplacer("/", "_", ".", "_", "-", "_")

// SanitizeMetricName replaces "/", ".", and "-" with "_" and removes leading slashes.
func SanitizeMetricName(name string) string {
	return metricNameSanitizer.Replace(strings.TrimPrefix(name, "/"))
}

type MetricDatum struct {
	Name  string
	Value float64
	Unit  string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Sender
type Sender interface {
	SendValue(name string, value float64, unit string) error
	SendBatch(metrics []MetricDatum) error
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
	if sender.client == nil {
		return nil
	}
	sender.client.EmitGauge(
		loggregator.WithGaugeSourceInfo(sender.sourceID, ""),
		loggregator.WithGaugeValue(name, value, unit),
	)
	return nil
}

func (sender *LoggregatorSender) SendBatch(metrics []MetricDatum) error {
	if sender.client == nil {
		return nil
	}
	for _, m := range metrics {
		_ = sender.SendValue(m.Name, m.Value, m.Unit)
	}
	return nil
}

// PrometheusSender maintains Prometheus Gauge metrics in a registry and serves them over HTTP /metrics.
type PrometheusSender struct {
	mu         sync.RWMutex
	registry   *prometheus.Registry
	gauges     map[string]prometheus.Gauge
	server     *http.Server
	sourceID   string
	origin     string
	instanceID string
	deployment string
	jobName    string
	jobIndex   string
	jobIP      string
}

func NewPrometheusSender(
	port int,
	sourceID, origin, instanceID, deployment, jobName, jobIndex, jobIP string,
	logger Logger,
) *PrometheusSender {
	registry := prometheus.NewRegistry()

	sender := &PrometheusSender{
		registry:   registry,
		gauges:     make(map[string]prometheus.Gauge),
		sourceID:   sourceID,
		origin:     origin,
		instanceID: instanceID,
		deployment: deployment,
		jobName:    jobName,
		jobIndex:   jobIndex,
		jobIP:      jobIP,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	sender.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := sender.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if logger != nil {
				logger.Error("Prometheus HTTP server error", err)
			}
		}
	}()

	return sender
}

func (s *PrometheusSender) SendValue(name string, value float64, unit string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanName := SanitizeMetricName(name)
	gauge, exists := s.gauges[cleanName]
	if !exists {
		constLabels := prometheus.Labels{
			"source_id": s.sourceID,
			"origin":    s.origin,
		}
		if s.instanceID != "" {
			constLabels["instance_id"] = s.instanceID
		}
		if s.deployment != "" {
			constLabels["deployment"] = s.deployment
		}
		if s.jobName != "" {
			constLabels["job_name"] = s.jobName
		}
		if s.jobIndex != "" {
			constLabels["job_index"] = s.jobIndex
		}
		if s.jobIP != "" {
			constLabels["job_ip"] = s.jobIP
		}

		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        cleanName,
			Help:        fmt.Sprintf("MySQL metric: %s (unit: %s)", name, unit),
			ConstLabels: constLabels,
		})
		s.registry.MustRegister(gauge)
		s.gauges[cleanName] = gauge
	}

	gauge.Set(value)
	return nil
}

func (s *PrometheusSender) SendBatch(metrics []MetricDatum) error {
	for _, m := range metrics {
		_ = s.SendValue(m.Name, m.Value, m.Unit)
	}
	return nil
}

func (s *PrometheusSender) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// MultiSender fans out emitted metrics to multiple Sender implementations.
type MultiSender struct {
	senders []Sender
}

func NewMultiSender(senders ...Sender) *MultiSender {
	return &MultiSender{senders: senders}
}

func (m *MultiSender) SendValue(name string, value float64, unit string) error {
	for _, s := range m.senders {
		_ = s.SendValue(name, value, unit)
	}
	return nil
}

func (m *MultiSender) SendBatch(metrics []MetricDatum) error {
	for _, s := range m.senders {
		_ = s.SendBatch(metrics)
	}
	return nil
}

func (m *MultiSender) Close() error {
	for _, s := range m.senders {
		if closer, ok := s.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
	return nil
}
