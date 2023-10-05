package alert_test

import (
	"bytes"
	"errors"
	"log/slog"
	"time"

	. "github.com/cloudfoundry/replication-canary/alert"
	"github.com/cloudfoundry/replication-canary/alert/alertfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwitchboardAlerter", func() {
	var (
		testWriter            *bytes.Buffer
		alerter               *SwitchboardAlerter
		fakeSwitchboardClient *alertfakes.FakeSwitchboardClient
	)

	BeforeEach(func() {
		testWriter = new(bytes.Buffer)
		// need to set 'LevelDebug' because otherwise it uses the default of Info
		testHandler := slog.NewJSONHandler(testWriter, &slog.HandlerOptions{Level: slog.LevelDebug})
		testLogger := slog.New(testHandler)

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

				Expect(testWriter.String()).To(ContainSubstring("Switchboard alerter enabling traffic"))
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

				Expect(testWriter.String()).To(ContainSubstring("Switchboard alerter configured to no-op"))
				Expect(fakeSwitchboardClient.EnableClusterTrafficCallCount()).To(Equal(0))
				Expect(fakeSwitchboardClient.DisableClusterTrafficCallCount()).To(Equal(0))
			})
		})
	})

	Describe("Unhealthy", func() {
		It("disables traffic to the cluster", func() {
			err := alerter.Unhealthy(time.Now())
			Expect(err).NotTo(HaveOccurred())

			Expect(testWriter.String()).To(ContainSubstring("Switchboard alerter disabling traffic"))
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

				Expect(testWriter.String()).To(ContainSubstring("Switchboard alerter disabling traffic"))
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

				Expect(testWriter.String()).To(ContainSubstring("Switchboard alerter configured to no-op"))
				Expect(fakeSwitchboardClient.EnableClusterTrafficCallCount()).To(Equal(0))
				Expect(fakeSwitchboardClient.DisableClusterTrafficCallCount()).To(Equal(0))
			})
		})
	})
})
