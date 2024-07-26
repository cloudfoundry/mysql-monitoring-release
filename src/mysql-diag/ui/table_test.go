package ui_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/ui"
)

var _ = Describe("Table", func() {
	format.TruncatedDiff = false

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

			`+-------------------------------------------------+-------+----------------+-------+----------------------+------------------------+
|                    INSTANCE                     | STATE | CLUSTER STATUS | SEQNO | PERSISTENT DISK USED |  EPHEMERAL DISK USED   |
+-------------------------------------------------+-------+----------------+-------+----------------------+------------------------+
|  [0] mysql/3a79c040-1d3a-4583-8566-9c7097760baa |       |                |       | N/A - ERROR          | N/A - ERROR            |
|  [1] mysql/4a136f80-d88d-447c-bce8-4b87492110a7 |       |                |       | 333B / 456B (73.0%)  | 333B / 1.4K (22.9%)    |
|  [2] mysql/64bfefb0-97fd-4e34-b0fb-499ccb684faa |       |                |       | 333B / 456B (73.0%)  | 32.5K / 142.2K (22.9%) |
+-------------------------------------------------+-------+----------------+-------+----------------------+------------------------+
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
			LocalIndex:    "befe0c28-b5f4",
			LastApplied:   10,
		})
		table.AddClusterInfo("mysql", "4a136f80-d88d-447c-bce8-4b87492110a7", &database.GaleraStatus{
			LocalState:    "localState2",
			ClusterSize:   2,
			ClusterStatus: "clusterStatus2",
			LocalIndex:    "4375e0a0-a811",
			LastApplied:   11,
		})

		table.AddClusterInfo("mysql", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.Render()

		Expect(buffer.String()).To(Equal(
			`+-------------------------------------------------+-------------+----------------+-------+----------------------+---------------------+
|                    INSTANCE                     |    STATE    | CLUSTER STATUS | SEQNO | PERSISTENT DISK USED | EPHEMERAL DISK USED |
+-------------------------------------------------+-------------+----------------+-------+----------------------+---------------------+
|  [0] mysql/4a136f80-d88d-447c-bce8-4b87492110a7 | localState2 | clusterStatus2 |    11 |                      |                     |
|  [1] mysql/64bfefb0-97fd-4e34-b0fb-499ccb684faa | localState1 | clusterStatus1 |    10 |                      |                     |
|  [2] mysql/3a79c040-1d3a-4583-8566-9c7097760baa | N/A - ERROR | N/A - ERROR    |    -1 |                      |                     |
+-------------------------------------------------+-------------+----------------+-------+----------------------+---------------------+
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
		table.AddDiskInfo("mysql", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.AddClusterInfo("mysql", "64bfefb0-97fd-4e34-b0fb-499ccb684faa", &database.GaleraStatus{
			LocalState:    "localState1",
			ClusterSize:   1,
			ClusterStatus: "clusterStatus1",
			LocalIndex:    "4375e0a0-a811",
			LastApplied:   7,
		})
		table.AddClusterInfo("mysql", "4a136f80-d88d-447c-bce8-4b87492110a7", &database.GaleraStatus{
			LocalState:    "localState2",
			ClusterSize:   2,
			ClusterStatus: "clusterStatus2",
			LocalIndex:    "befe0c28-b5f4",
			LastApplied:   5,
		})
		table.AddClusterInfo("mysql", "3a79c040-1d3a-4583-8566-9c7097760baa", nil)

		table.Render()

		Expect(buffer.String()).To(Equal(
			`+-------------------------------------------------+-------------+----------------+-------+----------------------+------------------------+
|                    INSTANCE                     |    STATE    | CLUSTER STATUS | SEQNO | PERSISTENT DISK USED |  EPHEMERAL DISK USED   |
+-------------------------------------------------+-------------+----------------+-------+----------------------+------------------------+
|  [0] mysql/64bfefb0-97fd-4e34-b0fb-499ccb684faa | localState1 | clusterStatus1 |     7 | 333B / 456B (73.0%)  | 32.5K / 142.2K (22.9%) |
|  [1] mysql/4a136f80-d88d-447c-bce8-4b87492110a7 | localState2 | clusterStatus2 |     5 | 333B / 456B (73.0%)  | 333B / 1.4K (22.9%)    |
|  [2] mysql/3a79c040-1d3a-4583-8566-9c7097760baa | N/A - ERROR | N/A - ERROR    |    -1 | N/A - ERROR          | N/A - ERROR            |
+-------------------------------------------------+-------------+----------------+-------+----------------------+------------------------+
`,
		))
	})
})
