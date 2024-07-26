package data

import (
	"github.com/cloudfoundry/mysql-diag/canaryclient"
	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
)

type Data struct {
	NodeClusterStatuses []*database.NodeClusterStatus

	NodeDiskInfo    []disk.NodeDiskInfo
	DiskSpaceIssues []disk.DiskSpaceIssue

	Unhealthy bool

	NeedsBootstrap bool
}

type aggregator struct {
	canary *config.CanaryConfig
	mySQL  config.MysqlConfig
}

func NewAggregator(canary *config.CanaryConfig, mysql config.MysqlConfig) aggregator {
	return aggregator{
		canary: canary,
		mySQL:  mysql,
	}
}

func (a aggregator) Aggregate() Data {
	unhealthy := canaryclient.Check(a.canary)

	nodeClusterStatuses := database.GetNodeClusterStatuses(a.mySQL)
	sequenceNumbers, isUnhealthy := mysqlAgentClient.GetSequenceNumbers(a.MySQL)
	if isUnhealthy {
		for node, seqNo := range sequenceNumbers {
			nodeClusterStatuses[node].Status.LastApplied = seqNo
		}
	}
	needsBootstrap := database.CheckClusterBootstrapStatus(nodeClusterStatuses)

	nodeDiskInfos := disk.GetNodeDiskInfos(a.mySQL)
	diskSpaceIssues := disk.CheckDiskStatus(nodeDiskInfos, a.mySQL.Threshold)

	return Data{
		NodeClusterStatuses: nodeClusterStatuses,
		NodeDiskInfo:        nodeDiskInfos,
		DiskSpaceIssues:     diskSpaceIssues,
		Unhealthy:           unhealthy,
		NeedsBootstrap:      needsBootstrap,
	}
}
