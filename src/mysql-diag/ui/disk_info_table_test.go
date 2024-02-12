package ui_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/ui"
)

var _ = Describe("DiskInfoTable", func() {

	It("prints a table with some rows", func() {
		buffer := new(bytes.Buffer)

		cst := ui.NewDiskInfoTable(buffer)

		cst.Add("name1", "64bfefb0-97fd-4e34-b0fb-499ccb684faa", &diagagentclient.InfoResponse{
			Persistent: diagagentclient.DiskInfo{
				BytesTotal:  456,
				BytesFree:   123,
				InodesTotal: 789000,
				InodesFree:  567000,
			},
			Ephemeral: diagagentclient.DiskInfo{
				BytesTotal:  145600,
				BytesFree:   112300,
				InodesTotal: 1789000,
				InodesFree:  1567000,
			},
		})
		cst.Add("name2", "4a136f80-d88d-447c-bce8-4b87492110a7", &diagagentclient.InfoResponse{
			Persistent: diagagentclient.DiskInfo{
				BytesTotal:  456,
				BytesFree:   123,
				InodesTotal: 7890000,
				InodesFree:  5670000,
			},
			Ephemeral: diagagentclient.DiskInfo{
				BytesTotal:  1456,
				BytesFree:   1123,
				InodesTotal: 1789000,
				InodesFree:  1567000,
			},
		})
		cst.Add("name3", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		cst.Render()

		Expect(buffer.String()).To(Equal(

			`+--------------------------------------------+----------------------+------------------------+
|                  INSTANCE                  | PERSISTENT DISK USED |  EPHEMERAL DISK USED   |
+--------------------------------------------+----------------------+------------------------+
| name1/64bfefb0-97fd-4e34-b0fb-499ccb684faa | 333B / 456B (73.0%)  | 32.5K / 142.2K (22.9%) |
| name2/4a136f80-d88d-447c-bce8-4b87492110a7 | 333B / 456B (73.0%)  | 333B / 1.4K (22.9%)    |
| name3/3a79c040-1d3a-4583-8566-9c7097760baa | N/A - ERROR          | N/A - ERROR            |
+--------------------------------------------+----------------------+------------------------+
`,
		))
	})
})
