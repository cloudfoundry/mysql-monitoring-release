package ui_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/ui"
)

var _ = Describe("ClusterStateTable", func() {

	It("prints a table with some rows", func() {
		buffer := new(bytes.Buffer)

		cst := ui.NewClusterStateTable(buffer)

		cst.Add("mysql", "uuid1", &database.GaleraStatus{
			LocalState:    "localState1",
			ClusterSize:   1,
			ClusterStatus: "clusterStatus1",
		})
		cst.Add("mysql", "uuid2", &database.GaleraStatus{
			LocalState:    "localState2",
			ClusterSize:   2,
			ClusterStatus: "clusterStatus2",
		})
		cst.Add("name3", "uuid3", nil)

		cst.Render()

		Expect(buffer.String()).To(Equal(
			`+-------------+-------------+----------------+
|  INSTANCE   |    STATE    | CLUSTER STATUS |
+-------------+-------------+----------------+
| mysql/uuid1 | localState1 | clusterStatus1 |
| mysql/uuid2 | localState2 | clusterStatus2 |
| name3/uuid3 | N/A - ERROR | N/A - ERROR    |
+-------------+-------------+----------------+
`,
		))
	})
})
