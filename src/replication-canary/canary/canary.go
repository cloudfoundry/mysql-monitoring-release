package canary

import (
	"time"

	"code.cloudfoundry.org/lager"

	"database/sql"

	"errors"

	"github.com/cloudfoundry/replication-canary/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SQLClient
type SQLClient interface {
	Setup(*sql.DB) error
	Write(*sql.DB, time.Time) error
	Check(*sql.DB, time.Time) (bool, error)
	Cleanup(*sql.DB) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Healthchecker
type Healthchecker interface {
	Healthy(*models.NamedConnection) (bool, error)
}

type Canary struct {
	sqlClient      SQLClient
	healthchecker  Healthchecker
	writeReadDelay time.Duration
	logger         lager.Logger
}

func NewCanary(
	sqlClient SQLClient,
	healthchecker Healthchecker,
	writeReadDelay time.Duration,
	logger lager.Logger,
) *Canary {
	return &Canary{
		sqlClient:      sqlClient,
		healthchecker:  healthchecker,
		writeReadDelay: writeReadDelay,
		logger:         logger,
	}
}

func (c *Canary) Setup(writeConn *models.NamedConnection) error {
	return c.sqlClient.Setup(writeConn.Connection)
}

// Chirp returns (true,nil) when we know replication succeeded
// Chirp returns (false,nil) when we know know replication failed
// Chirp returns (false,error) when we cannot determine if replication succeeded
// Chirp will never return (true,error)
func (c *Canary) Chirp(
	conns []*models.NamedConnection,
	writeConn *models.NamedConnection,
	timestamp time.Time,
) (bool, error) {
	c.logger.Debug("Beginning chirp", lager.Data{"timestamp": timestamp})

	err := c.checkWriteConn(timestamp, writeConn)
	if err != nil {
		return false, err
	}

	err = c.Write(timestamp, writeConn)
	if err != nil {
		return false, err
	}

	time.Sleep(c.writeReadDelay)

	replicationSuccess, err := c.Read(conns, timestamp)
	cleanupErr := c.cleanup(timestamp, writeConn)

	if err != nil {
		return false, err
	}

	if cleanupErr != nil {
		return false, cleanupErr
	}
	return replicationSuccess, nil
}

func (c *Canary) checkWriteConn(timestamp time.Time, writeConn *models.NamedConnection) error {
	c.logger.Debug(
		"Canary checking galera status for write connection",
		lager.Data{"timestamp": timestamp, "node": writeConn.Name},
	)

	writeHealthy, err := c.healthchecker.Healthy(writeConn)
	if err != nil {
		c.logger.Debug("Canary checking galera status errored", lager.Data{
			"node":         writeConn.Name,
			"errorMessage": err.Error(),
		})
		return err
	}

	if !writeHealthy {
		c.logger.Debug(
			"Canary found write connection's galera was unhealthy",
			lager.Data{"node": writeConn.Name},
		)
		return errors.New("write connection's galera is unhealthy")
	}

	return nil
}

// this function is exported for use in creating a write binary for exploring
func (c *Canary) Write(timestamp time.Time, writeConn *models.NamedConnection) error {
	c.logger.Debug(
		"Canary beginning write",
		lager.Data{"timestamp": timestamp, "node": writeConn.Name},
	)

	err := c.sqlClient.Write(writeConn.Connection, timestamp)
	if err != nil {
		return err
	}

	c.logger.Debug(
		"Canary completed write",
		lager.Data{"timestamp": timestamp, "node": writeConn.Name},
	)

	return nil
}

// this function is exported for use in creating a write binary for exploring
// already tested in the Chirp functions test
func (c *Canary) Read(conns []*models.NamedConnection, timestamp time.Time) (bool, error) {
	c.logger.Debug("Canary beginning reads", lager.Data{"timestamp": timestamp})
	// if there are more than 0 (false,nil) then we return 'false'
	var errs []error
	for _, conn := range conns {
		c.logger.Debug(
			"Canary checking galera status for read connection",
			lager.Data{"timestamp": timestamp, "node": conn.Name},
		)

		connHealthy, err := c.healthchecker.Healthy(conn)
		if err != nil {
			c.logger.Debug("Canary checking galera status errored", lager.Data{
				"node":         conn.Name,
				"errorMessage": err.Error(),
			})
			errs = append(errs, err)
			continue
		}

		if !connHealthy {
			c.logger.Debug(
				"Canary found read connection's galera was unhealthy",
				lager.Data{"node": conn.Name},
			)
			continue
		}

		c.logger.Debug(
			"Canary beginning check",
			lager.Data{"timestamp": timestamp, "node": conn.Name},
		)
		ok, err := c.sqlClient.Check(conn.Connection, timestamp)
		if !ok && err == nil {
			c.logger.Debug(
				"Canary detected replication failure",
				lager.Data{"timestamp": timestamp, "node": conn.Name},
			)
			return false, nil
		}

		c.logger.Debug(
			"Canary finished check",
			lager.Data{"timestamp": timestamp, "node": conn.Name},
		)

		if err != nil {
			c.logger.Debug("Canary errored during read", lager.Data{
				"timestamp":    timestamp,
				"errorMessage": err.Error(),
				"node":         conn.Name,
			})
			errs = append(errs, err)
		}
	}
	c.logger.Debug("Canary completed reads", lager.Data{"timestamp": timestamp})

	// If any of the Checks is non-deterministic, we are non-deterministic
	if len(errs) > 0 {
		return false, errs[0]
	}

	return true, nil
}

func (c *Canary) cleanup(timestamp time.Time, writeConn *models.NamedConnection) error {
	c.logger.Debug(
		"Canary beginning cleanup",
		lager.Data{"timestamp": timestamp, "node": writeConn.Name},
	)

	err := c.sqlClient.Cleanup(writeConn.Connection)

	c.logger.Debug(
		"Canary completed cleanup",
		lager.Data{"timestamp": timestamp, "node": writeConn.Name, "err": err},
	)

	return err
}
