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
	"github.com/cloudfoundry/mysql-diag/diskspaceissue"
	"github.com/cloudfoundry/mysql-diag/msg"
	"github.com/cloudfoundry/mysql-diag/ui"
)

const (
	defaultConfigPath = "/var/vcap/jobs/mysql-diag/config/mysql-diag-config.yml"
)

var configFilepath = flag.String("c", defaultConfigPath, "location of config file")

// returns true if the cluster needs bootstrap
func checkClusterBootstrapStatus(rows []*nodeStatus) bool {
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

func renderClusterTable(nodeList []*nodeStatus) {
	clusterStateTable := ui.NewClusterStateTable(os.Stdout)

	for _, row := range nodeList {
		n := row.node
		clusterStateTable.Add(n.Name, n.UUID, row.status)
	}

	clusterStateTable.Render()
}

type nodeStatus struct {
	node   config.MysqlNode
	status *database.GaleraStatus
}

func getNodesClusterInfo(mysqlConfig config.MysqlConfig) []*nodeStatus {
	channel := make(chan nodeStatus, len(mysqlConfig.Nodes))

	for _, n := range mysqlConfig.Nodes {
		n := n

		go func() {
			ac := database.NewDatabaseClient(mysqlConfig.Connection(n))
			galeraStatus, err := ac.Status()
			if err != nil {
				msg.PrintfErrorIntro("", "error retrieving galera status: %v", err)
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
	var diskStatus []diskspaceissue.DiskSpaceIssue
	flag.Parse()

	c, err := config.LoadFromFile(*configFilepath)
	if err != nil {
		msg.PrintfError("%v", err)
		os.Exit(1)
	}

	printCurrentTime()
	unhealthy := checkCanary(c.Canary)

	cNodes := getNodesClusterInfo(c.Mysql)
	renderClusterTable(cNodes)
	needsBootstrap := checkClusterBootstrapStatus(cNodes)

	if c.Mysql.Agent == nil {
		fmt.Println("Agent not configured, skipping disk check")
	} else {
		dNodes := disk.GetNodesDiskInfo(c.Mysql)
		disk.RenderDiskTable(dNodes)
		diskStatus = disk.CheckDiskStatus(dNodes, c.Mysql.Threshold)
	}

	messages := ui.Report(ui.ReporterParams{
		IsCanaryHealthy: !unhealthy,
		NeedsBootstrap:  needsBootstrap,
		DiskSpaceIssues: diskStatus,
	})

	for _, message := range messages {
		fmt.Println(message)
	}
}
