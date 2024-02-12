package ui

import (
	"fmt"
	"io"
	"sync"

	"github.com/olekukonko/tablewriter"

	"github.com/cloudfoundry/mysql-diag/database"
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

	cst.table.SetHeader([]string{"INSTANCE", "STATE", "CLUSTER STATUS"})

	return &cst
}

func (t *ClusterStateTable) Add(name string, uuid string, galeraStatus *database.GaleraStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	wsrepLocalState, wsrepClusterStatus := errorContent, errorContent

	if galeraStatus != nil {
		wsrepLocalState, wsrepClusterStatus = galeraStatus.LocalState, galeraStatus.ClusterStatus
	}

	row := []string{nameUUID, wsrepLocalState, wsrepClusterStatus}
	t.table.Append(row)
}

func (t *ClusterStateTable) Render() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.table.Render()
}
