package canary_test

import (
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/cloudfoundry/replication-canary/canary"

	"errors"

	"database/sql"

	_ "github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudfoundry/replication-canary/canary/canaryfakes"
	"github.com/cloudfoundry/replication-canary/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Canary", func() {
	var (
		err           error
		fakeSQLClient *canaryfakes.FakeSQLClient
		testLogger    *lagertest.TestLogger

		canary    *Canary
		timestamp time.Time

		conn1 *sql.DB
		conn2 *sql.DB
		conn3 *sql.DB

		conns     []*models.NamedConnection
		writeConn *models.NamedConnection

		fakeHealthchecker *canaryfakes.FakeHealthchecker
	)

	BeforeEach(func() {
		fakeSQLClient = new(canaryfakes.FakeSQLClient)
		testLogger = lagertest.NewTestLogger("Canary Test Logger")
		fakeHealthchecker = new(canaryfakes.FakeHealthchecker)

		canary = NewCanary(fakeSQLClient, fakeHealthchecker, 0, testLogger)

		var err error
		// Use sqlmock to generate fake *DB, do this by hacking sql.Open
		// we know sqlmock registers itself with "sqlmock" as its drivername
		conn1, err = sql.Open("sqlmock", "conn1")
		Expect(err).NotTo(HaveOccurred())
		conn2, err = sql.Open("sqlmock", "conn2")
		Expect(err).NotTo(HaveOccurred())
		conn3, err = sql.Open("sqlmock", "conn3")
		Expect(err).NotTo(HaveOccurred())

		conns = []*models.NamedConnection{
			&models.NamedConnection{
				Name:       "conn1",
				Connection: conn1,
			},
			&models.NamedConnection{
				Name:       "conn2",
				Connection: conn2,
			},
			&models.NamedConnection{
				Name:       "conn3",
				Connection: conn3,
			},
		}

		writeConn = &models.NamedConnection{
			Name:       "conn1",
			Connection: conn1,
		}

		timestamp = time.Now()

		fakeSQLClient.CheckReturns(true, nil)
		fakeSQLClient.CleanupReturns(nil)
		fakeHealthchecker.HealthyReturns(true, nil)
	})

	AfterEach(func() {
		conn1.Close()
		conn2.Close()
		conn3.Close()
	})

	Describe("Write", func() {
		It("writes a timestamp to the connection it receives", func() {
			err = canary.Write(timestamp, writeConn)
			Expect(err).NotTo(HaveOccurred())

			expectedConn, expectedTimestamp := fakeSQLClient.WriteArgsForCall(0)
			Expect(timestamp).To(Equal(expectedTimestamp))
			Expect(conn1).To(Equal(expectedConn))
		})

		Context("when writing returns an error", func() {
			BeforeEach(func() {
				fakeSQLClient.WriteReturns(errors.New("some write error"))
			})

			It("returns the same error", func() {
				err = canary.Write(timestamp, writeConn)
				Expect(err).To(MatchError(errors.New("some write error")))
			})

		})
	})

	Describe("Chirp", func() {
		It("writes once and reads N times", func() {
			ok, err := canary.Chirp(conns, writeConn, timestamp)
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())

			// Active backend host, chosen for the Write, has Healthy
			// called once before Write, and once before Read
			// all other nodes get called once
			Expect(fakeHealthchecker.HealthyCallCount()).To(Equal(4))
			Expect(fakeHealthchecker.HealthyArgsForCall(0).Connection).To(Equal(conn1))
			Expect(fakeHealthchecker.HealthyArgsForCall(1).Connection).To(Equal(conn1))
			Expect(fakeHealthchecker.HealthyArgsForCall(2).Connection).To(Equal(conn2))
			Expect(fakeHealthchecker.HealthyArgsForCall(3).Connection).To(Equal(conn3))

			Expect(fakeSQLClient.WriteCallCount()).To(Equal(1))
			Expect(fakeSQLClient.CheckCallCount()).To(Equal(3))

			dsn, t := fakeSQLClient.WriteArgsForCall(0)
			Expect(dsn).To(Equal(conn1))
			Expect(t).To(Equal(timestamp))

			c, d := fakeSQLClient.CheckArgsForCall(0)
			Expect(c).To(Equal(conn1))
			Expect(d).To(Equal(timestamp))

			c, d = fakeSQLClient.CheckArgsForCall(1)
			Expect(c).To(Equal(conn2))
			Expect(d).To(Equal(timestamp))

			c, d = fakeSQLClient.CheckArgsForCall(2)
			Expect(c).To(Equal(conn3))
			Expect(d).To(Equal(timestamp))
		})

		It("cleans up", func() {
			_, err := canary.Chirp(conns, writeConn, timestamp)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSQLClient.CleanupCallCount()).To(Equal(1))
		})

		Context("when getting the write node's galera health status errors", func() {
			BeforeEach(func() {
				fakeHealthchecker.HealthyReturns(false, errors.New("some-error"))
			})

			It("returns (false,error) without checking", func() {
				ok, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(ok).To(BeFalse())
				Expect(err).To(MatchError(errors.New("some-error")))

				Expect(fakeSQLClient.CheckCallCount()).To(Equal(0))
			})
		})

		Context("when the write node's galera is unhealthy", func() {
			BeforeEach(func() {
				fakeHealthchecker.HealthyReturns(false, nil)
			})

			It("returns (false,error) without checking", func() {
				ok, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(ok).To(BeFalse())
				Expect(err).To(HaveOccurred())

				Expect(fakeSQLClient.CheckCallCount()).To(Equal(0))
			})
		})

		Context("when the write errors", func() {
			BeforeEach(func() {
				fakeSQLClient.WriteReturns(errors.New("some write error"))
			})

			It("returns (false,error)", func() {
				ok, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(ok).To(BeFalse())
				Expect(err).To(MatchError(errors.New("some write error")))

				// Verify that even if we bail early when writing fails, we did at least once
				// ask to see if it was healthy
				Expect(fakeHealthchecker.HealthyCallCount()).To(Equal(1))
			})
		})

		Context("when a node's galera is unhealthy", func() {
			BeforeEach(func() {
				fakeHealthchecker.HealthyStub = func(db *models.NamedConnection) (bool, error) {
					return (db.Name != "conn2"), nil
				}
			})

			It("is not used in determining replication status", func() {
				_, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(err).NotTo(HaveOccurred())

				// Only checks 2 instead of the 3
				Expect(fakeSQLClient.CheckCallCount()).To(Equal(2))

				expectedConn1, _ := fakeSQLClient.CheckArgsForCall(0)
				Expect(expectedConn1).To(Equal(conn1))

				// Skip checking conn2
				expectedConn2, _ := fakeSQLClient.CheckArgsForCall(1)
				Expect(expectedConn2).To(Equal(conn3))
			})
		})

		Context("when checking a node's galera status errors", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("galera check error")
				fakeHealthchecker.HealthyStub = func(db *models.NamedConnection) (bool, error) {
					if db.Name == "conn2" {
						return false, expectedErr
					}

					return true, nil
				}
			})

			It("is not used in determining replication status", func() {
				_, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(err).To(Equal(expectedErr))

				// Only checks 2 instead of the 3
				Expect(fakeSQLClient.CheckCallCount()).To(Equal(2))

				expectedConn1, _ := fakeSQLClient.CheckArgsForCall(0)
				Expect(expectedConn1).To(Equal(conn1))

				// Skip checking conn2
				expectedConn2, _ := fakeSQLClient.CheckArgsForCall(1)
				Expect(expectedConn2).To(Equal(conn3))
			})
		})

		Context("when all checks are non-deterministic", func() {
			BeforeEach(func() {
				var i int

				fakeSQLClient.CheckStub = func(*sql.DB, time.Time) (bool, error) {
					i++
					if i == 1 {
						return false, errors.New("first error")
					} else {
						return false, errors.New("second error")
					}
				}
			})

			It("returns (false,error) where the error is the first error", func() {
				ok, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(ok).To(BeFalse())
				Expect(err).To(MatchError(errors.New("first error")))
			})

			It("cleans up", func() {
				_, _ = canary.Chirp(conns, writeConn, timestamp)

				Expect(fakeSQLClient.CleanupCallCount()).To(Equal(1))
			})

			Context("when the cleanup fails", func() {
				BeforeEach(func() {
					fakeSQLClient.CleanupReturns(errors.New("cleanup err"))
				})

				It("returns the first error", func() {
					ok, err := canary.Chirp(conns, writeConn, timestamp)
					Expect(ok).To(BeFalse())
					Expect(err).To(MatchError(errors.New("first error")))
				})
			})
		})

		Context("when the first check is non-deterministic, but the second is a deterministic failure", func() {
			BeforeEach(func() {
				var i int
				fakeSQLClient.CheckStub = func(*sql.DB, time.Time) (bool, error) {
					i++
					if i == 1 {
						return false, errors.New("first error")
					} else {
						return false, nil
					}
				}
			})

			It("returns (false,nil)", func() {
				ok, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(ok).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})

			It("cleans up", func() {
				_, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeSQLClient.CleanupCallCount()).To(Equal(1))
			})

			Context("when the cleanup fails", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("cleanup err")
					fakeSQLClient.CleanupReturns(expectedErr)
				})

				It("returns the cleanup error", func() {
					ok, err := canary.Chirp(conns, writeConn, timestamp)
					Expect(ok).To(BeFalse())
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})

		Context("when the cleanup is unsuccessful", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some cleanup error")
				fakeSQLClient.CleanupReturns(expectedErr)
			})

			It("returns (false,error) where error is the cleanup error", func() {
				ok, err := canary.Chirp(conns, writeConn, timestamp)
				Expect(ok).To(BeFalse())
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
