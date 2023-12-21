package emit_test

import (
	"errors"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-metrics/emit"
	"github.com/cloudfoundry/mysql-metrics/emit/emitfakes"
	"github.com/hashicorp/go-multierror"
)

type Sleeper struct {
	sync.Mutex
	callCount              int
	lastDurationCalledWith time.Duration
}

func (s *Sleeper) Sleep(d time.Duration) {
	s.Lock()
	defer s.Unlock()

	s.callCount++
	s.lastDurationCalledWith = d
	time.Sleep(1 * time.Millisecond)
}

func (s *Sleeper) CallCount() int {
	s.Lock()
	defer s.Unlock()

	return s.callCount
}

func (s *Sleeper) LastDuration() time.Duration {
	s.Lock()
	defer s.Unlock()

	return s.lastDurationCalledWith
}

var _ = Describe("Emitter", func() {
	var (
		fakeProcessor *emitfakes.FakeProcessor
		fakeLogger    *emitfakes.FakeLogger
		emitter       emit.Emitter
		sleeper       *Sleeper
		sleepDuration time.Duration
	)

	BeforeEach(func() {
		fakeProcessor = &emitfakes.FakeProcessor{}
		fakeLogger = &emitfakes.FakeLogger{}
		sleeper = &Sleeper{}
		sleepDuration = time.Duration(2) * time.Second

		emitter = emit.NewEmitter(
			fakeProcessor,
			sleepDuration,
			sleeper.Sleep,
			fakeLogger,
		)
	})

	It("calls processor.Process in a loop", func() {
		Expect(sleeper.CallCount()).To(Equal(0))

		go emitter.Start() // start this in the background

		Eventually(func() int {
			return sleeper.CallCount()
		}).Should(BeNumerically(">", 100))

		Expect(sleeper.LastDuration()).To(Equal(sleepDuration))
		Expect(fakeProcessor.ProcessCallCount()).To(BeNumerically(">", 100))
	})

	Context("error cases", func() {
		It("logs errors as they occur and continues to loop", func() {
			errs := multierror.Append(errors.New("something bad happened"),
				errors.New("something else happened"),
				errors.New("this thing is busted"),
			)

			fakeProcessor.ProcessReturnsOnCall(0, errs)
			go emitter.Start()

			Eventually(func() int {
				return fakeProcessor.ProcessCallCount()
			}).Should(BeNumerically(">", 2))

			errorMessage, err := fakeLogger.ErrorArgsForCall(0)
			Expect(errorMessage).To(Equal("error processing metrics"))
			Expect(err).To(MatchError("something bad happened"))

			errorMessage, err = fakeLogger.ErrorArgsForCall(1)
			Expect(errorMessage).To(Equal("error processing metrics"))
			Expect(err).To(MatchError("something else happened"))

			errorMessage, err = fakeLogger.ErrorArgsForCall(2)
			Expect(errorMessage).To(Equal("error processing metrics"))
			Expect(err).To(MatchError("this thing is busted"))

		})
	})
})
