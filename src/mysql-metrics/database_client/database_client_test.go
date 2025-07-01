package database_client_test

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configPackage "github.com/cloudfoundry/mysql-metrics/config"
	"github.com/cloudfoundry/mysql-metrics/database_client"
)

var _ = Describe("DatabaseClient", func() {
	var (
		conn             *sql.DB
		mock             sqlmock.Sqlmock
		err              error
		dc               *database_client.DbClient
		config           *configPackage.Config
		hbDatabase       string
		hbTable          string
		hbTableQualified string
	)

	BeforeEach(func() {
		conn, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		hbDatabase = "someHbDatabase"
		hbTable = "someHbTable"
		hbTableQualified = database_client.QuoteIdentifier(hbDatabase) +
			"." + database_client.QuoteIdentifier(hbTable)

		config = &configPackage.Config{
			HeartbeatDatabase: hbDatabase,
			HeartbeatTable:    hbTable,
		}

		dc = database_client.NewDatabaseClient(conn, config)
	})

	Describe("IsAvailable", func() {
		It("returns true when executing a test query against the DB succeeds", func() {
			row := sqlmock.NewRows([]string{"variable_name", "value"}).
				AddRow("something", "doesntmatter")
			mock.ExpectQuery(`SHOW GLOBAL STATUS`).WillReturnRows(row)

			result := dc.IsAvailable()
			Expect(result).To(BeTrue())
		})

		It("returns false when executing a test query against the DB results in an error", func() {
			mock.ExpectQuery(`SHOW GLOBAL STATUS`).WillReturnError(errors.New("db unavailable"))

			result := dc.IsAvailable()
			Expect(result).To(BeFalse())
		})
	})

	Describe("ShowGlobalStatus", func() {
		It("returns a map of lowercased status variables and their values", func() {
			row := sqlmock.NewRows([]string{"variable_name", "value"}).
				AddRow("something", "doesntmatter").
				AddRow("Other Thing", "123.4")
			mock.ExpectQuery(`SHOW GLOBAL STATUS`).WillReturnRows(row)

			status, err := dc.ShowGlobalStatus()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(status)).To(Equal(2))
			Expect(status).To(Equal(map[string]string{
				"something":   "doesntmatter",
				"other thing": "123.4",
			}))
		})

		It("returns an error and no data when the query fails", func() {
			mock.ExpectQuery(`SHOW GLOBAL STATUS`).WillReturnError(errors.New("db unavailable"))
			_, err := dc.ShowGlobalStatus()
			Expect(err).To(MatchError("db unavailable"))
		})
	})

	Describe("ShowGlobalVariables", func() {
		It("returns a map of lowercased status variables and their values", func() {
			row := sqlmock.NewRows([]string{"variable_name", "value"}).
				AddRow("Max_connections", "12")
			mock.ExpectQuery(`SHOW GLOBAL VARIABLES`).WillReturnRows(row)

			vars, err := dc.ShowGlobalVariables()
			Expect(err).NotTo(HaveOccurred())
			Expect(vars).To(Equal(map[string]string{
				"max_connections": "12",
			}))
		})

		It("returns an error and no data when the query fails", func() {
			mock.ExpectQuery(`SHOW GLOBAL VARIABLES`).WillReturnError(errors.New("db unavailable"))
			_, err := dc.ShowGlobalVariables()
			Expect(err).To(MatchError("db unavailable"))
		})
	})

	Describe("ShowReplicaStatus", func() {
		It("returns an empty map when SHOW REPLICA STATUS returns an empty result (leader node)", func() {
			mock.ExpectQuery("SHOW REPLICA STATUS").WillReturnRows(sqlmock.NewRows([]string{}))

			vars, err := dc.ShowReplicaStatus()
			Expect(err).NotTo(HaveOccurred())
			Expect(vars).To(Equal(map[string]string{}))
		})

		It("returns a map of lowercased slave status variables and their values when SHOW REPLICA STATUS returns a non-empty result", func() {
			row := sqlmock.NewRows([]string{
				"Some_Slave_Status_Metric",
				"Slave_SQL_Running_State",
				"SQL_Lib_Not_Able_To_Parse",
			}).AddRow(
				[]uint8("123.4"),
				nil,
				[]uint8(nil),
			)
			mock.ExpectQuery("SHOW REPLICA STATUS").WillReturnRows(row)

			vars, err := dc.ShowReplicaStatus()
			Expect(err).NotTo(HaveOccurred())
			Expect(vars).To(Equal(map[string]string{
				"some_slave_status_metric":  "123.4",
				"slave_sql_running_state":   "NULL",
				"sql_lib_not_able_to_parse": "",
			}))
		})

		It("returns an error and no data when the query fails", func() {
			mock.ExpectQuery(`SHOW REPLICA STATUS`).WillReturnError(errors.New("db unavailable"))
			_, err := dc.ShowReplicaStatus()
			Expect(err).To(MatchError("db unavailable"))
		})
	})

	Describe("ServicePlansDiskAllocated", func() {
		It("returns a 0 when no service instances have been provisioned", func() {
			row := sqlmock.NewRows([]string{
				"service_plans_disk_allocated",
			}).AddRow(
				nil,
			)
			mock.ExpectQuery("SELECT SUM\\(max_storage_mb\\) AS service_plans_disk_allocated FROM mysql_broker\\.service_instances").WillReturnRows(row)

			vars, err := dc.ServicePlansDiskAllocated()
			Expect(err).NotTo(HaveOccurred())
			Expect(vars).To(Equal(map[string]string{
				"service_plans_disk_allocated": "0",
			}))
		})

		It("returns the sum of all provisioned service instance plans", func() {
			row := sqlmock.NewRows([]string{
				"service_plans_disk_allocated",
			}).AddRow(
				[]uint8("200"),
			)
			mock.ExpectQuery("SELECT SUM\\(max_storage_mb\\) AS service_plans_disk_allocated FROM mysql_broker\\.service_instances").WillReturnRows(row)

			vars, err := dc.ServicePlansDiskAllocated()
			Expect(err).NotTo(HaveOccurred())
			Expect(vars).To(Equal(map[string]string{
				"service_plans_disk_allocated": "200",
			}))
		})

		It("returns an error and no data when the query fails", func() {
			mock.ExpectQuery("SELECT SUM\\(max_storage_mb\\) AS service_plans_disk_allocated FROM mysql_broker\\.service_instances").WillReturnError(errors.New("db unavailable"))
			_, err := dc.ServicePlansDiskAllocated()
			Expect(err).To(MatchError("db unavailable"))
		})
	})

	Describe("IsFollower", func() {
		It("returns true when the node is a follower", func() {
			rows := sqlmock.NewRows([]string{
				"SomeMasterSlaveStatusThing",
			}).AddRow(
				[]uint8("foobar"),
			)
			mock.ExpectQuery("SHOW REPLICA STATUS").WillReturnRows(rows)

			isFollower, err := dc.IsFollower()
			Expect(isFollower).To(BeTrue())
			Expect(err).To(BeNil())
		})

		It("returns false when the node is a leader or not in leader follower mode", func() {
			rows := sqlmock.NewRows([]string{})
			mock.ExpectQuery("SHOW REPLICA STATUS").WillReturnRows(rows)

			isFollower, err := dc.IsFollower()
			Expect(isFollower).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("returns an error when the query fails", func() {
			mock.ExpectQuery(`SHOW REPLICA STATUS`).WillReturnError(errors.New("db unavailable"))
			_, err := dc.IsFollower()
			Expect(err).To(MatchError("db unavailable"))
		})
	})

	Describe("HeartbeatStatus", func() {
		It("returns a map containing seconds_since_leader_heartbeat", func() {
			row := sqlmock.NewRows([]string{
				"seconds_since_leader_heartbeat",
			}).AddRow(
				[]uint8("3"),
			)

			mock.ExpectQuery(
				`SELECT UNIX_TIMESTAMP\(NOW\(\)\) - UNIX_TIMESTAMP\(timestamp\) AS seconds_since_leader_heartbeat FROM ` +
					hbTableQualified).
				WillReturnRows(row)

			status, err := dc.HeartbeatStatus()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(status)).To(Equal(1))
			Expect(status).To(Equal(map[string]string{
				"seconds_since_leader_heartbeat": "3",
			}))
		})

		It("returns an error and no data when the query fails", func() {
			mock.ExpectQuery(
				`SELECT UNIX_TIMESTAMP\(NOW\(\)\) - UNIX_TIMESTAMP\(timestamp\) AS seconds_since_leader_heartbeat FROM ` + hbTableQualified).
				WillReturnError(errors.New("db unavailable"))
			_, err := dc.HeartbeatStatus()
			Expect(err).To(MatchError("db unavailable"))
		})
	})

	Describe("QuoteIdentifier", func() {
		It("quotes identifier while escaping existing quotes", func() {
			Expect(database_client.QuoteIdentifier("foobar")).To(Equal("`foobar`"))
			Expect(database_client.QuoteIdentifier("bar`baz")).To(Equal("`bar``baz`"))
		})
	})

	Describe("FindLastBackupTimestamp", func() {
		Context("when there is a timestamp", func() {
			It("returns the timestamp", func() {
				parsedTime, _ := time.Parse("2006-01-02 15:04:05", "2020-07-14 21:28:16.000000")

				row := sqlmock.NewRows([]string{"timestamp"}).AddRow([]uint8("2020-07-14 21:28:16.000000"))
				mock.ExpectQuery(`SELECT ts AS timestamp FROM backup_metrics.backup_times ORDER BY ts DESC LIMIT 1`).WillReturnRows(row)

				vars, err := dc.FindLastBackupTimestamp()
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal(parsedTime))
			})
		})

		Context("when there is no timestamp", func() {
			It("returns an empty timestamp", func() {
				mock.ExpectQuery("SELECT ts AS timestamp FROM backup_metrics.backup_times ORDER BY ts DESC LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{}))

				vars, err := dc.FindLastBackupTimestamp()
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal(time.Time{}))
			})
		})

		Context("when the timestamp is invalid", func() {
			It("returns an error", func() {
				mock.ExpectQuery("SELECT ts AS timestamp FROM backup_metrics.backup_times ORDER BY ts DESC LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow([]uint8("bad time")))

				_, err := dc.FindLastBackupTimestamp()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`cannot parse "bad time"`))
			})
		})

		Context("when the query fails", func() {
			It("returns an error and no data", func() {
				mock.ExpectQuery("SELECT ts AS timestamp FROM backup_metrics.backup_times ORDER BY ts DESC LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{})).
					WillReturnError(errors.New("db unavailable"))
				_, err := dc.FindLastBackupTimestamp()
				Expect(err).To(MatchError("db unavailable"))
			})
		})
	})
})
