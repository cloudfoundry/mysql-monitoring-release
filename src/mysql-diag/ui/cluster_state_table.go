package ui

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"mysql-diag/database"
	"io"
	"strconv"
	"sync"
)

type ClusterStateTable struct {
	table *tablewriter.Table
	mutex sync.RWMutex
}

const errorContent = "N/A - ERROR"

func NewClusterStateTable(writer io.Writer) *ClusterStateTable {
	cst := ClusterStateTable{
		table: tablewriter.NewWriter(writer),
	}

	cst.table.SetHeader([]string{"HOST", "NAME/UUID", "WSREP LOCAL STATE", "WSREP CLUSTER STATUS", "WSREP CLUSTER SIZE"})

	return &cst
}

func (t *ClusterStateTable) Add(host string, name string, uuid string, galeraStatus *database.GaleraStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(uuid) > 8 {
		uuid = uuid[:8]
	}
	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	wsrepLocalState, wsrepClusterStatus, wsrepClusterSize := errorContent, errorContent, errorContent

	if galeraStatus != nil {
		wsrepLocalState, wsrepClusterStatus, wsrepClusterSize = galeraStatus.LocalState, galeraStatus.ClusterStatus, strconv.Itoa(galeraStatus.ClusterSize)
	}

	row := []string{host, nameUUID, wsrepLocalState, wsrepClusterStatus, wsrepClusterSize}
	t.table.Append(row)
}

func (t *ClusterStateTable) Render() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.table.Render()
}
