package alert_test

import (
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/cloudfoundry/replication-canary/alert"

	"time"

	"errors"

	"github.com/cloudfoundry/replication-canary/alert/alertfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwitchboardAlerter", func() {
	var (
		testLogger *lagertest.TestLogger

		alerter               *SwitchboardAlerter
		fakeSwitchboardClient *alertfakes.FakeSwitchboardClient
	)

	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("Switchboard alerter test")
		fakeSwitchboardClient = new(alertfakes.FakeSwitchboardClient)

		alerter = &SwitchboardAlerter{
			Logger:            testLogger,
			SwitchboardClient: fakeSwitchboardClient,
		}
	})

	Describe("NotUnhealthy", func() {
		It("enables traffic to the cluster", func() {
			err := alerter.NotUnhealthy(time.Now())
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSwitchboardClient.EnableClusterTrafficCallCount()).To(Equal(1))
		})

		Context("when enabling cluster traffic errors", func() {
			var (
				expectedError error
			)

			BeforeEach(func() {
				expectedError = errors.New("some-traffic-enabling-error")

				fakeSwitchboardClient.EnableClusterTrafficReturns(expectedError)
			})

			It("returns the error", func() {
				err := alerter.NotUnhealthy(time.Now())

				Expect(err).To(Equal(expectedError))
			})
		})

		Context("when configured to no-op", func() {
			BeforeEach(func() {
				alerter.NoOp = true
			})

			It("successfully takes no action", func() {
				err := alerter.NotUnhealthy(time.Now())
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeSwitchboardClient.EnableClusterTrafficCallCount()).To(Equal(0))
				Expect(fakeSwitchboardClient.DisableClusterTrafficCallCount()).To(Equal(0))
			})
		})
	})

	Describe("Unhealthy", func() {
		It("disables traffic to the cluster", func() {
			err := alerter.Unhealthy(time.Now())
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSwitchboardClient.DisableClusterTrafficCallCount()).To(Equal(1))
		})

		Context("when disabling cluster traffic errors", func() {
			var (
				expectedError error
			)

			BeforeEach(func() {
				expectedError = errors.New("some-traffic-disabling-error")

				fakeSwitchboardClient.DisableClusterTrafficReturns(expectedError)
			})

			It("returns the error", func() {
				err := alerter.Unhealthy(time.Now())

				Expect(err).To(Equal(expectedError))
			})
		})

		Context("when configured to no-op", func() {
			BeforeEach(func() {
				alerter.NoOp = true
			})

			It("successfully takes no action", func() {
				err := alerter.Unhealthy(time.Now())
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeSwitchboardClient.EnableClusterTrafficCallCount()).To(Equal(0))
				Expect(fakeSwitchboardClient.DisableClusterTrafficCallCount()).To(Equal(0))
			})
		})
	})
})
