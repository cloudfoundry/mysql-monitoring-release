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
	unhealthy := canaryclient.Check(c.Canary)

	nodeClusterStatuses := database.GetNodeClusterStatuses(c.Mysql)
	needsBootstrap := database.CheckClusterBootstrapStatus(nodeClusterStatuses)

	nodeDiskInfos := disk.GetNodeDiskInfos(c.Mysql)
	diskSpaceIssues := disk.CheckDiskStatus(nodeDiskInfos, c.Mysql.Threshold)

	table := ui.NewTable(os.Stdout)
	table.AddClusterDataToTable(nodeClusterStatuses)
	table.AddDiskDataToTable(nodeDiskInfos)
	table.Render()

	messages := ui.Report(ui.ReporterParams{
		IsCanaryHealthy:     !unhealthy,
		NeedsBootstrap:      needsBootstrap,
		DiskSpaceIssues:     diskSpaceIssues,
		NodeClusterStatuses: nodeClusterStatuses,
	})

	for _, message := range messages {
		fmt.Println(message)
	}
}
