package data

import (
	"github.com/cloudfoundry/mysql-diag/canaryclient"
	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
	mysqlAgentClient "github.com/cloudfoundry/mysql-diag/galera-agent-client"
)

type Data struct {
	NodeClusterStatuses []*database.NodeClusterStatus

	NodeDiskInfo    []disk.NodeDiskInfo
	DiskSpaceIssues []disk.DiskSpaceIssue

	Unhealthy bool

	NeedsBootstrap bool
}

type aggregator struct {
	canary      *config.CanaryConfig
	mySQL       config.MysqlConfig
	galeraAgent *config.GaleraAgentConfig
}

func NewAggregator(canary *config.CanaryConfig, mysql config.MysqlConfig, galeraAgent *config.GaleraAgentConfig) aggregator {
	return aggregator{
		canary:      canary,
		mySQL:       mysql,
		galeraAgent: galeraAgent,
	}
}

func (a aggregator) Aggregate() Data {
	unhealthy := canaryclient.Check(a.canary)

	nodeClusterStatuses := make(map[string]*database.NodeClusterStatus)
	for _, node := range a.mySQL.Nodes {
		nodeClusterStatuses[node.Host] = &database.NodeClusterStatus{
			Node:   node,
			Status: nil,
		}
	}
	database.GetNodeClusterStatuses(a.mySQL, nodeClusterStatuses)
	needsBootstrap := database.CheckClusterBootstrapStatus(nodeClusterStatuses)
	mysqlAgentClient.GetSequenceNumbers(a.galeraAgent, nodeClusterStatuses)
	nodeDiskInfos := disk.GetNodeDiskInfos(a.mySQL)
	diskSpaceIssues := disk.CheckDiskStatus(nodeDiskInfos, a.mySQL.Threshold)

	statuses := make([]*database.NodeClusterStatus, 0, len(nodeClusterStatuses))
	for _, value := range nodeClusterStatuses {
		statuses = append(statuses, value)
	}

	return Data{
		NodeClusterStatuses: statuses,
		NodeDiskInfo:        nodeDiskInfos,
		DiskSpaceIssues:     diskSpaceIssues,
		Unhealthy:           unhealthy,
		NeedsBootstrap:      needsBootstrap,
	}
}
