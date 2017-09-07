package metrics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"
	 _ "github.com/go-sql-driver/mysql"

	"fmt"
	"time"
	"os"
)

var _ = Describe("mysql events", func() {
	getLeaderConnection := func() (*sql.DB, error) {
		return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/replication_monitoring?parseTime=true",
			os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), 3306))
	}
	getFollowerConnection := func() (*sql.DB, error) {
		return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/replication_monitoring?parseTime=true",
			os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_FOLLOWER_HOST"), 3306))
	}

	getLeaderJob := func() (string) {
		return os.Getenv("JOB_ID")
	}

	getFollowerJob := func() (string) {
		return os.Getenv("FOLLOWER_JOB_ID")
	}

	It("records the timestamp at some interval on the leader node", func() {
		conn, err := getLeaderConnection()

		Expect(err).NotTo(HaveOccurred())

		rows, err := conn.Query("select timestamp from replication_monitoring.heartbeat")
		Expect(err).NotTo(HaveOccurred())

		var timestamp time.Time

		Expect(rows.Next()).To(BeTrue())
		rows.Scan(&timestamp)
		Expect(rows.Next()).To(BeFalse())

		Expect(timestamp.Unix()).ToNot(Equal(0))

		Eventually(func() time.Time {
			rows, _ = conn.Query("select timestamp from replication_monitoring.heartbeat")
			var newerTimestamp time.Time

			rows.Next()
			rows.Scan(&newerTimestamp)

			return newerTimestamp
		}, "7s", "1s").Should(BeTemporally(">", timestamp))
	})

	It("inserts the current bosh job id on the leader node", func() {
		conn, err := getLeaderConnection()

		leaderJobId := getLeaderJob()

		rows, err := conn.Query("select server_id from replication_monitoring.heartbeat")
		Expect(err).NotTo(HaveOccurred())

		var serverIdFromTable string

		Expect(rows.Next()).To(BeTrue())
		rows.Scan(&serverIdFromTable)
		Expect(rows.Next()).To(BeFalse())

		Expect(serverIdFromTable).To(Equal(leaderJobId))

		Consistently(func() string {
			rows, _ := conn.Query("select server_id from replication_monitoring.heartbeat")
			var newerServerIdFromTable string

			Expect(rows.Next()).To(BeTrue())
			rows.Scan(&newerServerIdFromTable)
			Expect(rows.Next()).To(BeFalse())

			return newerServerIdFromTable
		}, "6s", "1s").Should(Equal(leaderJobId))
	})

	// it doesn't run on the follower
	It("only runs the event on the leader node (and replicates to the follower)", func() {
		Expect(getFollowerJob()).NotTo(Equal(getLeaderJob()))

		conn, err := getFollowerConnection()

		rows, err := conn.Query("select server_id from replication_monitoring.heartbeat")
		Expect(err).NotTo(HaveOccurred())

		var serverIdOnFollower string

		Expect(rows.Next()).To(BeTrue())
		rows.Scan(&serverIdOnFollower)
		Expect(rows.Next()).To(BeFalse())

		Expect(serverIdOnFollower).To(Equal(getLeaderJob()))
	})
})
