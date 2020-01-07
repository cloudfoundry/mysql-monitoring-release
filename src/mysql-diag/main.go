package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/canaryclient"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/config"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/database"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/disk"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/msg"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/ui"
)

const (
	defaultConfigPath = "/var/vcap/jobs/mysql-diag/config/mysql-diag-config.yml"
)

var configFilepath = flag.String("c", defaultConfigPath, "location of config file")

// returns true if the cluster needs bootstrap
func checkClusterStatus(mysqlConfig config.MysqlConfig) bool {
	clusterStateTable := ui.NewClusterStateTable(os.Stdout)

	rows := getClusterStatus(mysqlConfig)
	for _, row := range rows {
		n := row.node
		clusterStateTable.Add(n.Host, n.Name, n.UUID, row.status)
	}

	clusterStateTable.Render()

	statuses := make([]*database.GaleraStatus, len(rows))
	for i, row := range rows {
		statuses[i] = row.status
	}

	if database.DoWeNeedBootstrap(statuses) {
		return true
	} else {
		fmt.Println("I don't think bootstrap is necessary")
		return false
	}
}

type nodeStatus struct {
	node   config.MysqlNode
	status *database.GaleraStatus
}

func getClusterStatus(mysqlConfig config.MysqlConfig) []*nodeStatus {
	channel := make(chan nodeStatus, len(mysqlConfig.Nodes))

	for _, n := range mysqlConfig.Nodes {
		n := n

		go func() {
			intro := fmt.Sprintf("Checking cluster status of %s/%s at %s... ", n.Name, n.UUID, n.Host)
			fmt.Println(intro)

			ac := database.NewDatabaseClient(mysqlConfig.Connection(n))
			galeraStatus, err := ac.Status()
			if err != nil {
				msg.PrintfErrorIntro(intro, "%v", err)
			} else {
				fmt.Println(intro + "done")
			}

			channel <- nodeStatus{node: n, status: galeraStatus}
		}()
	}

	var nodeStatuses []*nodeStatus
	for i := 0; i < len(mysqlConfig.Nodes); i++ {
		ns := <-channel
		nodeStatuses = append(nodeStatuses, &ns)
	}

	return nodeStatuses
}

// Returns true if the canary is unhealthy. Otherwise it's either healthy or unknown.
func checkCanary(config *config.CanaryConfig) bool {
	if config == nil {
		fmt.Println("Canary not configured, skipping health check")
		return false
	}

	intro := "Checking canary status... "
	fmt.Println(intro)

	client := canaryclient.NewCanaryClient("127.0.0.1", config.ApiPort, *config)
	healthy, err := client.Status()
	if err != nil {
		msg.PrintfErrorIntro(intro, "%v", err)
		return false
	} else {
		if healthy {
			fmt.Println(intro + msg.Happy("healthy"))
			return false
		} else {
			fmt.Println(intro + msg.Alert("unhealthy"))
			return true
		}
	}
}

func printCurrentTime() {
	fmt.Println(time.Now().UTC().Format(time.UnixDate))
}

func main() {
	flag.Parse()

	c, err := config.LoadFromFile(*configFilepath)
	if err != nil {
		msg.PrintfError("%v", err)
		os.Exit(1)
	}

	printCurrentTime()
	unhealthy := checkCanary(c.Canary)
	needsBootstrap := checkClusterStatus(c.Mysql)
	diskSpaceIssues := disk.CheckDiskStatus(c.Mysql)

	messages := ui.Report(ui.ReporterParams{
		IsCanaryHealthy: !unhealthy,
		NeedsBootstrap:  needsBootstrap,
		DiskSpaceIssues: diskSpaceIssues,
	}, c)

	for _, message := range messages {
		fmt.Println(message)
	}
}
