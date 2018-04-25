package disk

import (
	"mysql-diag/config"
	"mysql-diag/diagagentclient"
)

type NodeDiskInfo struct {
	Node config.MysqlNode
	Info *diagagentclient.InfoResponse
}
