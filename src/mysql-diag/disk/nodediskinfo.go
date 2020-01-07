package disk

import (
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/config"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/diagagentclient"
)

type NodeDiskInfo struct {
	Node config.MysqlNode
	Info *diagagentclient.InfoResponse
}
