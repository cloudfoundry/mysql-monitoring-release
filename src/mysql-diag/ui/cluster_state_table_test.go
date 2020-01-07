package ui_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/database"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/ui"
)

var _ = Describe("ClusterStateTable", func() {

	It("prints a table with some rows", func() {
		buffer := new(bytes.Buffer)

		cst := ui.NewClusterStateTable(buffer)

		cst.Add("host1", "name1", "uuid1", &database.GaleraStatus{
			LocalState:    "localState1",
			ClusterSize:   1,
			ClusterStatus: "clusterStatus1",
		})
		cst.Add("host2", "name2", "uuid2", &database.GaleraStatus{
			LocalState:    "localState2",
			ClusterSize:   2,
			ClusterStatus: "clusterStatus2",
		})
		cst.Add("host3", "name3", "uuid3", nil)

		cst.Render()

		Expect(buffer.String()).To(Equal(
			`+-------+-------------+-------------------+----------------------+--------------------+
| HOST  |  NAME/UUID  | WSREP LOCAL STATE | WSREP CLUSTER STATUS | WSREP CLUSTER SIZE |
+-------+-------------+-------------------+----------------------+--------------------+
| host1 | name1/uuid1 | localState1       | clusterStatus1       |                  1 |
| host2 | name2/uuid2 | localState2       | clusterStatus2       |                  2 |
| host3 | name3/uuid3 | N/A - ERROR       | N/A - ERROR          | N/A - ERROR        |
+-------+-------------+-------------------+----------------------+--------------------+
`,
		))
	})
})
