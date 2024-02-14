package ui

import (
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/olekukonko/tablewriter"

	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
)

type Table struct {
	table *tablewriter.Table
	mutex sync.RWMutex

	diskInfo    []diskInfo
	clusterInfo []clusterInfo
}

//const errorContent = "N/A - ERROR"

func NewTable(writer io.Writer) *Table {
	t := Table{
		table: tablewriter.NewWriter(writer),
	}

	t.table.SetAutoWrapText(false)
	t.table.SetHeader([]string{"INSTANCE", "STATE", "CLUSTER STATUS", "PERSISTENT DISK USED", "EPHEMERAL DISK USED"})

	return &t
}

func (t *Table) AddDiskInfo(name string, uuid string, info *diagagentclient.InfoResponse) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	persistentDisk, ephemeralDisk := errorContent, errorContent

	if info != nil {
		persistentDisk = prettifyDisk(info.Persistent)
		ephemeralDisk = prettifyDisk(info.Ephemeral)
	}

	t.diskInfo = append(t.diskInfo, diskInfo{
		instance:       nameUUID,
		persistentDisk: persistentDisk,
		ephemeralDisk:  ephemeralDisk,
	})
}

//func prettifyDisk(info diagagentclient.DiskInfo) string {
//	percentageUsed := float64(info.BytesTotal-info.BytesFree) / float64(info.BytesTotal) * 100
//	return fmt.Sprintf("%s / %s (%.1f%%)", bytefmt.ByteSize(info.BytesTotal-info.BytesFree), bytefmt.ByteSize(info.BytesTotal), percentageUsed)
//}

type diskInfo struct {
	instance       string
	persistentDisk string
	ephemeralDisk  string
}

func (t *Table) AddClusterInfo(name string, uuid string, galeraStatus *database.GaleraStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	wsrepLocalState, wsrepClusterStatus := errorContent, errorContent

	if galeraStatus != nil {
		wsrepLocalState, wsrepClusterStatus = galeraStatus.LocalState, galeraStatus.ClusterStatus
	}

	t.clusterInfo = append(t.clusterInfo, clusterInfo{
		instance:      nameUUID,
		localState:    wsrepLocalState,
		clusterStatus: wsrepClusterStatus,
	})
}

type clusterInfo struct {
	instance      string
	localState    string
	clusterStatus string
}

func (t *Table) aggregateInfo() {
	result := map[string]*row{}

	for _, d := range t.diskInfo {
		if _, ok := result[d.instance]; !ok {
			result[d.instance] = &row{}
		}

		if r, ok := result[d.instance]; ok {
			r.persistentDisk = d.persistentDisk
			r.ephemeralDisk = d.ephemeralDisk
		}
	}

	for _, n := range t.clusterInfo {
		if _, ok := result[n.instance]; !ok {
			result[n.instance] = &row{}
		}

		if r, ok := result[n.instance]; ok {
			r.localState = n.localState
			r.clusterStatus = n.clusterStatus
		}
	}

	var rows [][]string
	for k, v := range result {
		rows = append(rows, []string{k, v.localState, v.clusterStatus, v.persistentDisk, v.ephemeralDisk})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	for _, r := range rows {
		t.table.Append([]string{r[0], r[1], r[2], r[3], r[4]})
	}
}

type row struct {
	persistentDisk string
	ephemeralDisk  string
	localState     string
	clusterStatus  string
}

func (t *Table) Render() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.aggregateInfo()
	t.table.Render()
}
