package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/data"
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

	aggregator := data.NewAggregator(c.Mysql, c.GaleraAgent)
	aggregatedData := aggregator.Aggregate()

	table := ui.NewTable(os.Stdout)
	table.AddClusterData(aggregatedData.NodeClusterStatuses)
	table.AddDiskData(aggregatedData.NodeDiskInfo)
	table.Render()

	messages := ui.Report(ui.ReporterParams{
		NeedsBootstrap:      aggregatedData.NeedsBootstrap,
		DiskSpaceIssues:     aggregatedData.DiskSpaceIssues,
		NodeClusterStatuses: aggregatedData.NodeClusterStatuses,
	})

	for _, message := range messages {
		fmt.Println(message)
	}
}
