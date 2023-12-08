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

func pointer[T any](v T) *T {
	return &v
}

// Returns false if the canary is unhealthy. true if healthy or unknown, nil if canary is disabled.
func checkCanary(config *config.CanaryConfig) *bool {
	if config == nil {
		return nil
	}

	intro := "Checking canary status... "
	client := canaryclient.NewCanaryClient("127.0.0.1", config.ApiPort, *config)
	healthy, err := client.Status()
	if err != nil {
		msg.PrintfErrorIntro(intro, "%v", err)
		return pointer(true)
	} else {
		if healthy {
			fmt.Println(intro + msg.Happy("healthy"))
			return pointer(true)
		} else {
			fmt.Println(intro + msg.Alert("unhealthy"))
			return pointer(false)
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
	messages := ui.Report(ui.ReporterParams{
		IsCanaryHealthy: checkCanary(c.Canary),
		NeedsBootstrap:  checkClusterStatus(c.Mysql),
		DiskSpaceIssues: disk.CheckDiskStatus(c.Mysql),
	})

	for _, message := range messages {
		fmt.Println(message)
	}
}
