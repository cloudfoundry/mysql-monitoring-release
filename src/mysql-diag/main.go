package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/mysql-diag/canaryclient"
	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
	"github.com/cloudfoundry/mysql-diag/msg"
	"github.com/cloudfoundry/mysql-diag/ui"
)

const (
	defaultConfigPath = "/var/vcap/jobs/mysql-diag/config/mysql-diag-config.yml"
)

var configFilepath = flag.String("c", defaultConfigPath, "location of config file")

// returns true if the cluster needs bootstrap
func checkClusterBootstrapStatus(rows []*nodeClusterStatus) bool {
	statuses := make([]*database.GaleraStatus, len(rows))
	for i, row := range rows {
		statuses[i] = row.status
	}

	if database.DoWeNeedBootstrap(statuses) {
		return true
	} else {
		return false
	}
}

func addClusterDataToTable(nodeStatuses []*nodeClusterStatus, table *ui.Table) {
	for _, row := range nodeStatuses {
		n := row.node
		table.AddClusterInfo(n.Name, n.UUID, row.status)
	}
}

type nodeClusterStatus struct {
	node   config.MysqlNode
	status *database.GaleraStatus
}

func getNodeClusterStatuses(mysqlConfig config.MysqlConfig) []*nodeClusterStatus {
	channel := make(chan nodeClusterStatus, len(mysqlConfig.Nodes))

	for _, n := range mysqlConfig.Nodes {
		n := n

		go func() {
			ac := database.NewDatabaseClient(mysqlConfig.Connection(n))
			galeraStatus, err := ac.Status()
			if err != nil {
				msg.PrintfErrorIntro("", "error retrieving galera status: %v", err)
			}

			channel <- nodeClusterStatus{node: n, status: galeraStatus}
		}()
	}

	var nodeStatuses []*nodeClusterStatus
	for i := 0; i < len(mysqlConfig.Nodes); i++ {
		ns := <-channel
		nodeStatuses = append(nodeStatuses, &ns)
	}

	return nodeStatuses
}

// Returns true if the canary is unhealthy. Otherwise, it's either healthy or unknown.
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

	nodeClusterStatuses := getNodeClusterStatuses(c.Mysql)
	needsBootstrap := checkClusterBootstrapStatus(nodeClusterStatuses)

	nodeDiskInfos := disk.GetNodeDiskInfos(c.Mysql)
	diskSpaceIssues := disk.CheckDiskStatus(nodeDiskInfos, c.Mysql.Threshold)

	table := ui.NewTable(os.Stdout)
	addClusterDataToTable(nodeClusterStatuses, table)
	disk.AddDiskDataToTable(nodeDiskInfos, table)
	table.Render()

	messages := ui.Report(ui.ReporterParams{
		IsCanaryHealthy: !unhealthy,
		NeedsBootstrap:  needsBootstrap,
		DiskSpaceIssues: diskSpaceIssues,
	})

	for _, message := range messages {
		fmt.Println(message)
	}
}
