package metrics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/go-sql-driver/mysql"

	boshCliDirector "github.com/cloudfoundry/bosh-cli/director"
	boshlogger "github.com/cloudfoundry/bosh-utils/logger"

	"acceptance_tests/helpers"
	"log"
	"os"
	"time"
)

var _ = Describe("mysql events", func() {
	var (
		boshEnvironment      string
		boshDeploymentName   string
		mysqlMetricsUsername string
		mysqlMetricsPassword string
		databaseName         string
		databasePort         int
		mysqlInstances       []boshCliDirector.Instance
		director             boshCliDirector.Director
		leaderInstance       boshCliDirector.Instance
		followerInstance     boshCliDirector.Instance
	)

	BeforeSuite(func() {
		boshEnvironment = helpers.GetEnvVar("BOSH_ENVIRONMENT")
		boshDeploymentName = helpers.GetEnvVar("BOSH_DEPLOYMENT")
		boshClientName := helpers.GetEnvVar("BOSH_CLIENT")
		boshClientSecret := helpers.GetEnvVar("BOSH_CLIENT_SECRET")
		boshCACert := helpers.GetEnvVar("BOSH_CA_CERT")
		mysqlMetricsUsername = "mysql-metrics"
		databaseName = "replication_monitoring"
		databasePort = 3306

		var boshLogger = boshlogger.New(boshlogger.LevelNone, log.New(os.Stdout, "", log.LstdFlags))

		config := boshCliDirector.Config{
			Host:         boshEnvironment,
			Port:         25555,
			CACert:       boshCACert,
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

		deployment, err := director.FindDeployment(boshDeploymentName)
		Expect(err).NotTo(HaveOccurred())

		manifest, err := deployment.Manifest()
		Expect(err).NotTo(HaveOccurred())

		mysqlMetricsPassword = helpers.GetManifestValue(manifest, "/instance_groups/name=mysql/properties/mysql_metrics_password")

		mysqlInstances = getMysqlInstancesFromDirector(boshDeploymentName, director)
		Expect(len(mysqlInstances)).To(Equal(2)) // leader + follower

		leaderInstance, followerInstance = identifyLeaderAndFollower(mysqlInstances, databaseConnectionConfig{
			databaseName: databaseName,
			databasePort: databasePort,
			username: mysqlMetricsUsername,
			password: mysqlMetricsPassword,
		})

		Expect(leaderInstance).NotTo(BeNil())
		Expect(followerInstance).NotTo(BeNil())
		Expect(leaderInstance).NotTo(Equal(followerInstance))
	})

	It("records the timestamp at some interval on the leader node", func() {
		conn := helpers.GetMysqlConnection(leaderInstance.IPs[0], databasePort, mysqlMetricsUsername, mysqlMetricsPassword, databaseName)

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
		conn := helpers.GetMysqlConnection(leaderInstance.IPs[0], databasePort, mysqlMetricsUsername, mysqlMetricsPassword, databaseName)

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
		conn := helpers.GetMysqlConnection(followerInstance.IPs[0], databasePort, mysqlMetricsUsername, mysqlMetricsPassword, databaseName)

		rows, err := conn.Query("select server_id from replication_monitoring.heartbeat")
		Expect(err).NotTo(HaveOccurred())

		var serverIdOnFollower string

		Expect(rows.Next()).To(BeTrue())
		rows.Scan(&serverIdOnFollower)
		Expect(rows.Next()).To(BeFalse())

		Expect(serverIdOnFollower).To(Equal(leaderInstance.ID))
	})
})

type databaseConnectionConfig struct {
	databaseName string
	databasePort int
	username     string
	password     string
}

func filterMysqlInstances(instances []boshCliDirector.Instance) []boshCliDirector.Instance {
	mysqlInstances := make([]boshCliDirector.Instance, 0)
	for _, instance := range instances {
		if instance.Group == "mysql" {
			mysqlInstances = append(mysqlInstances, instance)
		}
	}
	return mysqlInstances
}

func getMysqlInstancesFromDirector(boshDeploymentName string, director boshCliDirector.Director) []boshCliDirector.Instance {
	deployment, err := director.FindDeployment(boshDeploymentName)
	Expect(err).NotTo(HaveOccurred())

	allInstances, err := deployment.Instances()
	Expect(err).NotTo(HaveOccurred())

	return filterMysqlInstances(allInstances)
}

func identifyLeaderAndFollower(instances []boshCliDirector.Instance, config databaseConnectionConfig) (boshCliDirector.Instance, boshCliDirector.Instance) {
	var leader, follower boshCliDirector.Instance
	for _, instance := range instances {
		conn := helpers.GetMysqlConnection(
			instance.IPs[0],
			config.databasePort,
			config.username,
			config.password,
			config.databaseName,
		)

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
