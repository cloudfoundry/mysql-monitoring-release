package metrics

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator/v9"
	"github.com/cloudfoundry/mysql-metrics/config"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var metricNameSanitizer = strings.NewReplacer("/", "_", ".", "_", "-", "_")

// SanitizeMetricName replaces "/", ".", and "-" with "_" and removes leading slashes.
func SanitizeMetricName(name string) string {
	return metricNameSanitizer.Replace(strings.TrimPrefix(name, "/"))
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func isOtelEndpointReachable(endpoint string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
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

type OtlpSender struct {
	mu         sync.Mutex
	conn       *grpc.ClientConn
	client     colmetricspb.MetricsServiceClient
	sourceID   string
	origin     string
	instanceID string
	deployment string
	jobName    string
	jobIndex   string
	jobIP      string
	logger     Logger
}

func NewOtlpSender(
	endpoint string,
	caPath, certPath, keyPath string,
	serverName string,
	sourceID, origin, instanceID, deployment, jobName, jobIndex, jobIP string,
	logger Logger,
) (*OtlpSender, error) {
	if endpoint == "" {
		endpoint = config.DefaultOtelEndpoint
	}
	if serverName == "" {
		serverName = config.DefaultOtelServerName
	}

	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read otel ca cert (%s): %w", caPath, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load otel client keypair (%s, %s): %w", certPath, keyPath, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", endpoint, err)
	}

	client := colmetricspb.NewMetricsServiceClient(conn)

	sender := &OtlpSender{
		conn:       conn,
		client:     client,
		sourceID:   sourceID,
		origin:     origin,
		instanceID: instanceID,
		deployment: deployment,
		jobName:    jobName,
		jobIndex:   jobIndex,
		jobIP:      jobIP,
		logger:     logger,
	}

	return sender, nil
}

// BuildExportBatchRequest constructs a batched protobuf request for OTLP metrics export.
func (s *OtlpSender) BuildExportBatchRequest(metrics []MetricDatum) *colmetricspb.ExportMetricsServiceRequest {
	now := uint64(time.Now().UnixNano())

	resourceAttrs := []*commonpb.KeyValue{
		{
			Key:   "service.name",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "mysql-metrics"}},
		},
		{
			Key:   "service.instance.id",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.instanceID}},
		},
	}

	if s.deployment != "" {
		resourceAttrs = append(resourceAttrs, &commonpb.KeyValue{
			Key:   "bosh.deployment",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.deployment}},
		})
	}
	if s.jobName != "" {
		resourceAttrs = append(resourceAttrs, &commonpb.KeyValue{
			Key:   "bosh.instance_group",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.jobName}},
		})
	}
	if s.jobIndex != "" {
		resourceAttrs = append(resourceAttrs, &commonpb.KeyValue{
			Key:   "bosh.index",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.jobIndex}},
		})
	}
	if s.jobIP != "" {
		resourceAttrs = append(resourceAttrs, &commonpb.KeyValue{
			Key:   "bosh.ip",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.jobIP}},
		})
	}

	otlpMetrics := make([]*metricspb.Metric, 0, len(metrics))
	for _, m := range metrics {
		cleanName := SanitizeMetricName(m.Name)
		otlpMetrics = append(otlpMetrics, &metricspb.Metric{
			Name: cleanName,
			Unit: m.Unit,
			Data: &metricspb.Metric_Gauge{
				Gauge: &metricspb.Gauge{
					DataPoints: []*metricspb.NumberDataPoint{
						{
							TimeUnixNano: now,
							Value:        &metricspb.NumberDataPoint_AsDouble{AsDouble: m.Value},
							Attributes: []*commonpb.KeyValue{
								{
									Key:   "source_id",
									Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.sourceID}},
								},
								{
									Key:   "origin",
									Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: s.origin}},
								},
							},
						},
					},
				},
			},
		})
	}

	return &colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: resourceAttrs,
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name:    "mysql-metrics",
							Version: "1.0.0",
						},
						Metrics: otlpMetrics,
					},
				},
			},
		},
	}
}

func (s *OtlpSender) SendBatch(metrics []MetricDatum) error {
	if len(metrics) == 0 {
		return nil
	}

	req := s.BuildExportBatchRequest(metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.client.Export(ctx, req)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("OtlpSender Export error", err)
		}
		return err
	}

	return nil
}

func (s *OtlpSender) SendValue(name string, value float64, unit string) error {
	return s.SendBatch([]MetricDatum{{Name: name, Value: value, Unit: unit}})
}

// Close gracefully closes the underlying gRPC connection.
func (s *OtlpSender) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.client = nil
		return err
	}
	return nil
}

// NewSender creates the primary metrics sender (OTel via OTLP gRPC) if OTel collector
// is reachable and certificates are present, or falls back to legacy Loggregator.
func NewSender(cfg config.Config, logger Logger) (Sender, error) {
	// Primary: Use OTel if certificates exist AND otel endpoint is listening
	if fileExists(cfg.OtelCAPath) && fileExists(cfg.OtelCertPath) && fileExists(cfg.OtelKeyPath) {
		if isOtelEndpointReachable(cfg.OtelEndpoint, 500*time.Millisecond) {
			otlpSender, err := NewOtlpSender(
				cfg.OtelEndpoint,
				cfg.OtelCAPath,
				cfg.OtelCertPath,
				cfg.OtelKeyPath,
				cfg.OtelServerName,
				cfg.SourceID,
				cfg.Origin,
				cfg.InstanceID,
				cfg.Deployment,
				cfg.JobName,
				cfg.JobIndex,
				cfg.JobIP,
				logger,
			)
			if err == nil {
				if logger != nil {
					logger.Debug("OtlpSender initialized successfully", nil)
				}
				return otlpSender, nil
			}
			if logger != nil {
				logger.Error("failed to initialize otlp sender", err)
			}
		} else {
			if logger != nil {
				logger.Debug("otel endpoint not reachable, falling back to loggregator", map[string]interface{}{
					"endpoint": cfg.OtelEndpoint,
				})
			}
		}
	}

	// Fallback: Use legacy Loggregator only if OTel is not available
	if cfg.LoggregatorCAPath != "" &&
		fileExists(cfg.LoggregatorCAPath) &&
		fileExists(cfg.LoggregatorClientCertPath) &&
		fileExists(cfg.LoggregatorClientKeyPath) {

		tlsConfig, err := loggregator.NewIngressTLSConfig(
			cfg.LoggregatorCAPath,
			cfg.LoggregatorClientCertPath,
			cfg.LoggregatorClientKeyPath,
		)
		if err != nil {
			if logger != nil {
				logger.Error("failed to create loggregator tls config", err)
			}
			return nil, fmt.Errorf("failed to create loggregator tls config: %w", err)
		}

		ingressClient, err := loggregator.NewIngressClient(
			tlsConfig,
			loggregator.WithAddr("localhost:3458"),
			loggregator.WithTag("source_id", cfg.SourceID),
			loggregator.WithTag("origin", cfg.Origin),
		)
		if err != nil {
			if logger != nil {
				logger.Error("failed to create loggregator client", err)
			}
			return nil, fmt.Errorf("failed to create loggregator client: %w", err)
		}

		if logger != nil {
			logger.Debug("LoggregatorSender initialized successfully (fallback)", nil)
		}
		return NewLoggregatorSender(ingressClient, cfg.SourceID), nil
	}

	return nil, errors.New("no metrics sender could be initialized (neither otel nor loggregator credentials found)")
}
