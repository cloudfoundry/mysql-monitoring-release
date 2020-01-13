package metrics_test

import (
	"errors"

	"fmt"
	"github.com/cloudfoundry/mysql-metrics/metrics"
	"github.com/cloudfoundry/mysql-metrics/metrics/metricsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricWriter", func() {
	var (
		metricWriter *metrics.MetricWriter
		origin       = "somewhere-nice"
	)

	Describe("when the metric has an error", func() {
		It("logs an error", func() {
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

			Expect(fakeSender.SendValueCallCount()).To(Equal(0))
			Expect(fakeLogger.DebugCallCount()).To(Equal(1))
			debugMessage, debugData := fakeLogger.DebugArgsForCall(0)
			Expect(debugMessage).To(Equal("Metric had error"))
			metricFromDebug := debugData["metric"].(*metrics.Metric)
			Expect(metricFromDebug).To(Equal(metric))

			Expect(fakeLogger.ErrorCallCount()).To(Equal(0))
		})
	})

	Describe("when the metric has no error", func() {
		It("sends a metric ", func() {
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

			Expect(fakeSender.SendValueCallCount()).To(Equal(2))

			keyArg, valueArg, unitArg := fakeSender.SendValueArgsForCall(0)
			Expect(keyArg).To(Equal(fmt.Sprintf("/%s/%s", origin, key1)))
			Expect(valueArg).To(Equal(value1))
			Expect(unitArg).To(Equal(unit1))

			keyArg, valueArg, unitArg = fakeSender.SendValueArgsForCall(1)
			Expect(keyArg).To(Equal(fmt.Sprintf("/%s/%s", origin, key2)))
			Expect(valueArg).To(Equal(value2))
			Expect(unitArg).To(Equal(unit2))

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
				fakeSender.SendValueReturns(dropsondeError)

				err := metricWriter.Write([]*metrics.Metric{metric})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeLogger.DebugCallCount()).To(Equal(1))
				debugMessage, debugData := fakeLogger.DebugArgsForCall(0)
				Expect(debugMessage).To(ContainSubstring("Emitted metric"))
				metricFromDebug := debugData["metric"].(*metrics.Metric)
				Expect(metricFromDebug).To(Equal(metric))

				Expect(fakeSender.SendValueCallCount()).To(Equal(1))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				errorMessage, errorErr := fakeLogger.ErrorArgsForCall(0)
				Expect(errorMessage).To(Equal("Error calling metrics sender"))
				Expect(errorErr).To(Equal(dropsondeError))
			})
		})
	})
})
