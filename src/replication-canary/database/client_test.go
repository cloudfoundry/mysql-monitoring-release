package database_test

import (
	dsql "database/sql"
	"errors"
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudfoundry/replication-canary/database"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {

	const (
		rowCountQuery = `^SELECT COUNT\(id\) AS row_count FROM chirps$`
		deleteQuery   = `^DELETE FROM chirps ORDER BY id ASC LIMIT \?$`
	)

	var (
		testLogger *lagertest.TestLogger
		client     *database.Client

		timestamp time.Time
	)

	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("database client test")

		sessionVariables := make(map[string]string)
		sessionVariables["wsrep_sync_wait"] = "1"
		client = database.NewClient(sessionVariables, testLogger)

		timestamp = time.Now()
	})

	Describe("Setup", func() {
		It("sets up the table if it does not exist", func() {
			db, mock, err := sqlmock.New()
			Expect(err).NotTo(HaveOccurred())

			defer db.Close()

			mock.ExpectExec(`CREATE TABLE IF NOT EXISTS chirps \(id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, data VARCHAR\(255\) NOT NULL\) ENGINE=InnoDB`).WillReturnResult(sqlmock.NewResult(1, 1))

			err = client.Setup(db)
			Expect(err).NotTo(HaveOccurred())

			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
		})
	})

	Describe("Write", func() {
		It("writes to the connection", func() {
			db, mock, err := sqlmock.New()
			Expect(err).NotTo(HaveOccurred())

			defer db.Close()

			mock.ExpectExec(`INSERT INTO chirps \(data\)`).WithArgs(timestamp.String()).WillReturnResult(sqlmock.NewResult(1, 1))

			err = client.Write(db, timestamp)
			Expect(err).NotTo(HaveOccurred())

			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
		})
	})

	Describe("Check", func() {
		var (
			db   *dsql.DB
			mock sqlmock.Sqlmock
			err  error
		)

		BeforeEach(func() {
			db, mock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())

			db.Close()
		})

		It("verifies the data exists on that connection", func() {
			mock.ExpectBegin()
			mock.ExpectExec(`SET SESSION wsrep_sync_wait=1`).WillReturnResult(sqlmock.NewResult(1, 1))
			rows := sqlmock.NewRows([]string{"data"}).AddRow(timestamp.String())
			mock.ExpectQuery(`SELECT data FROM chirps WHERE data =`).WithArgs(timestamp.String()).WillReturnRows(rows)
			mock.ExpectRollback()

			_, err = client.Check(db, timestamp)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error if there's an error while setting the session variable", func() {
			dbErr := errors.New("something")
			mock.ExpectBegin()
			mock.ExpectExec(`SET SESSION wsrep_sync_wait=1`).WillReturnError(dbErr)
			mock.ExpectRollback()

			ret, err := client.Check(db, timestamp)
			Expect(ret).To(BeFalse())
			Expect(err).To(MatchError(dbErr))
		})

		It("returns error if there's an error in the query", func() {
			dbErr := errors.New("something")
			mock.ExpectBegin()
			mock.ExpectExec(`SET SESSION wsrep_sync_wait=1`).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(`SELECT data FROM chirps WHERE data =`).WithArgs(timestamp.String()).WillReturnError(dbErr)
			mock.ExpectRollback()

			_, err = client.Check(db, timestamp)
			Expect(err).To(MatchError(dbErr))
		})

		It("returns false if the row returned does not contain the correct data", func() {
			mock.ExpectBegin()
			mock.ExpectExec(`SET SESSION wsrep_sync_wait=1`).WillReturnResult(sqlmock.NewResult(1, 1))
			rows := sqlmock.NewRows([]string{"data"}).AddRow("some-other-data")
			mock.ExpectQuery(`SELECT data FROM chirps WHERE data =`).WithArgs(timestamp.String()).WillReturnRows(rows)
			mock.ExpectRollback()

			ok, err := client.Check(db, timestamp)
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeFalse())
		})

		It("swallows sql.ErrNoRow", func() {
			mock.ExpectBegin()
			mock.ExpectExec(`SET SESSION wsrep_sync_wait=1`).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(`SELECT data FROM chirps WHERE data =`).WithArgs(timestamp.String()).WillReturnError(dsql.ErrNoRows)
			mock.ExpectRollback()

			ok, err := client.Check(db, timestamp)
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeFalse())
		})

		It("returns true if the row returned contains the correct data", func() {
			mock.ExpectBegin()
			mock.ExpectExec(`SET SESSION wsrep_sync_wait=1`).WillReturnResult(sqlmock.NewResult(1, 1))
			rows := sqlmock.NewRows([]string{"data"}).AddRow(timestamp.String())
			mock.ExpectQuery(`SELECT data FROM chirps WHERE data =`).WithArgs(timestamp.String()).WillReturnRows(rows)
			mock.ExpectRollback()

			ok, err := client.Check(db, timestamp)
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
	})

	Describe("Cleanup", func() {
		var (
			db   *dsql.DB
			mock sqlmock.Sqlmock
			err  error
		)

		BeforeEach(func() {
			db, mock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			db.Close()
		})

		Context("when the number of rows exceeds 4,320", func() {
			It("deletes any rows not in the most recent 4,320", func() {
				rows := sqlmock.NewRows([]string{"row_count"}).AddRow(5000)

				mock.ExpectBegin()
				mock.ExpectQuery(rowCountQuery).WillReturnRows(rows)

				prepare := mock.ExpectPrepare(deleteQuery)
				prepare.ExpectExec().WithArgs(680).WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectCommit()

				err := client.Cleanup(db)

				Expect(err).NotTo(HaveOccurred())
				Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
			})
		})
		Context("when the number of rows is fewer than 4,320", func() {
			It("does not attempt to delete rows", func() {
				rows := sqlmock.NewRows([]string{"row_count"}).AddRow(4000)

				mock.ExpectBegin()
				mock.ExpectQuery(rowCountQuery).WillReturnRows(rows)

				prepare := mock.ExpectPrepare(deleteQuery)
				prepare.ExpectExec()

				err := client.Cleanup(db)

				Expect(err).NotTo(HaveOccurred())
				Expect(mock.ExpectationsWereMet()).To(HaveOccurred())
			})
		})

		It("performs a transaction rollback if there is an error", func() {
			mock.ExpectBegin()
			mock.ExpectQuery(rowCountQuery).WillReturnError(errors.New("some exec error"))
			mock.ExpectRollback()

			err := client.Cleanup(db)

			Expect(err).To(HaveOccurred())
			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
		})

		Context("error handling", func() {
			AfterEach(func() {
				Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
			})
			Context("when creating a transaction fails", func() {
				It("returns an error", func() {
					expectedErr := errors.New("some error")

					mock.ExpectBegin().WillReturnError(expectedErr)

					err := client.Cleanup(db)
					Expect(err).To(Equal(expectedErr))
				})
			})
			Context("when the select query fails", func() {
				It("returns an error", func() {
					expectedErr := errors.New("some error")

					mock.ExpectBegin()
					mock.ExpectQuery(rowCountQuery).WillReturnError(expectedErr)
					mock.ExpectRollback()

					err := client.Cleanup(db)
					Expect(err).To(Equal(expectedErr))
				})
			})
			Context("when scanning the result of the select fails", func() {
				It("returns an error", func() {
					rows := sqlmock.NewRows([]string{"row_count"}).AddRow("string")

					mock.ExpectBegin()
					mock.ExpectQuery(rowCountQuery).WillReturnRows(rows)
					mock.ExpectRollback()

					err := client.Cleanup(db)
					Expect(err).To(HaveOccurred())
				})
			})
			Context("when rows.Err fails", func() {
				It("returns an error", func() {
					expectedErr := errors.New("some error")
					rows := sqlmock.NewRows([]string{"row_count"}).AddRow(100)

					mock.ExpectBegin()
					mock.ExpectQuery(rowCountQuery).WillReturnRows(rows)
					rows.RowError(0, expectedErr)
					mock.ExpectRollback()

					err := client.Cleanup(db)
					Expect(err).To(Equal(expectedErr))
				})
			})
			Context("when preparing the delete query fails", func() {
				It("returns an error", func() {
					expectedErr := errors.New("some error")
					rows := sqlmock.NewRows([]string{"row_count"}).AddRow(5000)

					mock.ExpectBegin()
					mock.ExpectQuery(rowCountQuery).WillReturnRows(rows)
					mock.ExpectPrepare(deleteQuery).WillReturnError(expectedErr)
					mock.ExpectRollback()

					err := client.Cleanup(db)
					Expect(err).To(Equal(expectedErr))
				})
			})
			Context("when executing the delete query fails", func() {
				It("returns an error", func() {
					expectedErr := errors.New("some error")
					rows := sqlmock.NewRows([]string{"row_count"}).AddRow(5000)

					mock.ExpectBegin()
					mock.ExpectQuery(rowCountQuery).WillReturnRows(rows)
					prepare := mock.ExpectPrepare(deleteQuery)
					prepare.ExpectExec().WithArgs(680).WillReturnError(expectedErr)
					mock.ExpectRollback()

					err := client.Cleanup(db)
					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
