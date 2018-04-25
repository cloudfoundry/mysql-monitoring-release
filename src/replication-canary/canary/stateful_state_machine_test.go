package canary_test

import (
	"code.cloudfoundry.org/lager/lagertest"
	. "replication-canary/canary"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatefulStateMachine", func() {
	var (
		testLogger *lagertest.TestLogger

		s   *StatefulStateMachine
		now time.Time

		onBecomesUnhealthyCallCount    int
		onBecomesNotUnhealthyCallCount int
	)

	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("stateful state machine test")

		onBecomesUnhealthyCallCount = 0
		onBecomesNotUnhealthyCallCount = 0
		s = &StatefulStateMachine{
			Logger: testLogger,
			OnBecomesUnhealthy: func(time.Time) {
				onBecomesUnhealthyCallCount++
			},
			OnBecomesNotUnhealthy: func(time.Time) {
				onBecomesNotUnhealthyCallCount++
			},
		}
		now = time.Now()
	})

	Describe("BecomesUnhealthy", func() {
		It("sets the state to unhealthy", func() {
			s.State = NotUnhealthy

			s.BecomesUnhealthy(now)

			Expect(s.State).To(Equal(Unhealthy))
		})

		It("does nothing when it is already Unhealthy", func() {
			s.State = Unhealthy

			s.BecomesUnhealthy(now)

			Expect(onBecomesUnhealthyCallCount).To(Equal(0))
			Expect(onBecomesNotUnhealthyCallCount).To(Equal(0))
		})

		It("invokes the OnBecomesUnhealthy when it starts as NotUnhealthy", func() {
			s.State = NotUnhealthy

			s.BecomesUnhealthy(now)

			Expect(onBecomesUnhealthyCallCount).To(Equal(1))
			Expect(onBecomesNotUnhealthyCallCount).To(Equal(0))
		})
	})

	Describe("BecomesNotUnhealthy", func() {
		It("sets the state to not unhealthy", func() {
			s.State = Unhealthy

			s.BecomesNotUnhealthy(now)

			Expect(s.State).To(Equal(NotUnhealthy))
		})

		It("does nothing when it is already NotUnhealthy", func() {
			s.State = NotUnhealthy

			s.BecomesNotUnhealthy(now)

			Expect(onBecomesNotUnhealthyCallCount).To(Equal(0))
			Expect(onBecomesUnhealthyCallCount).To(Equal(0))
		})

		It("invokes the OnBecomesNotUnhealthy when it starts as Unhealthy", func() {
			s.State = Unhealthy

			s.BecomesNotUnhealthy(now)

			Expect(onBecomesNotUnhealthyCallCount).To(Equal(1))
			Expect(onBecomesUnhealthyCallCount).To(Equal(0))
		})
	})

	Describe("RemainsInSameState", func() {
		Context("when the state is unhealthy", func() {
			It("does not change the state", func() {
				s.State = Unhealthy

				s.RemainsInSameState(now)

				Expect(s.State).To(Equal(Unhealthy))
				Expect(onBecomesNotUnhealthyCallCount).To(Equal(0))
				Expect(onBecomesUnhealthyCallCount).To(Equal(0))
			})
		})
		Context("when the state is not unhealthy", func() {
			It("does not change the state", func() {
				s.State = NotUnhealthy

				s.RemainsInSameState(now)

				Expect(s.State).To(Equal(NotUnhealthy))
				Expect(onBecomesNotUnhealthyCallCount).To(Equal(0))
				Expect(onBecomesUnhealthyCallCount).To(Equal(0))
			})
		})
	})

})
