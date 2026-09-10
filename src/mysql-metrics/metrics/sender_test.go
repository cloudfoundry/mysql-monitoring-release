package metrics_test

import (
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-metrics/metrics"
)

var _ = Describe("PrometheusSender", func() {
	var (
		sender *metrics.PrometheusSender
		port   int
	)

	BeforeEach(func() {
		port = 19828
		sender = metrics.NewPrometheusSender(
			port,
			"test-source-id",
			"test-origin",
			"inst-123",
			"test-deployment",
			"mysql",
			"0",
			"10.0.0.1",
			nil,
		)
		time.Sleep(50 * time.Millisecond)
	})

	AfterEach(func() {
		if sender != nil {
			_ = sender.Close()
		}
	})

	It("exposes registered metrics on HTTP /metrics endpoint with labels", func() {
		err := sender.SendValue("/mysql/threads_connected", 42.0, "Count")
		Expect(err).NotTo(HaveOccurred())

		err = sender.SendValue("/system/cpu_utilization", 15.5, "Percentage")
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())

			bodyStr := string(body)
			g.Expect(bodyStr).To(ContainSubstring("mysql_threads_connected{deployment=\"test-deployment\",instance_id=\"inst-123\",job_index=\"0\",job_ip=\"10.0.0.1\",job_name=\"mysql\",origin=\"test-origin\",source_id=\"test-source-id\"} 42"))
			g.Expect(bodyStr).To(ContainSubstring("system_cpu_utilization{deployment=\"test-deployment\",instance_id=\"inst-123\",job_index=\"0\",job_ip=\"10.0.0.1\",job_name=\"mysql\",origin=\"test-origin\",source_id=\"test-source-id\"} 15.5"))
		}, "2s", "50ms").Should(Succeed())
	})

	It("handles batch sending of metrics", func() {
		batch := []metrics.MetricDatum{
			{Name: "mysql/connections", Value: 100, Unit: "Count"},
			{Name: "mysql/queries", Value: 5000, Unit: "Count"},
		}
		err := sender.SendBatch(batch)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())

			bodyStr := string(body)
			g.Expect(bodyStr).To(ContainSubstring("mysql_connections"))
			g.Expect(bodyStr).To(ContainSubstring("mysql_queries"))
		}, "2s", "50ms").Should(Succeed())
	})
})

var _ = Describe("MultiSender", func() {
	It("fans out metric sending to all underlying senders", func() {
		promSender := metrics.NewPrometheusSender(19829, "src", "orig", "id", "dep", "job", "0", "ip", nil)
		defer promSender.Close()

		multiSender := metrics.NewMultiSender(promSender)
		err := multiSender.SendValue("test_metric", 99.0, "Gauge")
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			resp, err := http.Get("http://127.0.0.1:19829/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(string(body)).To(ContainSubstring("test_metric"))
		}, "2s", "50ms").Should(Succeed())
	})
})

var _ = Describe("SanitizeMetricName", func() {
	It("replaces slashes, dots, and hyphens with underscores", func() {
		Expect(metrics.SanitizeMetricName("/mysql/galera-status.wsrep_ready")).To(Equal("mysql_galera_status_wsrep_ready"))
		Expect(metrics.SanitizeMetricName("cpu.idle-time")).To(Equal("cpu_idle_time"))
	})
})
