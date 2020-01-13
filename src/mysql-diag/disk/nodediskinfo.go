package disk

import (
	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
)

type NodeDiskInfo struct {
	Node config.MysqlNode
	Info *diagagentclient.InfoResponse
}
