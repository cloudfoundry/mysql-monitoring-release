package galera_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "replication-canary/galera"
	"replication-canary/models"
)

var _ = Describe("Client", func() {
	Describe("Healthy", func() {
		var (
			conn *sql.DB
			mock sqlmock.Sqlmock
			err  error

			client *Client
			logger lager.Logger

			namedConn *models.NamedConnection
		)

		BeforeEach(func() {
			conn, mock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())

			logger = lagertest.NewTestLogger("galera client")

			namedConn = &models.NamedConnection{
				Connection: conn,
			}

			client = &Client{
				Logger: logger,
			}
		})

		It("returns an error when the query returns an error", func() {
			mock.ExpectQuery(`SHOW STATUS LIKE 'wsrep_local_state'`).WillReturnError(errors.New("some-error"))

			_, err := client.Healthy(namedConn)

			Expect(err).To(HaveOccurred())
			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
		})

		It("returns an error when no rows are found (not a galera cluster)", func() {
			mock.ExpectQuery(`SHOW STATUS LIKE 'wsrep_local_state'`).WillReturnError(sql.ErrNoRows)

			_, err := client.Healthy(namedConn)

			Expect(err).To(MatchError(ContainSubstring("not a galera db")))
			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
		})

		DescribeTable("states",
			func(stateValue int, healthy bool) {
				rows := sqlmock.NewRows([]string{"variable_name", "value"}).AddRow("wsrep_local_state", stateValue)

				mock.ExpectQuery(`SHOW STATUS LIKE 'wsrep_local_state'`).WillReturnRows(rows)

				status, err := client.Healthy(namedConn)

				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(healthy))
				Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
			},
			Entry("Joining", 1, false),
			Entry("Donor/Desynced", 2, false),
			Entry("Joined", 3, false),
			Entry("Synced", 4, true),
			Entry("0", 0, false),
			Entry("-1", -1, false),
			Entry("5", 5, false),
		)
	})
})
