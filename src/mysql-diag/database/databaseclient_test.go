package database_test

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mysql-diag/database"
)

var _ = Describe("database client", func() {
	var (
		conn *sql.DB
		mock sqlmock.Sqlmock
		err  error
		ac   *database.DatabaseClient
	)

	BeforeEach(func() {
		conn, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		ac = database.NewDatabaseClient(conn)
	})

	It("returns an error when the query returns an error", func() {
		mock.ExpectQuery(`SHOW STATUS LIKE 'wsrep_%'`).WillReturnError(errors.New("some-error"))

		_, err := ac.Status()

		Expect(err).To(HaveOccurred())
	})

	It("returns an error when no rows are found (not a galera cluster)", func() {
		mock.ExpectQuery(`SHOW STATUS LIKE 'wsrep_%'`).WillReturnError(sql.ErrNoRows)

		_, err := ac.Status()

		Expect(err).To(MatchError(ContainSubstring("not a galera db")))
	})

	It("returns the galera state", func() {
		rows1 := sqlmock.NewRows([]string{"variable_name", "value"}).
			AddRow("wsrep_local_state_comment", "upsidedown").
			AddRow("wsrep_cluster_size", "3").
			AddRow("wsrep_cluster_state_uuid", "0b646f90-c164-11e6-a904-67f70a31986c").
			AddRow("wsrep_cluster_status", "Primary")
		mock.ExpectQuery(`SHOW STATUS LIKE 'wsrep_%'`).WillReturnRows(rows1)

		rows2 := sqlmock.NewRows([]string{"variable_name", "value"}).
			AddRow("read_only", "ON")
		mock.ExpectQuery(`SHOW GLOBAL VARIABLES LIKE 'read_only'`).WillReturnRows(rows2)

		status, err := ac.Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(status.LocalState).To(Equal("upsidedown"))
		Expect(status.ClusterSize).To(Equal(3))
		Expect(status.ClusterStatus).To(Equal("Primary"))
		Expect(status.ReadOnly).To(BeTrue())
	})
})
