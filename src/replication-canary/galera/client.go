package galera

import (
	"code.cloudfoundry.org/lager"
	"database/sql"
	"errors"
	"replication-canary/models"
)

type Client struct {
	Logger lager.Logger
}

type State int

const (
	Invalid State = iota
	Joining
	DonorDesynced
	Joined
	Synced
)

func (c Client) Healthy(db *models.NamedConnection) (bool, error) {
	var variableName string
	var value int

	err := db.Connection.QueryRow("SHOW STATUS LIKE 'wsrep_local_state'").Scan(&variableName, &value)

	if err == sql.ErrNoRows {
		c.Logger.Debug("No rows found containing 'wsrep_local_state'", lager.Data{"node": db.Name})

		return false, errors.New("wsrep_local_state variable not set (possibly not a galera db)")
	} else if err != nil {
		c.Logger.Debug("Error getting 'wsrep_local_state'", lager.Data{"node": db.Name, "errorMessage": err.Error()})

		return false, err
	}

	if State(value) != Synced {
		c.Logger.Debug("Node is not synced", lager.Data{"node": db.Name, "wsrep_local_state": value})
		return false, nil
	}

	return true, nil
}
