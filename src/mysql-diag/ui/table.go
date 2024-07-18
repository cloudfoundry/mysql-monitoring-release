package ui

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"sync"

	"code.cloudfoundry.org/bytefmt"
	"github.com/olekukonko/tablewriter"

	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/disk"
	"github.com/cloudfoundry/mysql-diag/msg"
)

type Table struct {
	table *tablewriter.Table
	mutex sync.RWMutex

	diskInfo    []diskInfo
	clusterInfo []clusterInfo
}

const errorContent = "N/A - ERROR"
const maxUUID = "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"

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

func prettifyDisk(info diagagentclient.DiskInfo) string {
	percentageUsed := float64(info.BytesTotal-info.BytesFree) / float64(info.BytesTotal) * 100
	return fmt.Sprintf("%s / %s (%.1f%%)", bytefmt.ByteSize(info.BytesTotal-info.BytesFree), bytefmt.ByteSize(info.BytesTotal), percentageUsed)
}

type diskInfo struct {
	instance       string
	persistentDisk string
	ephemeralDisk  string
}

func (t *Table) AddClusterInfo(name string, uuid string, galeraStatus *database.GaleraStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	wsrepLocalState, wsrepClusterStatus, wsrepLocalIndex := errorContent, errorContent, maxUUID

	if galeraStatus != nil {
		wsrepLocalState, wsrepClusterStatus, wsrepLocalIndex = galeraStatus.LocalState, galeraStatus.ClusterStatus, galeraStatus.LocalIndex
	}

	t.clusterInfo = append(t.clusterInfo, clusterInfo{
		instance:      nameUUID,
		localState:    wsrepLocalState,
		clusterStatus: wsrepClusterStatus,
		localIndex:    wsrepLocalIndex,
	})
}

func (t *Table) AddClusterData(nodeStatuses []*database.NodeClusterStatus) {
	for _, ns := range nodeStatuses {
		n := ns.Node
		t.AddClusterInfo(n.Name, n.UUID, ns.Status)
	}
}

func (t *Table) AddDiskData(nodeDiskInfos []disk.NodeDiskInfo) {
	if HasAtLeastOneInfo(nodeDiskInfos) {
		for _, nd := range nodeDiskInfos {
			n := nd.Node
			t.AddDiskInfo(n.Name, n.UUID, nd.Info)
		}
	} else {
		fmt.Println(msg.Alert("Unable to gather disk usage information, moving on. Run bosh vms --vitals for this information."))
	}
}

func HasAtLeastOneInfo(infos []disk.NodeDiskInfo) bool {
	return slices.ContainsFunc(infos, func(i disk.NodeDiskInfo) bool {
		return i.Info != nil
	})
}

type clusterInfo struct {
	instance      string
	localState    string
	clusterStatus string
	localIndex    string
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
			r.localIndex = n.localIndex
		}
	}

	var rows [][]string
	for k, v := range result {
		rows = append(rows, []string{k, v.localIndex, v.localState, v.clusterStatus, v.persistentDisk, v.ephemeralDisk})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		}

		return rows[i][1] < rows[j][1]
	})

	for idx, r := range rows {
		t.table.Append([]string{fmt.Sprintf(" [%d] %s", idx, r[0]), r[2], r[3], r[4], r[5]})
	}
}

type row struct {
	persistentDisk string
	ephemeralDisk  string
	localState     string
	clusterStatus  string
	localIndex     string
}

func (t *Table) Render() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.aggregateInfo()
	t.table.Render()
}
