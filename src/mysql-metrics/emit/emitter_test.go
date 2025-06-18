package emit_test

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/mysql-metrics/emit"
	"github.com/cloudfoundry/mysql-metrics/emit/emitfakes"
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
		emitter       emit.Emitter
		sleeper       *Sleeper
		sleepDuration time.Duration
		logBuffer     *gbytes.Buffer
	)

	BeforeEach(func() {
		// Set up global slog to write to a buffer for test assertions
		logBuffer = gbytes.NewBuffer()
		slog.SetDefault(slog.New(slog.NewJSONHandler(io.MultiWriter(GinkgoWriter, logBuffer), &slog.HandlerOptions{Level: slog.LevelDebug})))

		fakeProcessor = &emitfakes.FakeProcessor{}
		sleeper = &Sleeper{}
		sleepDuration = time.Duration(2) * time.Second

		emitter = emit.NewEmitter(
			fakeProcessor,
			sleepDuration,
			sleeper.Sleep,
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
			errs := errors.Join(
				errors.New("something bad happened"),
				errors.New("something else happened"),
				errors.New("this thing is busted"),
			)

			fakeProcessor.ProcessReturnsOnCall(0, errs)
			go emitter.Start()

			Eventually(func() int {
				return fakeProcessor.ProcessCallCount()
			}).Should(BeNumerically(">", 2))

			Eventually(logBuffer).Should(gbytes.Say(`"level":"ERROR"`))
			Eventually(logBuffer).Should(gbytes.Say(`"msg":"error processing metrics"`))
			Eventually(logBuffer).Should(gbytes.Say("something bad happened"))
			Eventually(logBuffer).Should(gbytes.Say("something else happened"))
			Eventually(logBuffer).Should(gbytes.Say("this thing is busted"))
		})
	})
})
