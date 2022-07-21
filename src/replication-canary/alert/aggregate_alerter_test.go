package alert_test

import (
	. "github.com/cloudfoundry/replication-canary/alert"

	"errors"
	"time"

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
		var (
			now time.Time
		)

		BeforeEach(func() {
			now = time.Now()

			fakeAlerter1.NotUnhealthyReturns(nil)
			fakeAlerter2.NotUnhealthyReturns(nil)
		})

		It("forwards the call to everything it contains", func() {
			err := alerter.NotUnhealthy(now)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAlerter1.NotUnhealthyCallCount()).To(Equal(1))
			t1 := fakeAlerter1.NotUnhealthyArgsForCall(0)
			Expect(t1).To(Equal(now))

			Expect(fakeAlerter2.NotUnhealthyCallCount()).To(Equal(1))
			t2 := fakeAlerter2.NotUnhealthyArgsForCall(0)
			Expect(t2).To(Equal(now))
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
				err := alerter.NotUnhealthy(now)
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), Not(ContainSubstring("some-error-in-alerter-2"))))

				Expect(fakeAlerter1.NotUnhealthyCallCount()).To(Equal(1))
				t1 := fakeAlerter1.NotUnhealthyArgsForCall(0)
				Expect(t1).To(Equal(now))

				Expect(fakeAlerter2.NotUnhealthyCallCount()).To(Equal(1))
				t2 := fakeAlerter2.NotUnhealthyArgsForCall(0)
				Expect(t2).To(Equal(now))
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
				err := alerter.NotUnhealthy(now)
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), ContainSubstring("some-error-in-alerter-2")))

				Expect(fakeAlerter1.NotUnhealthyCallCount()).To(Equal(1))
				t1 := fakeAlerter1.NotUnhealthyArgsForCall(0)
				Expect(t1).To(Equal(now))

				Expect(fakeAlerter2.NotUnhealthyCallCount()).To(Equal(1))
				t2 := fakeAlerter2.NotUnhealthyArgsForCall(0)
				Expect(t2).To(Equal(now))
			})
		})
	})

	Describe("Unhealthy", func() {
		var (
			now time.Time
		)

		BeforeEach(func() {
			now = time.Now()

			fakeAlerter1.UnhealthyReturns(nil)
			fakeAlerter2.UnhealthyReturns(nil)
		})

		It("forwards the call to everything it contains", func() {
			err := alerter.Unhealthy(now)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAlerter1.UnhealthyCallCount()).To(Equal(1))
			t1 := fakeAlerter1.UnhealthyArgsForCall(0)
			Expect(t1).To(Equal(now))

			Expect(fakeAlerter2.UnhealthyCallCount()).To(Equal(1))
			t2 := fakeAlerter2.UnhealthyArgsForCall(0)
			Expect(t2).To(Equal(now))
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
				err := alerter.Unhealthy(now)
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), Not(ContainSubstring("some-error-in-alerter-2"))))

				Expect(fakeAlerter1.UnhealthyCallCount()).To(Equal(1))
				t1 := fakeAlerter1.UnhealthyArgsForCall(0)
				Expect(t1).To(Equal(now))

				Expect(fakeAlerter2.UnhealthyCallCount()).To(Equal(1))
				t2 := fakeAlerter2.UnhealthyArgsForCall(0)
				Expect(t2).To(Equal(now))
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
				err := alerter.Unhealthy(now)
				Expect(err.Error()).To(And(ContainSubstring("some-error-in-alerter-1"), ContainSubstring("some-error-in-alerter-2")))

				Expect(fakeAlerter1.UnhealthyCallCount()).To(Equal(1))
				t1 := fakeAlerter1.UnhealthyArgsForCall(0)
				Expect(t1).To(Equal(now))

				Expect(fakeAlerter2.UnhealthyCallCount()).To(Equal(1))
				t2 := fakeAlerter2.UnhealthyArgsForCall(0)
				Expect(t2).To(Equal(now))
			})
		})
	})
})
