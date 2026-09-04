package metrics_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/mysql-metrics/metrics"
	"github.com/cloudfoundry/mysql-metrics/metrics/metricsfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricWriter", func() {
	var (
		metricWriter *metrics.MetricWriter
		origin       = "somewhere-nice"
	)

	Describe("when the metric has an error", func() {
		It("logs an error and does not send batch", func() {
			fakeSender := new(metricsfakes.FakeSender)
			fakeLogger := new(metricsfakes.FakeLogger)
			metricWriter = metrics.NewMetricWriter(fakeSender, fakeLogger, origin)

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

			Expect(fakeSender.SendBatchCallCount()).To(Equal(0))
			Expect(fakeLogger.DebugCallCount()).To(Equal(1))
			debugMessage, debugData := fakeLogger.DebugArgsForCall(0)
			Expect(debugMessage).To(Equal("Metric had error"))
			metricFromDebug := debugData["metric"].(*metrics.Metric)
			Expect(metricFromDebug).To(Equal(metric))

			Expect(fakeLogger.ErrorCallCount()).To(Equal(0))
		})
	})

	Describe("when the metric has no error", func() {
		It("sends a batched metric request", func() {
			fakeSender := new(metricsfakes.FakeSender)
			fakeLogger := new(metricsfakes.FakeLogger)
			metricWriter = metrics.NewMetricWriter(fakeSender, fakeLogger, origin)

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

			Expect(fakeSender.SendBatchCallCount()).To(Equal(1))

			batch := fakeSender.SendBatchArgsForCall(0)
			Expect(batch).To(HaveLen(2))

			Expect(batch[0].Name).To(Equal(fmt.Sprintf("/%s/%s", origin, key1)))
			Expect(batch[0].Value).To(Equal(value1))
			Expect(batch[0].Unit).To(Equal(unit1))

			Expect(batch[1].Name).To(Equal(fmt.Sprintf("/%s/%s", origin, key2)))
			Expect(batch[1].Value).To(Equal(value2))
			Expect(batch[1].Unit).To(Equal(unit2))

			Expect(fakeLogger.DebugCallCount()).To(Equal(2))

			debugMessage, data := fakeLogger.DebugArgsForCall(0)
			Expect(debugMessage).To(ContainSubstring("Emitted metric"))
			metricFromDebug := data["metric"].(*metrics.Metric)
			Expect(metricFromDebug).To(Equal(metric1))

			debugMessage, data = fakeLogger.DebugArgsForCall(1)
			Expect(debugMessage).To(ContainSubstring("Emitted metric"))
			metricFromDebug = data["metric"].(*metrics.Metric)
			Expect(metricFromDebug).To(Equal(metric2))

			Expect(fakeLogger.ErrorCallCount()).To(Equal(0))
		})

		Describe("when the sender errors", func() {
			It("log.debug's the metric, but logs an error", func() {
				fakeSender := new(metricsfakes.FakeSender)
				fakeLogger := new(metricsfakes.FakeLogger)
				metricWriter = metrics.NewMetricWriter(fakeSender, fakeLogger, origin)

				key := "metrics-key"
				value := 123.5
				unit := "unit"
				metric := &metrics.Metric{Key: key, Value: value, Unit: unit, RawValue: "123.5000"}

				dropsondeError := errors.New("dropsonde broke somehow")
				fakeSender.SendBatchReturns(dropsondeError)

				err := metricWriter.Write([]*metrics.Metric{metric})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeLogger.DebugCallCount()).To(Equal(1))
				debugMessage, debugData := fakeLogger.DebugArgsForCall(0)
				Expect(debugMessage).To(ContainSubstring("Emitted metric"))
				metricFromDebug := debugData["metric"].(*metrics.Metric)
				Expect(metricFromDebug).To(Equal(metric))

				Expect(fakeSender.SendBatchCallCount()).To(Equal(1))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				errorMessage, errorErr := fakeLogger.ErrorArgsForCall(0)
				Expect(errorMessage).To(Equal("Error calling metrics sender"))
				Expect(errorErr).To(Equal(dropsondeError))
			})
		})
	})
})
