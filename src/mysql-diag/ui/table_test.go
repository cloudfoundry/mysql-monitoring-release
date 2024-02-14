package ui_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/ui"
)

var _ = Describe("Table", func() {
	It("prints a table with disk info", func() {
		buffer := new(bytes.Buffer)

		table := ui.NewTable(buffer)

		table.AddDiskInfo("mysql", "64bfefb0-97fd-4e34-b0fb-499ccb684faa", &diagagentclient.InfoResponse{
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
		table.AddDiskInfo("mysql", "4a136f80-d88d-447c-bce8-4b87492110a7", &diagagentclient.InfoResponse{
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
		table.AddDiskInfo("mysql", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.Render()

		Expect(buffer.String()).To(Equal(

			`+--------------------------------------------+-------+----------------+----------------------+------------------------+
|                  INSTANCE                  | STATE | CLUSTER STATUS | PERSISTENT DISK USED |  EPHEMERAL DISK USED   |
+--------------------------------------------+-------+----------------+----------------------+------------------------+
| mysql/3a79c040-1d3a-4583-8566-9c7097760baa |       |                | N/A - ERROR          | N/A - ERROR            |
| mysql/4a136f80-d88d-447c-bce8-4b87492110a7 |       |                | 333B / 456B (73.0%)  | 333B / 1.4K (22.9%)    |
| mysql/64bfefb0-97fd-4e34-b0fb-499ccb684faa |       |                | 333B / 456B (73.0%)  | 32.5K / 142.2K (22.9%) |
+--------------------------------------------+-------+----------------+----------------------+------------------------+
`,
		))
	})

	It("prints a table with cluster state info", func() {
		buffer := new(bytes.Buffer)

		table := ui.NewTable(buffer)

		table.AddClusterInfo("mysql", "64bfefb0-97fd-4e34-b0fb-499ccb684faa", &database.GaleraStatus{
			LocalState:    "localState1",
			ClusterSize:   1,
			ClusterStatus: "clusterStatus1",
		})
		table.AddClusterInfo("mysql", "4a136f80-d88d-447c-bce8-4b87492110a7", &database.GaleraStatus{
			LocalState:    "localState2",
			ClusterSize:   2,
			ClusterStatus: "clusterStatus2",
		})
		table.AddClusterInfo("name3", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.Render()

		Expect(buffer.String()).To(Equal(
			`+--------------------------------------------+-------------+----------------+----------------------+---------------------+
|                  INSTANCE                  |    STATE    | CLUSTER STATUS | PERSISTENT DISK USED | EPHEMERAL DISK USED |
+--------------------------------------------+-------------+----------------+----------------------+---------------------+
| mysql/4a136f80-d88d-447c-bce8-4b87492110a7 | localState2 | clusterStatus2 |                      |                     |
| mysql/64bfefb0-97fd-4e34-b0fb-499ccb684faa | localState1 | clusterStatus1 |                      |                     |
| name3/3a79c040-1d3a-4583-8566-9c7097760baa | N/A - ERROR | N/A - ERROR    |                      |                     |
+--------------------------------------------+-------------+----------------+----------------------+---------------------+
`,
		))
	})

	It("prints a table with disk and cluster info grouped by instance", func() {
		buffer := new(bytes.Buffer)

		table := ui.NewTable(buffer)

		table.AddDiskInfo("mysql", "64bfefb0-97fd-4e34-b0fb-499ccb684faa", &diagagentclient.InfoResponse{
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
		table.AddDiskInfo("mysql", "4a136f80-d88d-447c-bce8-4b87492110a7", &diagagentclient.InfoResponse{
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
		table.AddDiskInfo("name3", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.AddClusterInfo("mysql", "64bfefb0-97fd-4e34-b0fb-499ccb684faa", &database.GaleraStatus{
			LocalState:    "localState1",
			ClusterSize:   1,
			ClusterStatus: "clusterStatus1",
		})
		table.AddClusterInfo("mysql", "4a136f80-d88d-447c-bce8-4b87492110a7", &database.GaleraStatus{
			LocalState:    "localState2",
			ClusterSize:   2,
			ClusterStatus: "clusterStatus2",
		})
		table.AddClusterInfo("name3", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.Render()

		Expect(buffer.String()).To(Equal(
			`+--------------------------------------------+-------------+----------------+----------------------+------------------------+
|                  INSTANCE                  |    STATE    | CLUSTER STATUS | PERSISTENT DISK USED |  EPHEMERAL DISK USED   |
+--------------------------------------------+-------------+----------------+----------------------+------------------------+
| mysql/4a136f80-d88d-447c-bce8-4b87492110a7 | localState2 | clusterStatus2 | 333B / 456B (73.0%)  | 333B / 1.4K (22.9%)    |
| mysql/64bfefb0-97fd-4e34-b0fb-499ccb684faa | localState1 | clusterStatus1 | 333B / 456B (73.0%)  | 32.5K / 142.2K (22.9%) |
| name3/3a79c040-1d3a-4583-8566-9c7097760baa | N/A - ERROR | N/A - ERROR    | N/A - ERROR          | N/A - ERROR            |
+--------------------------------------------+-------------+----------------+----------------------+------------------------+
`,
		))
	})
})
