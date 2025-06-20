package emit

import (
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Logger
type Logger interface {
	Error(string, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Processor
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
			e.logger.Error("error processing metrics", err)
		}
		e.sleeper(e.interval)
	}
}
