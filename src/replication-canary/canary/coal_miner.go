package canary

import (
	"time"

	"github.com/cloudfoundry/replication-canary/models"

	"code.cloudfoundry.org/lager"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ConnectionFactory
type ConnectionFactory interface {
	Conns() ([]*models.NamedConnection, error)
	WriteConn() (*models.NamedConnection, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Chirper
type Chirper interface {
	Chirp(
		conns []*models.NamedConnection,
		writeConn *models.NamedConnection,
		timestamp time.Time,
	) (bool, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Alerter
type Alerter interface {
	NotUnhealthy(time.Time) error
	Unhealthy(time.Time) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . StateMachine
type StateMachine interface {
	BecomesUnhealthy(time.Time)
	BecomesNotUnhealthy(time.Time)
	RemainsInSameState(time.Time)
	GetState() State
}

type CoalMiner struct {
	connectionFactory ConnectionFactory
	chirper           Chirper
	alerter           Alerter
	logger            lager.Logger

	StateMachine StateMachine
}

func NewCoalMiner(
	connectionFactory ConnectionFactory,
	chirper Chirper,
	alerter Alerter,
	logger lager.Logger,
) *CoalMiner {
	return &CoalMiner{
		connectionFactory: connectionFactory,
		chirper:           chirper,
		alerter:           alerter,
		logger:            logger,

		StateMachine: &StatefulStateMachine{
			Logger: logger,
			OnBecomesUnhealthy: func(timestamp time.Time) {
				err := alerter.Unhealthy(timestamp)
				if err != nil {
					logger.Error("failure in alerting", err)
				}
			},
			OnBecomesNotUnhealthy: func(timestamp time.Time) {
				err := alerter.NotUnhealthy(timestamp)
				if err != nil {
					logger.Error("failure in alerting", err)
				}
			},
		},
	}
}

func (c *CoalMiner) LetSing(timer <-chan time.Time) {
	for timestamp := range timer {
		writeConn, err := c.connectionFactory.WriteConn()
		if err != nil {
			c.logger.Error(
				"Coalminer failed to obtain write connection",
				err,
				lager.Data{"timestamp": timestamp},
			)
			c.StateMachine.RemainsInSameState(timestamp)
			continue
		}

		conns, err := c.connectionFactory.Conns()
		if err != nil {
			c.logger.Error(
				"Coalminer failed to obtain connections",
				err,
				lager.Data{"timestamp": timestamp},
			)
			c.StateMachine.RemainsInSameState(timestamp)
			continue
		}

		ok, err := c.chirper.Chirp(conns, writeConn, timestamp)
		c.logger.Debug("Coalminer considering state", lager.Data{"timestamp": timestamp})
		c.ParseReplicationHealth(ok, timestamp, err)
	}
}

// Only exported for the purses of using in reader binary
func (c *CoalMiner) ParseReplicationHealth(ok bool, timestamp time.Time, err error) {
	if !ok && err == nil {
		c.logger.Debug(
			"Coalminer received not ok and nil error (replication failure)",
			lager.Data{"timestamp": timestamp},
		)
		c.StateMachine.BecomesUnhealthy(timestamp)
	}
	if ok {
		c.logger.Debug(
			"Coalminer received ok (replication success)",
			lager.Data{"timestamp": timestamp},
		)
		c.StateMachine.BecomesNotUnhealthy(timestamp)
	}
	if !ok && err != nil {
		c.logger.Debug(
			"Coalminer received not ok and non-nil error (cannot determine replication state)",
			lager.Data{"timestamp": timestamp},
		)
		c.StateMachine.BecomesNotUnhealthy(timestamp)
	}

	if err != nil {
		c.logger.Error("failure in chirping", err)
	}
}
