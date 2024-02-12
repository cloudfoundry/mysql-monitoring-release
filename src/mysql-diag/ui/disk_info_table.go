package ui

import (
	"fmt"
	"io"
	"sync"

	"code.cloudfoundry.org/bytefmt"
	"github.com/olekukonko/tablewriter"

	"github.com/cloudfoundry/mysql-diag/diagagentclient"
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
	cst.table.SetHeader([]string{"INSTANCE", "PERSISTENT DISK USED", "EPHEMERAL DISK USED"})

	return &cst
}

func (t *DiskInfoTable) Add(name string, uuid string, info *diagagentclient.InfoResponse) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nameUUID := fmt.Sprintf("%s/%s", name, uuid)

	persistentDisk, ephemeralDisk := errorContent, errorContent

	if info != nil {
		persistentDisk = prettifyDisk(info.Persistent)
		ephemeralDisk = prettifyDisk(info.Ephemeral)
	}

	row := []string{nameUUID, persistentDisk, ephemeralDisk}
	t.table.Append(row)
}

func prettifyDisk(info diagagentclient.DiskInfo) string {
	percentageUsed := float64(info.BytesTotal-info.BytesFree) / float64(info.BytesTotal) * 100
	return fmt.Sprintf("%s / %s (%.1f%%)", bytefmt.ByteSize(info.BytesTotal-info.BytesFree), bytefmt.ByteSize(info.BytesTotal), percentageUsed)
}

func (t *DiskInfoTable) Render() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.table.Render()
}
