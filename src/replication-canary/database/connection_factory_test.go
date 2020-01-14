package database_test

import (
	"errors"

	"code.cloudfoundry.org/lager/lagertest"

	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/cloudfoundry/replication-canary/database"
	"github.com/cloudfoundry/replication-canary/database/databasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/replication-canary/config"
)

var _ = Describe("Connection Factory", func() {
	var (
		fakeSwitchboardClients []SwitchboardClient
		fakeSwitchboardClient  *databasefakes.FakeSwitchboardClient
		testLogger             *lagertest.TestLogger
		connectionFactory      *ConnectionFactory

		dsns []string

		db *sql.DB
	)

	BeforeEach(func() {
		dsns = []string{
			"fake_username:fake_password@tcp(192.0.2.20:3306)/fake_db",
			"fake_username:fake_password@tcp(192.0.2.2:3306)/fake_db",
		}
		fakeSwitchboardClient = &databasefakes.FakeSwitchboardClient{}
		fakeSwitchboardClients = []SwitchboardClient{fakeSwitchboardClient}

		testLogger = lagertest.NewTestLogger("database connection factory test")

		config := &config.Config{
			MySQL: config.MySQL{
				ClusterIPs: []string{"192.0.2.20", "192.0.2.2"},
				Port:       3306,
			},
			Canary: config.Canary{
				Database: "fake_db",
				Username: "fake_username",
				Password: "fake_password",
			},
		}

		connectionFactory = NewConnectionFactoryFromConfig(
			config,
			fakeSwitchboardClients,
			testLogger,
		)

		var err error
		db, _, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		connectionFactory.OpenConn = func(dsn string) (*sql.DB, error) {
			return db, nil
		}
	})
	AfterEach(func() {
		db.Close()
	})

	Describe("Conns", func() {
		Context("when connecting to MariaDB", func() {
			It("uses the correct connection string", func() {
				var connectedDSNs []string
				connectionFactory.OpenConn = func(dsn string) (*sql.DB, error) {
					connectedDSNs = append(connectedDSNs, dsn)
					return db, nil
				}
				_, err := connectionFactory.Conns()
				Expect(err).NotTo(HaveOccurred())
				Expect(connectedDSNs).To(Equal(dsns))
			})
		})
	})

	Describe("WriteConn", func() {
		BeforeEach(func() {
			fakeSwitchboardClient.ActiveBackendHostReturns("192.0.2.2", nil)
		})

		It("asks the switchboard client for the active backend host", func() {
			_, err := connectionFactory.WriteConn()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSwitchboardClient.ActiveBackendHostCallCount()).To(Equal(1))
		})

		It("returns the connection that corresponds to the host backend", func() {
			allConns, err := connectionFactory.Conns()
			Expect(err).NotTo(HaveOccurred())

			conn, err := connectionFactory.WriteConn()
			Expect(err).NotTo(HaveOccurred())

			Expect(conn).To(Equal(allConns[1]))
		})

		Context("when the switchboard client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some switchboard error")

				fakeSwitchboardClient.ActiveBackendHostReturns("", expectedErr)
			})

			It("returns the error", func() {
				_, err := connectionFactory.WriteConn()
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})
