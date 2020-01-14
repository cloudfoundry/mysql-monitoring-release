package canary_test

import (
	. "github.com/cloudfoundry/replication-canary/canary"

	"errors"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/replication-canary/canary/canaryfakes"
)

var _ = Describe("CoalMiner", func() {
	var (
		miner                 *CoalMiner
		fakeConnectionFactory *canaryfakes.FakeConnectionFactory
		fakeChirper           *canaryfakes.FakeChirper
		fakeAlerter           *canaryfakes.FakeAlerter
		fakeStateMachine      *canaryfakes.FakeStateMachine
		fakeLogger            *lagertest.TestLogger

		t   time.Time
		err error
	)

	BeforeEach(func() {
		fakeConnectionFactory = new(canaryfakes.FakeConnectionFactory)
		fakeAlerter = new(canaryfakes.FakeAlerter)
		fakeChirper = new(canaryfakes.FakeChirper)
		fakeStateMachine = new(canaryfakes.FakeStateMachine)
		fakeLogger = lagertest.NewTestLogger("coal-miner-test")

		const longForm = "Jan 2, 2006 at 3:04pm (MST)"
		t, err = time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("NewCoalMiner", func() {
		var (
			m StateMachine
		)

		BeforeEach(func() {
			miner = NewCoalMiner(fakeConnectionFactory, fakeChirper, fakeAlerter, fakeLogger)
			m = miner.StateMachine
		})

		It("creates a StatefulStateMachine", func() {
			_, ok := (m).(*StatefulStateMachine)
			Expect(ok).To(BeTrue())
		})

		It("has an OnBecomesUnhealthy that alerts that it's unhealthy", func() {
			sm, _ := (m).(*StatefulStateMachine)
			sm.OnBecomesUnhealthy(t)

			Expect(fakeAlerter.UnhealthyCallCount()).To(Equal(1))
			Expect(fakeAlerter.NotUnhealthyCallCount()).To(Equal(0))
		})

		It("has an OnBecomesNotUnhealthy that alerts that it's not unhealthy", func() {
			sm, _ := (m).(*StatefulStateMachine)
			sm.OnBecomesNotUnhealthy(t)

			Expect(fakeAlerter.NotUnhealthyCallCount()).To(Equal(1))
			Expect(fakeAlerter.UnhealthyCallCount()).To(Equal(0))
		})
	})

	Describe("LetFly", func() {
		var (
			timer chan (time.Time)
		)

		BeforeEach(func() {
			timer = make(chan (time.Time))

			miner = NewCoalMiner(
				fakeConnectionFactory,
				fakeChirper,
				fakeAlerter,
				fakeLogger,
			)
			miner.StateMachine = fakeStateMachine

			fakeChirper.ChirpReturns(true, nil)
		})

		AfterEach(func() {
			close(timer)
		})

		It("chirps the canary", func(done Done) {
			go miner.LetSing(timer)
			timer <- t

			Eventually(fakeChirper.ChirpCallCount).Should(Equal(1))

			close(done)
		})

		Context("when the canary fails to obtain a write connection", func() {
			BeforeEach(func() {
				fakeConnectionFactory.WriteConnReturns(nil, errors.New("write conn error"))
			})

			It("changes the state to NotUnhealthy", func(done Done) {
				go miner.LetSing(timer)
				timer <- t

				Eventually(fakeStateMachine.RemainsInSameStateCallCount).Should(Equal(1))
				Consistently(fakeStateMachine.BecomesNotUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeStateMachine.BecomesUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeChirper.ChirpCallCount()).Should(Equal(0))

				close(done)
			})
		})

		Context("when the canary fails to obtain conns", func() {
			BeforeEach(func() {
				fakeConnectionFactory.ConnsReturns(nil, errors.New("conns error"))
			})

			It("changes the state to NotUnhealthy", func(done Done) {
				go miner.LetSing(timer)
				timer <- t

				Eventually(fakeStateMachine.RemainsInSameStateCallCount).Should(Equal(1))
				Consistently(fakeStateMachine.BecomesNotUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeStateMachine.BecomesUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeChirper.ChirpCallCount()).Should(Equal(0))

				close(done)
			})
		})

		Context("when the canary returns (false,err)", func() {
			BeforeEach(func() {
				fakeChirper.ChirpReturns(false, errors.New("some error"))
			})

			It("changes the state to NotUnhealthy", func(done Done) {
				go miner.LetSing(timer)
				timer <- t

				Eventually(fakeStateMachine.BecomesNotUnhealthyCallCount).Should(Equal(1))
				Consistently(fakeStateMachine.BecomesUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeStateMachine.RemainsInSameStateCallCount).Should(Equal(0))

				close(done)
			})
		})

		Context("when the canary returns (true,nil)", func() {
			BeforeEach(func() {
				fakeChirper.ChirpReturns(true, nil)
			})

			It("changes the state to NotUnhealthy", func(done Done) {
				go miner.LetSing(timer)
				timer <- t

				Eventually(fakeStateMachine.BecomesNotUnhealthyCallCount).Should(Equal(1))
				Consistently(fakeStateMachine.BecomesUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeStateMachine.RemainsInSameStateCallCount).Should(Equal(0))

				close(done)
			})
		})

		Context("when the canary returns (false,nil)", func() {
			BeforeEach(func() {
				fakeChirper.ChirpReturns(false, nil)
			})

			It("changes the state to Unhealthy", func(done Done) {
				go miner.LetSing(timer)
				timer <- t

				Eventually(fakeStateMachine.BecomesUnhealthyCallCount).Should(Equal(1))
				Consistently(fakeStateMachine.BecomesNotUnhealthyCallCount).Should(Equal(0))
				Consistently(fakeStateMachine.RemainsInSameStateCallCount).Should(Equal(0))

				close(done)
			})
		})
	})
})
