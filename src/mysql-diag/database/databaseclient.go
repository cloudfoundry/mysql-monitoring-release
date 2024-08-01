package database

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/msg"
)

type GaleraStatus struct {
	LocalState    string
	ClusterSize   int
	ClusterStatus string
	ReadOnly      bool
	LocalIndex    string
	LastApplied   int
}

type NodeClusterStatus struct {
	Node   config.MysqlNode
	Status *GaleraStatus
}

type DatabaseClient struct {
	connection *sql.DB
}

func NewDatabaseClient(connection *sql.DB) *DatabaseClient {
	return &DatabaseClient{
		connection: connection,
	}
}

func (c *DatabaseClient) getReadOnly(status *GaleraStatus) error {
	var unused string
	var readOnly string

	err := c.connection.QueryRow("SHOW GLOBAL VARIABLES LIKE 'read_only'").Scan(&unused, &readOnly)
	if err != nil {
		return err
	}

	if readOnly == "ON" {
		status.ReadOnly = true
	}

	return nil
}

func (c *DatabaseClient) getGaleraStatus(status *GaleraStatus) error {

	rows, err := c.connection.Query("SHOW STATUS LIKE 'wsrep_%'")

	if err == sql.ErrNoRows {
		return errors.New("wsrep_% status variables missing (possibly not a galera db)")
	} else if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value string

		err := rows.Scan(&key, &value)
		if err != nil {
			return err
		}

		switch key {
		case "wsrep_local_state_comment":
			status.LocalState = value
		case "wsrep_cluster_status":
			status.ClusterStatus = value
		case "wsrep_cluster_size":
			status.ClusterSize, err = strconv.Atoi(value)
			if err != nil {
				return err
			}
		case "wsrep_local_index":
			status.LocalIndex = value
		case "wsrep_last_applied":
			status.LastApplied, err = strconv.Atoi(value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *DatabaseClient) Status() (*GaleraStatus, error) {

	status := GaleraStatus{}

	err := c.getGaleraStatus(&status)
	if err != nil {
		return nil, err
	}

	err = c.getReadOnly(&status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func GetNodeClusterStatuses(mysqlConfig config.MysqlConfig, nodeClusterStatus map[string]*NodeClusterStatus) {
	channel := make(chan NodeClusterStatus, len(mysqlConfig.Nodes))

	for _, status := range nodeClusterStatus {
		n := status.Node
		s := status.Status

		go func() {
			ac := NewDatabaseClient(mysqlConfig.Connection(n))
			galeraStatus, err := ac.Status()
			if err != nil {
				msg.PrintfErrorIntro("", "error retrieving galera status from node %s: %v", n.Name+"/"+n.UUID, err)
			}
			if galeraStatus == nil {
				galeraStatus = s
			}
			channel <- NodeClusterStatus{Node: n, Status: galeraStatus}
		}()
	}

	for i := 0; i < len(mysqlConfig.Nodes); i++ {
		ns := <-channel
		nodeClusterStatus[ns.Node.Host] = &ns
	}
}
