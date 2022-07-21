package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry/replication-canary/config"
	"github.com/cloudfoundry/replication-canary/models"

	"github.com/go-sql-driver/mysql"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SwitchboardClient
type SwitchboardClient interface {
	ActiveBackendHost() (string, error)
}

type ConnectionFactory struct {
	switchboardClients []SwitchboardClient
	logger             lager.Logger

	clusterIPs            []string
	port                  int
	galeraHealthcheckPort int

	canaryDatabase string
	canaryUsername string
	canaryPassword string

	conns []*models.NamedConnection

	OpenConn func(dsn string) (*sql.DB, error)
}

func NewConnectionFactoryFromConfig(
	c *config.Config,
	switchboardClients []SwitchboardClient,
	logger lager.Logger,
) *ConnectionFactory {
	return &ConnectionFactory{
		switchboardClients: switchboardClients,
		logger:             logger,

		clusterIPs:            c.MySQL.ClusterIPs,
		port:                  c.MySQL.Port,
		galeraHealthcheckPort: c.MySQL.GaleraHealthcheckPort,

		canaryDatabase: c.Canary.Database,
		canaryUsername: c.Canary.Username,
		canaryPassword: c.Canary.Password,
		OpenConn:       openConnection,
	}
}

func (c *ConnectionFactory) Conns() ([]*models.NamedConnection, error) {
	if len(c.conns) > 0 {
		return c.conns, nil
	}

	var errs []error

	var conns []*models.NamedConnection
	for _, ip := range c.clusterIPs {
		cfg := &mysql.Config{
			User:                 c.canaryUsername,
			Passwd:               c.canaryPassword,
			DBName:               c.canaryDatabase,
			Net:                  "tcp",
			Addr:                 fmt.Sprintf("%s:%d", ip, c.port),
			AllowNativePasswords: true,
			CheckConnLiveness:    true,
			MaxAllowedPacket:     4194304,
			TLSConfig:            "preferred",
		}

		conn, err := c.OpenConn(cfg.FormatDSN())
		conns = append(conns, &models.NamedConnection{
			Name:       cfg.Addr,
			Connection: conn,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Close all open connections if any of them errored, so we don't leak connections
	if len(errs) > 0 {
		for _, conn := range conns {
			conn.Connection.Close()
		}

		return nil, errs[0]
	}

	c.conns = conns

	return c.conns, nil
}

func (c *ConnectionFactory) WriteConn() (*models.NamedConnection, error) {
	conns, err := c.Conns()
	if err != nil {
		return nil, err
	}
	var host string
	for _, switchboardClient := range c.switchboardClients {
		host, err = switchboardClient.ActiveBackendHost()
		if host != "" {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	c.logger.Info("using write connection", lager.Data{"host": host})

	hostPrefix := fmt.Sprintf("%s:", host)
	for _, namedConn := range conns {
		if strings.HasPrefix(namedConn.Name, hostPrefix) {
			return namedConn, nil
		}
	}

	return nil, errors.New("no connection found for active write host")
}

func openConnection(dsn string) (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}
