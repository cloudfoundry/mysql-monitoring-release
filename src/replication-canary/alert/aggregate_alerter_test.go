package alert_test

import (
	"errors"

	. "github.com/cloudfoundry/replication-canary/alert"
	"github.com/cloudfoundry/replication-canary/alert/alertfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AggregateAlerter", func() {
	var (
		fakeAlerter1 *alertfakes.FakeAlerter
		fakeAlerter2 *alertfakes.FakeAlerter
		alerter      AggregateAlerter
	)

	BeforeEach(func() {
		fakeAlerter1 = new(alertfakes.FakeAlerter)
		fakeAlerter2 = new(alertfakes.FakeAlerter)

		alerter = AggregateAlerter{
			fakeAlerter1,
			fakeAlerter2,
		}
	})

	Describe("NotUnhealthy", func() {
		BeforeEach(func() {
			fakeAlerter1.NotUnhealthyReturns(nil)
			fakeAlerter2.NotUnhealthyReturns(nil)
		})

		It("forwards the call to everything it contains", func() {
			err := alerter.NotUnhealthy()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAlerter1.NotUnhealthyCallCount()).To(Equal(1))
			Expect(fakeAlerter2.NotUnhealthyCallCount()).To(Equal(1))
		})

		Context("when the left-most one fails", func() {
			var (
				alerter1Error error
			)

			BeforeEach(func() {
				alerter1Error = errors.New("some-error-in-alerter-1")
				fakeAlerter1.NotUnhealthyReturns(alerter1Error)
			})

			It("still forwards the call to everything it contains", func() {
				err := alerter.NotUnhealthy()
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), Not(ContainSubstring("some-error-in-alerter-2"))))

				Expect(fakeAlerter1.NotUnhealthyCallCount()).To(Equal(1))
				Expect(fakeAlerter2.NotUnhealthyCallCount()).To(Equal(1))
			})
		})

		Context("when multiple fail", func() {
			var (
				alerter1Error error
				alerter2Error error
			)

			BeforeEach(func() {
				alerter1Error = errors.New("some-error-in-alerter-1")
				fakeAlerter1.NotUnhealthyReturns(alerter1Error)
				alerter2Error = errors.New("some-error-in-alerter-2")
				fakeAlerter2.NotUnhealthyReturns(alerter2Error)
			})

			It("returns a multi-error of them", func() {
				err := alerter.NotUnhealthy()
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), ContainSubstring("some-error-in-alerter-2")))

				Expect(fakeAlerter1.NotUnhealthyCallCount()).To(Equal(1))
				Expect(fakeAlerter2.NotUnhealthyCallCount()).To(Equal(1))
			})
		})
	})

	Describe("Unhealthy", func() {
		BeforeEach(func() {
			fakeAlerter1.UnhealthyReturns(nil)
			fakeAlerter2.UnhealthyReturns(nil)
		})

		It("forwards the call to everything it contains", func() {
			err := alerter.Unhealthy()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAlerter1.UnhealthyCallCount()).To(Equal(1))
			Expect(fakeAlerter2.UnhealthyCallCount()).To(Equal(1))
		})

		Context("when the left-most one fails", func() {
			var (
				alerter1Error error
			)

			BeforeEach(func() {
				alerter1Error = errors.New("some-error-in-alerter-1")
				fakeAlerter1.UnhealthyReturns(alerter1Error)
			})

			It("still forwards the call to everything it contains", func() {
				err := alerter.Unhealthy()
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), Not(ContainSubstring("some-error-in-alerter-2"))))

				Expect(fakeAlerter1.UnhealthyCallCount()).To(Equal(1))
				Expect(fakeAlerter2.UnhealthyCallCount()).To(Equal(1))
			})
		})

		Context("when multiple fail", func() {
			var (
				alerter1Error error
				alerter2Error error
			)

			BeforeEach(func() {
				alerter1Error = errors.New("some-error-in-alerter-1")
				fakeAlerter1.UnhealthyReturns(alerter1Error)
				alerter2Error = errors.New("some-error-in-alerter-2")
				fakeAlerter2.UnhealthyReturns(alerter2Error)
			})

			It("returns a multi-error of them", func() {
				err := alerter.Unhealthy()
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), ContainSubstring("some-error-in-alerter-2")))

				Expect(fakeAlerter1.UnhealthyCallCount()).To(Equal(1))
				Expect(fakeAlerter2.UnhealthyCallCount()).To(Equal(1))
			})
		})
	})
})
