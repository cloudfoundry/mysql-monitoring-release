package metrics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"
	 _ "github.com/go-sql-driver/mysql"

	boshCliDirector "github.com/cloudfoundry/bosh-cli/director"
	boshlogger "github.com/cloudfoundry/bosh-utils/logger"

	"fmt"
	"time"
	"os"
	"log"
)

var _ = Describe("mysql events", func() {
	// connect to bosh // inject bosh director ip/env from os.Getenv
	// maybe also inject bosh deployment
	// grab mysql user and password from manifest // dedicated-mysql-utils has methods for this
	// using bosh figure out the two mysql IPs
	// run queries on each mysql IP to determine which is leader which is follower // possibly steal from adapter release
	// use bosh to figure out instance id for leader/follower

	var (
		boshEnvironment    string
		boshDeploymentName string
		mysqlInstances           []boshCliDirector.Instance
		director           boshCliDirector.Director
		leaderInstance boshCliDirector.Instance
		followerInstance boshCliDirector.Instance
	)

	getMysqlConnection := func(hostIp string) (*sql.DB, error) {
		return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/replication_monitoring?parseTime=true",
			os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), hostIp, 3306))
	}

	filterMysqlInstances := func (instances []boshCliDirector.Instance) ([]boshCliDirector.Instance) {
		mysqlInstances := make([]boshCliDirector.Instance, 0)
		for _, instance := range instances {
			if instance.Group == "mysql" {
				mysqlInstances = append(mysqlInstances, instance)
			}
		}
		return mysqlInstances
	}

	getMysqlInstancesFromDirector := func (director boshCliDirector.Director) ([]boshCliDirector.Instance){
		deployment, err := director.FindDeployment(boshDeploymentName)
		Expect(err).NotTo(HaveOccurred())

		allInstances, err := deployment.Instances()
		Expect(err).NotTo(HaveOccurred())

		return filterMysqlInstances(allInstances)
	}

	identifyLeaderAndFollower := func(instances []boshCliDirector.Instance) (boshCliDirector.Instance, boshCliDirector.Instance) {
		var leader, follower boshCliDirector.Instance
		for _, instance := range instances {
			conn, err := getMysqlConnection(instance.IPs[0])
			Expect(err).NotTo(HaveOccurred())

			rows, err := conn.Query("show slave status")
			Expect(err).NotTo(HaveOccurred())

			if rows.Next() {
				follower = instance
			} else {
				leader = instance
			}
		}
		return leader, follower
	}



	BeforeSuite(func() {
		boshEnvironment = os.Getenv("BOSH_ENVIRONMENT")
		boshDeploymentName = os.Getenv("BOSH_DEPLOYMENT")
		boshClientName := os.Getenv("BOSH_CLIENT")
		boshClientSecret := os.Getenv("BOSH_CLIENT_SECRET") // TODO blow up if empty
		boshCACert := os.Getenv("BOSH_CA_CERT")

		var boshLogger = boshlogger.New(boshlogger.LevelDebug, log.New(os.Stdout, "", log.LstdFlags))

		config := boshCliDirector.Config{
			Host:         boshEnvironment,
			Port:         25555,
			CACert: boshCACert,
			Client:       boshClientName,
			ClientSecret: boshClientSecret,
		}
		var err error
		director, err = boshCliDirector.NewFactory(boshLogger).New(
			config,
			boshCliDirector.NewNoopTaskReporter(),
			boshCliDirector.NewNoopFileReporter(),
		)
		Expect(err).NotTo(HaveOccurred())

		mysqlInstances = getMysqlInstancesFromDirector(director)
		Expect(len(mysqlInstances)).To(Equal(2)) // leader + follower

		leaderInstance, followerInstance = identifyLeaderAndFollower(mysqlInstances)

		Expect(leaderInstance).NotTo(BeNil())
		Expect(followerInstance).NotTo(BeNil())
		Expect(leaderInstance).NotTo(Equal(followerInstance))
	})

	It("records the timestamp at some interval on the leader node", func() {
		conn, err := getMysqlConnection(leaderInstance.IPs[0])

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
		conn, err := getMysqlConnection(leaderInstance.IPs[0])

		rows, err := conn.Query("select server_id from replication_monitoring.heartbeat")
		Expect(err).NotTo(HaveOccurred())

		var serverIdFromTable string

		Expect(rows.Next()).To(BeTrue())
		rows.Scan(&serverIdFromTable)
		Expect(rows.Next()).To(BeFalse())

		Expect(serverIdFromTable).To(Equal(leaderInstance.ID))

		Consistently(func() string {
			rows, _ := conn.Query("select server_id from replication_monitoring.heartbeat")
			var newerServerIdFromTable string

			Expect(rows.Next()).To(BeTrue())
			rows.Scan(&newerServerIdFromTable)
			Expect(rows.Next()).To(BeFalse())

			return newerServerIdFromTable
		}, "6s", "1s").Should(Equal(leaderInstance.ID))
	})

	It("only runs the event on the leader node (and replicates to the follower)", func() {
		conn, err := getMysqlConnection(followerInstance.IPs[0])

		rows, err := conn.Query("select server_id from replication_monitoring.heartbeat")
		Expect(err).NotTo(HaveOccurred())

		var serverIdOnFollower string

		Expect(rows.Next()).To(BeTrue())
		rows.Scan(&serverIdOnFollower)
		Expect(rows.Next()).To(BeFalse())

		Expect(serverIdOnFollower).To(Equal(leaderInstance.ID))
	})
})
