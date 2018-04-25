package ui

import (
	"fmt"
	"github.com/cloudfoundry/bytefmt"
	"github.com/olekukonko/tablewriter"
	"mysql-diag/diagagentclient"
	"io"
	"sync"
)

type DiskInfoTable struct {
	table *tablewriter.Table
	mutex sync.RWMutex
}

func NewDiskInfoTable(writer io.Writer) *DiskInfoTable {
	cst := DiskInfoTable{
		table: tablewriter.NewWriter(writer),
	}

	cst.table.SetAutoWrapText(false)
	cst.table.SetHeader([]string{"HOST", "NAME/UUID", "PERSISTENT DISK USED", "EPHEMERAL DISK USED"})

	return &cst
}

func (t *DiskInfoTable) Add(host string, name string, uuid string, info *diagagentclient.InfoResponse) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(uuid) > 8 {
		uuid = uuid[:8]
	}
	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	persistentDisk, ephemeralDisk := errorContent, errorContent

	if info != nil {
		persistentDisk = prettifyDisk(info.Persistent)
		ephemeralDisk = prettifyDisk(info.Ephemeral)
	}

	row := []string{host, nameUUID, persistentDisk, ephemeralDisk}
	t.table.Append(row)
}

func prettifyDisk(info diagagentclient.DiskInfo) string {
	percentageUsed := float64(info.BytesTotal-info.BytesFree) / float64(info.BytesTotal) * 100
	diskStr := fmt.Sprintf("%.1f%% of %s", percentageUsed, bytefmt.ByteSize(info.BytesTotal))

	inodePercentageUsed := float64(info.InodesTotal-info.InodesFree) / float64(info.InodesTotal) * 100
	inodeStr := fmt.Sprintf("%.1f%% of %.2fM", inodePercentageUsed, float64(info.InodesTotal)/float64(1000000))

	return fmt.Sprintf("%s (%s inodes)", diskStr, inodeStr)
}

func (t *DiskInfoTable) Render() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.table.Render()
}
