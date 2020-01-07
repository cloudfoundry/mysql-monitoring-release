package database_client

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	configPackage "github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/config"
)

type DbClient struct {
	connection *sql.DB
	config     *configPackage.Config
}

func QuoteIdentifier(identifier string) string {
	return "`" + strings.Replace(identifier, "`", "``", -1) + "`"
}

func NewDatabaseClient(connection *sql.DB, config *configPackage.Config) *DbClient {
	return &DbClient{
		connection: connection,
		config:     config,
	}
}

func (dc *DbClient) IsAvailable() bool {
	_, err := dc.runKeyValueQuery("SHOW GLOBAL STATUS")
	return nil == err
}

func (dc *DbClient) ShowGlobalStatus() (map[string]string, error) {
	return dc.runKeyValueQuery("SHOW GLOBAL STATUS")
}

func (dc *DbClient) ShowGlobalVariables() (map[string]string, error) {
	return dc.runKeyValueQuery("SHOW GLOBAL VARIABLES")
}

func (dc *DbClient) ShowSlaveStatus() (map[string]string, error) {
	return dc.runSingleRowQuery("SHOW SLAVE STATUS", []interface{}{})
}

func (dc *DbClient) IsFollower() (bool, error) {
	slaveStatus, err := dc.ShowSlaveStatus()
	return len(slaveStatus) > 0, err
}

func (dc *DbClient) HeartbeatStatus() (map[string]string, error) {
	sql := "SELECT UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(timestamp) AS seconds_since_leader_heartbeat FROM " +
		QuoteIdentifier(dc.config.HeartbeatDatabase) + "." + QuoteIdentifier(dc.config.HeartbeatTable)
	return dc.runSingleRowQuery(sql, []interface{}{})
}

func (dc *DbClient) ServicePlansDiskAllocated() (map[string]string, error) {
	row, err := dc.runSingleRowQuery("SELECT SUM(max_storage_mb) AS service_plans_disk_allocated FROM mysql_broker.service_instances",
		[]interface{}{})
	if err != nil {
		return nil, err
	}
	if row["service_plans_disk_allocated"] == "NULL" {
		row["service_plans_disk_allocated"] = "0"
	}
	return row, nil
}

func (dc *DbClient) runSingleRowQuery(query string, params []interface{}) (map[string]string, error) {
	rows, err := dc.connection.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columns {
			columnPointers[i] = &columnValues[i]
		}
		err := rows.Scan(columnPointers...)
		if err != nil {
			panic(err)
		}
		for index, col := range columns {
			value := columnValues[index]
			if value == nil {
				value = []uint8("NULL")
			}
			result[strings.ToLower(col)] = string(value.([]uint8))
		}
	}

	return result, nil
}

func (dc *DbClient) runKeyValueQuery(query string) (map[string]string, error) {
	rows, err := dc.connection.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	globalStatusVariables := make(map[string]string)
	for rows.Next() {
		var key, value string
		rows.Scan(&key, &value)
		globalStatusVariables[strings.ToLower(key)] = value

	}

	return globalStatusVariables, nil
}
