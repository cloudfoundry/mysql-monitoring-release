package database_client

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	configPackage "github.com/cloudfoundry/mysql-metrics/config"
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

func (dc *DbClient) ShowReplicaStatus() (map[string]string, error) {
	return dc.runSingleRowQuery("SHOW REPLICA STATUS", []interface{}{})
}

func (dc *DbClient) IsFollower() (bool, error) {
	replicaStatus, err := dc.ShowReplicaStatus()
	return len(replicaStatus) > 0, err
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

func (dc *DbClient) FindLastBackupTimestamp() (time.Time, error) {
	results, err := dc.runSingleRowQuery("SELECT ts AS timestamp FROM backup_metrics.backup_times ORDER BY ts DESC LIMIT 1", []interface{}{})
	if err != nil {
		return time.Time{}, err
	}

	if results["timestamp"] == "" {
		return time.Time{}, nil
	}

	value, err := time.Parse("2006-01-02 15:04:05", results["timestamp"])
	if err != nil {
		return time.Time{}, err
	}

	return value, nil
}

func (dc *DbClient) runSingleRowQuery(query string, params []any) (map[string]string, error) {
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

	nullOrString := func(v any) string {
		value := v.(*sql.NullString)
		if value.Valid {
			return value.String
		} else {
			return "NULL"
		}
	}

	if rows.Next() {
		scanValues := make([]any, len(columns))
		for i := range scanValues {
			scanValues[i] = new(sql.NullString)
		}

		if err = rows.Scan(scanValues...); err != nil {
			return nil, err
		}
		for i, col := range columns {
			result[strings.ToLower(col)] = nullOrString(scanValues[i])
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
