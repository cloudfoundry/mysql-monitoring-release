package ui_test

import (
	"bytes"

	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiskInfoTable", func() {

	It("prints a table with some rows", func() {
		buffer := new(bytes.Buffer)

		cst := ui.NewDiskInfoTable(buffer)

		cst.Add("host1", "name1", "uuid1", &diagagentclient.InfoResponse{
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
		cst.Add("host2", "name2", "uuid2", &diagagentclient.InfoResponse{
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
		cst.Add("host3", "name3", "uuid3", nil)

		cst.Render()

		Expect(buffer.String()).To(Equal(
			`+-------+-------------+---------------------------------------+-----------------------------------------+
| HOST  |  NAME/UUID  |         PERSISTENT DISK USED          |           EPHEMERAL DISK USED           |
+-------+-------------+---------------------------------------+-----------------------------------------+
| host1 | name1/uuid1 | 73.0% of 456B (28.1% of 0.79M inodes) | 22.9% of 142.2K (12.4% of 1.79M inodes) |
| host2 | name2/uuid2 | 73.0% of 456B (28.1% of 7.89M inodes) | 22.9% of 1.4K (12.4% of 1.79M inodes)   |
| host3 | name3/uuid3 | N/A - ERROR                           | N/A - ERROR                             |
+-------+-------------+---------------------------------------+-----------------------------------------+
`,
		))
	})
})
