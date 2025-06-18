package metrics_test

import (
	"errors"
	"fmt"
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/mysql-metrics/metrics"
	"github.com/cloudfoundry/mysql-metrics/metrics/metricsfakes"
)

var _ = Describe("MetricWriter", func() {
	var (
		metricWriter *metrics.MetricWriter
		origin       = "somewhere-nice"
		logBuffer    *gbytes.Buffer
	)

	BeforeEach(func() {
		// Set up global slog to write to a buffer for test assertions
		logBuffer = gbytes.NewBuffer()
		slog.SetDefault(slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})))
	})

	Describe("when the metric has an error", func() {
		It("logs an error", func() {
			fakeSender := new(metricsfakes.FakeSender)
			metricWriter = metrics.NewMetricWriter(fakeSender, origin)

			key := "metrics-key"
			value := 0.0
			unit := "unit"
			metricError := errors.New("something busted")
			metric := &metrics.Metric{
				Key:      key,
				Value:    value,
				Unit:     unit,
				RawValue: "cannotconvert",
				Error:    metricError,
			}

			err := metricWriter.Write([]*metrics.Metric{metric})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSender.SendValueCallCount()).To(Equal(0))
			Eventually(logBuffer).Should(gbytes.Say(`"level":"DEBUG"`))
			Eventually(logBuffer).Should(gbytes.Say(`"msg":"Metric had an error"`))
			Eventually(logBuffer).Should(gbytes.Say(`"error":"something busted"`))
			Consistently(logBuffer).ShouldNot(gbytes.Say(`"level":"ERROR"`))
		})
	})

	Describe("when the metric has no error", func() {
		It("sends a metric ", func() {
			fakeSender := new(metricsfakes.FakeSender)
			metricWriter = metrics.NewMetricWriter(fakeSender, origin)

			key1 := "metrics-key1"
			value1 := 123.5
			unit1 := "unit1"
			metric1 := &metrics.Metric{Key: key1, Value: value1, Unit: unit1, RawValue: "123.5000"}

			key2 := "metrics-key2"
			value2 := 876.3
			unit2 := "unit2"
			metric2 := &metrics.Metric{Key: key2, Value: value2, Unit: unit2, RawValue: "876.3000"}

			err := metricWriter.Write([]*metrics.Metric{metric1, metric2})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSender.SendValueCallCount()).To(Equal(2))

			keyArg, valueArg, unitArg := fakeSender.SendValueArgsForCall(0)
			Expect(keyArg).To(Equal(fmt.Sprintf("/%s/%s", origin, key1)))
			Expect(valueArg).To(Equal(value1))
			Expect(unitArg).To(Equal(unit1))

			keyArg, valueArg, unitArg = fakeSender.SendValueArgsForCall(1)
			Expect(keyArg).To(Equal(fmt.Sprintf("/%s/%s", origin, key2)))
			Expect(valueArg).To(Equal(value2))
			Expect(unitArg).To(Equal(unit2))

			Eventually(logBuffer).Should(gbytes.Say(`"level":"DEBUG"`))
			Eventually(logBuffer).Should(gbytes.Say(`"msg":"Emitted metric"`))
			Consistently(logBuffer).ShouldNot(gbytes.Say(`"level":"ERROR"`))
		})

		Describe("when the sender errors", func() {
			It("log.debug's the metric, but logs an error", func() {
				fakeSender := new(metricsfakes.FakeSender)
				metricWriter = metrics.NewMetricWriter(fakeSender, origin)

				key := "metrics-key"
				value := 123.5
				unit := "unit"
				metric := &metrics.Metric{Key: key, Value: value, Unit: unit, RawValue: "123.5000"}

				dropsondeError := errors.New("dropsonde broke somehow")
				fakeSender.SendValueReturns(dropsondeError)

				err := metricWriter.Write([]*metrics.Metric{metric})
				Expect(err).NotTo(HaveOccurred())

				Eventually(logBuffer).Should(gbytes.Say(`"level":"DEBUG"`))
				Eventually(logBuffer).Should(gbytes.Say(`"msg":"Emitted metric"`))
				Expect(fakeSender.SendValueCallCount()).To(Equal(1))
				Eventually(logBuffer).Should(gbytes.Say(`"level":"ERROR"`))
				Eventually(logBuffer).Should(gbytes.Say(`"msg":"Error calling metrics sender"`))
				Eventually(logBuffer).Should(gbytes.Say("dropsonde broke somehow"))
			})
		})
	})
})
