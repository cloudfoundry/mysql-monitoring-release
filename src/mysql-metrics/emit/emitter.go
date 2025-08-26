package emit

import (
	"log/slog"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Processor
type Processor interface {
	Process() error
}

type Emitter struct {
	processor Processor
	interval  time.Duration
	sleeper   func(time.Duration)
}

func NewEmitter(processor Processor, interval time.Duration, sleeper func(time.Duration)) Emitter {
	return Emitter{
		processor: processor,
		interval:  interval,
		sleeper:   sleeper,
	}
}

func (e Emitter) Start() {
	for {
		err := e.processor.Process()
		if err != nil {
			slog.Error("error processing metrics", "error", err)
		}
		e.sleeper(e.interval)
	}
}
