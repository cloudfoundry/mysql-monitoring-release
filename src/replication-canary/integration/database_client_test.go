package integration_test

import (
	"database/sql"
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	_ "github.com/go-sql-driver/mysql"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/replication-canary/database"
)

var _ = Describe("Database", func() {
	var (
		db  *sql.DB
		err error

		testLogger *lagertest.TestLogger
		client     *database.Client
	)

	BeforeEach(func() {
		db, err = sql.Open("mysql", databaseDSN)
		Expect(err).NotTo(HaveOccurred())

		testLogger = lagertest.NewTestLogger("database integration test")

		sessionVariables := make(map[string]string)
		sessionVariables["sql_log_off"] = "0"
		client = database.NewClient(sessionVariables, testLogger)
	})

	AfterEach(func() {
		defer db.Close()

		_, err = db.Exec("TRUNCATE TABLE chirps")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Cleanup", func() {
		BeforeEach(func() {
			err = client.Setup(db)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keeps any entries less than 4,320", func() {
			someData := "foo"

			_, err = db.Exec("INSERT INTO chirps (data) VALUES (?)", someData)
			Expect(err).NotTo(HaveOccurred())

			client.Cleanup(db)

			var count int
			err = db.QueryRow("SELECT count(*) FROM chirps").Scan(&count)
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(Equal(1))
		})

		It("keeps only the last 4,320 entries by id DESC", func() {
			for i := 0; i < 4322; i++ {
				someData := fmt.Sprintf("foo-%d", i)

				_, err = db.Exec("INSERT INTO chirps (data) VALUES (?)", someData)
				Expect(err).NotTo(HaveOccurred())
			}

			client.Cleanup(db)

			var count int
			err = db.QueryRow("SELECT count(*) FROM chirps").Scan(&count)
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(Equal(4320))

			rows, err := db.Query("SELECT data FROM chirps")
			Expect(err).NotTo(HaveOccurred())

			var datas []string
			defer rows.Close()
			for rows.Next() {
				var data string
				err = rows.Scan(&data)
				Expect(err).NotTo(HaveOccurred())
				datas = append(datas, data)
			}

			Expect(rows.Err()).NotTo(HaveOccurred())

			// Keep last 4320 of them
			var expectedDatas = []string{}
			for i := 2; i < 4322; i++ {
				expectedDatas = append(expectedDatas, fmt.Sprintf("foo-%d", i))
			}

			Expect(datas).To(Equal(expectedDatas))
		})
	})

	Describe("Check", func() {
		BeforeEach(func() {
			err = client.Setup(db)
			Expect(err).NotTo(HaveOccurred())
		})

		It("verifies that data exists in the chirps table", func() {
			existingData := time.Now()

			_, err = db.Exec("INSERT INTO chirps (data) VALUES (?)", existingData.String())
			Expect(err).NotTo(HaveOccurred())

			ok, err := client.Check(db, existingData)
			Expect(err).NotTo(HaveOccurred())

			Expect(ok).To(BeTrue())
		})

		It("fails if the data does not exist in the chirps table", func() {
			_, err = db.Exec("TRUNCATE TABLE chirps")
			Expect(err).NotTo(HaveOccurred())

			ok, err := client.Check(db, time.Now())

			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeFalse())
		})
	})

	Describe("Write", func() {
		BeforeEach(func() {
			err = client.Setup(db)
			Expect(err).NotTo(HaveOccurred())
		})

		It("inserts the data into the table", func() {
			writtenData := time.Now()
			client.Write(db, writtenData)

			var data string

			err = db.QueryRow("SELECT data FROM chirps").Scan(&data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(Equal(writtenData.String()))
		})
	})
})
