package emit

import (
	"github.com/hashicorp/go-multierror"
	"time"
)

//go:generate counterfeiter . Logger
type Logger interface {
	Error(string, error)
}

//go:generate counterfeiter . Processor
type Processor interface {
	Process() error
}

type Emitter struct {
	processor Processor
	interval  time.Duration
	sleeper   func(time.Duration)
	logger    Logger
}

func NewEmitter(processor Processor, interval time.Duration, sleeper func(time.Duration), logger Logger) Emitter {
	return Emitter{
		processor: processor,
		interval:  interval,
		sleeper:   sleeper,
		logger:    logger,
	}
}

func (e Emitter) Start() {
	for {
		err := e.processor.Process()
		if err != nil {
			combinedErr, ok := err.(*multierror.Error)
			if ok {
				for _, wrappedError := range combinedErr.WrappedErrors() {
					e.logger.Error("error processing metrics", wrappedError)
				}
			} else {
				e.logger.Error("error processing metrics", err)
			}
		}
		e.sleeper(e.interval)
	}
}
