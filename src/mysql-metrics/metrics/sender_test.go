package metrics_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/mysql-metrics/config"
	"github.com/cloudfoundry/mysql-metrics/metrics"
	"github.com/cloudfoundry/mysql-metrics/metrics/metricsfakes"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func generateTestCertificates(dir string) (caPath, certPath, keyPath string) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).NotTo(HaveOccurred())

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "otel-collector",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDer, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	Expect(err).NotTo(HaveOccurred())

	caPath = filepath.Join(dir, "ca.crt")
	certPath = filepath.Join(dir, "client.crt")
	keyPath = filepath.Join(dir, "client.key")

	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
	err = os.WriteFile(caPath, certPem, 0600)
	Expect(err).NotTo(HaveOccurred())
	err = os.WriteFile(certPath, certPem, 0600)
	Expect(err).NotTo(HaveOccurred())

	keyDer, err := x509.MarshalECPrivateKey(privKey)
	Expect(err).NotTo(HaveOccurred())
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDer})
	err = os.WriteFile(keyPath, keyPem, 0600)
	Expect(err).NotTo(HaveOccurred())

	return caPath, certPath, keyPath
}

var _ = Describe("Senders", func() {
	var fakeLogger *metricsfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = new(metricsfakes.FakeLogger)
	})

	Describe("LoggregatorSender", func() {
		It("handles nil client gracefully on SendValue", func() {
			sender := metrics.NewLoggregatorSender(nil, "p-mysql")
			err := sender.SendValue("test_metric", 1.0, "gauge")
			Expect(err).NotTo(HaveOccurred())
		})

		It("handles nil client gracefully on SendBatch", func() {
			sender := metrics.NewLoggregatorSender(nil, "p-mysql")
			err := sender.SendBatch([]metrics.MetricDatum{
				{Name: "test_metric1", Value: 1.0, Unit: "gauge"},
				{Name: "test_metric2", Value: 2.0, Unit: "count"},
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Metric Name Sanitization", func() {
		It("replaces slashes, dots, and hyphens with underscores and trims leading slash", func() {
			Expect(metrics.SanitizeMetricName("/p-mysql/galera/wsrep_cluster_size")).To(Equal("p_mysql_galera_wsrep_cluster_size"))
			Expect(metrics.SanitizeMetricName("cpu.idle-time")).To(Equal("cpu_idle_time"))
			Expect(metrics.SanitizeMetricName("simple_metric")).To(Equal("simple_metric"))
		})
	})

	Describe("OtlpSender", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "otlp-sender-test")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		Describe("initialization", func() {
			It("returns error when cert files do not exist", func() {
				sender, err := metrics.NewOtlpSender(
					"127.0.0.1:9564",
					"/nonexistent/ca.crt",
					"/nonexistent/client.crt",
					"/nonexistent/client.key",
					"otel-collector",
					"p-mysql",
					"p-mysql",
					"inst-1",
					"service-instance-xyz",
					"mysql",
					"0",
					"10.0.0.1",
					fakeLogger,
				)
				Expect(err).To(HaveOccurred())
				Expect(sender).To(BeNil())
			})

			It("initializes successfully with valid certificate files", func() {
				caPath, certPath, keyPath := generateTestCertificates(tempDir)
				sender, err := metrics.NewOtlpSender(
					"127.0.0.1:9564",
					caPath,
					certPath,
					keyPath,
					"otel-collector",
					"p-mysql",
					"p-mysql",
					"inst-1",
					"service-instance-xyz",
					"mysql",
					"0",
					"10.0.0.1",
					fakeLogger,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(sender).NotTo(BeNil())
				Expect(sender.Close()).To(Succeed())
			})
		})

		Describe("BuildExportBatchRequest", func() {
			It("builds a well-formed batched OTLP metric request with BOSH attributes and multiple metrics", func() {
				caPath, certPath, keyPath := generateTestCertificates(tempDir)
				sender, err := metrics.NewOtlpSender(
					"127.0.0.1:9564",
					caPath,
					certPath,
					keyPath,
					"otel-collector",
					"my-source-id",
					"my-origin",
					"my-instance-id",
					"my-deployment",
					"mysql",
					"0",
					"192.168.1.5",
					fakeLogger,
				)
				Expect(err).NotTo(HaveOccurred())
				defer sender.Close()

				batch := []metrics.MetricDatum{
					{Name: "/p-mysql/galera/wsrep_cluster_size", Value: 3.0, Unit: "count"},
					{Name: "/p-mysql/cpu/idle_time", Value: 95.5, Unit: "percent"},
				}

				req := sender.BuildExportBatchRequest(batch)
				Expect(req).NotTo(BeNil())
				Expect(req.ResourceMetrics).To(HaveLen(1))

				rm := req.ResourceMetrics[0]
				Expect(rm.Resource).NotTo(BeNil())

				// Check Resource Attributes
				resAttrs := make(map[string]string)
				for _, attr := range rm.Resource.Attributes {
					resAttrs[attr.Key] = attr.Value.GetStringValue()
				}
				Expect(resAttrs["service.name"]).To(Equal("mysql-metrics"))
				Expect(resAttrs["service.instance.id"]).To(Equal("my-instance-id"))
				Expect(resAttrs["bosh.deployment"]).To(Equal("my-deployment"))
				Expect(resAttrs["bosh.instance_group"]).To(Equal("mysql"))
				Expect(resAttrs["bosh.index"]).To(Equal("0"))
				Expect(resAttrs["bosh.ip"]).To(Equal("192.168.1.5"))

				// Check Scope Metrics
				Expect(rm.ScopeMetrics).To(HaveLen(1))
				sm := rm.ScopeMetrics[0]
				Expect(sm.Scope.Name).To(Equal("mysql-metrics"))
				Expect(sm.Scope.Version).To(Equal("1.0.0"))

				// Check Metrics in batch
				Expect(sm.Metrics).To(HaveLen(2))

				m1 := sm.Metrics[0]
				Expect(m1.Name).To(Equal("p_mysql_galera_wsrep_cluster_size"))
				Expect(m1.Unit).To(Equal("count"))
				gauge1 := m1.Data.(*metricspb.Metric_Gauge).Gauge
				Expect(gauge1.DataPoints).To(HaveLen(1))
				Expect(gauge1.DataPoints[0].Value.(*metricspb.NumberDataPoint_AsDouble).AsDouble).To(Equal(3.0))

				m2 := sm.Metrics[1]
				Expect(m2.Name).To(Equal("p_mysql_cpu_idle_time"))
				Expect(m2.Unit).To(Equal("percent"))
				gauge2 := m2.Data.(*metricspb.Metric_Gauge).Gauge
				Expect(gauge2.DataPoints).To(HaveLen(1))
				Expect(gauge2.DataPoints[0].Value.(*metricspb.NumberDataPoint_AsDouble).AsDouble).To(Equal(95.5))
			})
		})
	})

	Describe("NewSender factory", func() {
		var (
			tempDir string
			cfg     config.Config
		)

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "sender-factory-test")
			Expect(err).NotTo(HaveOccurred())

			cfg = config.Config{
				SourceID:   "p-mysql",
				Origin:     "p-mysql",
				InstanceID: "inst-1",
				Deployment: "my-deployment",
				JobName:    "mysql",
				JobIndex:   "0",
				JobIP:      "192.168.1.5",
			}
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		It("creates an OtlpSender when OTel certificates exist and endpoint is reachable", func() {
			caPath, certPath, keyPath := generateTestCertificates(tempDir)
			cfg.OtelCAPath = caPath
			cfg.OtelCertPath = certPath
			cfg.OtelKeyPath = keyPath

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())
			defer ln.Close()
			cfg.OtelEndpoint = ln.Addr().String()

			sender, err := metrics.NewSender(cfg, fakeLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(sender).NotTo(BeNil())
			_, isOtlp := sender.(*metrics.OtlpSender)
			Expect(isOtlp).To(BeTrue())
		})

		It("falls back to LoggregatorSender when OTel endpoint is unreachable", func() {
			caPath, certPath, keyPath := generateTestCertificates(tempDir)
			cfg.OtelCAPath = caPath
			cfg.OtelCertPath = certPath
			cfg.OtelKeyPath = keyPath
			// Pick a non-listening port for OTel
			cfg.OtelEndpoint = "127.0.0.1:59999"

			cfg.LoggregatorCAPath = caPath
			cfg.LoggregatorClientCertPath = certPath
			cfg.LoggregatorClientKeyPath = keyPath

			sender, err := metrics.NewSender(cfg, fakeLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(sender).NotTo(BeNil())
			_, isLoggregator := sender.(*metrics.LoggregatorSender)
			Expect(isLoggregator).To(BeTrue())
		})

		It("falls back to LoggregatorSender when OTel certs are absent but Loggregator certs exist", func() {
			caPath, certPath, keyPath := generateTestCertificates(tempDir)
			cfg.OtelCAPath = filepath.Join(tempDir, "nonexistent-otel-ca.crt")
			cfg.OtelCertPath = filepath.Join(tempDir, "nonexistent-otel-cert.crt")
			cfg.OtelKeyPath = filepath.Join(tempDir, "nonexistent-otel-key.key")

			cfg.LoggregatorCAPath = caPath
			cfg.LoggregatorClientCertPath = certPath
			cfg.LoggregatorClientKeyPath = keyPath

			sender, err := metrics.NewSender(cfg, fakeLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(sender).NotTo(BeNil())
			_, isLoggregator := sender.(*metrics.LoggregatorSender)
			Expect(isLoggregator).To(BeTrue())
		})

		It("returns an error when neither OTel nor Loggregator credentials exist", func() {
			cfg.OtelCAPath = filepath.Join(tempDir, "nonexistent-otel-ca.crt")
			cfg.OtelCertPath = filepath.Join(tempDir, "nonexistent-otel-cert.crt")
			cfg.OtelKeyPath = filepath.Join(tempDir, "nonexistent-otel-key.key")
			cfg.LoggregatorCAPath = filepath.Join(tempDir, "nonexistent-logg-ca.crt")
			cfg.LoggregatorClientCertPath = filepath.Join(tempDir, "nonexistent-logg-cert.crt")
			cfg.LoggregatorClientKeyPath = filepath.Join(tempDir, "nonexistent-logg-key.key")

			sender, err := metrics.NewSender(cfg, fakeLogger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no metrics sender could be initialized"))
			Expect(sender).To(BeNil())
		})
	})
})
