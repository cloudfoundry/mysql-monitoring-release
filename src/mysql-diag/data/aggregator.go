package data

import (
	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
	mysqlAgentClient "github.com/cloudfoundry/mysql-diag/galera-agent-client"
	"github.com/cloudfoundry/mysql-diag/proxy"
)

type Data struct {
	NodeClusterStatuses []*database.NodeClusterStatus

	NodeDiskInfo    []disk.NodeDiskInfo
	DiskSpaceIssues []disk.DiskSpaceIssue

	NeedsBootstrap bool

	Proxies []Proxy
}

type Proxy struct {
	Name     string
	Backends []proxy.Backend
}

type aggregator struct {
	mySQL        config.MysqlConfig
	galeraAgent  *config.GaleraAgentConfig
	proxyClients []proxy.Client
}

func NewAggregator(mysql config.MysqlConfig, galeraAgent *config.GaleraAgentConfig, proxyClients []proxy.Client) aggregator {
	return aggregator{
		mySQL:        mysql,
		galeraAgent:  galeraAgent,
		proxyClients: proxyClients,
	}
}

func (a aggregator) Aggregate() Data {
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

	var proxies []Proxy
	for _, client := range a.proxyClients {
		backends := client.Backends()
		proxies = append(proxies, Proxy{
			Name:     client.Name,
			Backends: backends,
		})
	}
	return Data{
		NodeClusterStatuses: statuses,
		NodeDiskInfo:        nodeDiskInfos,
		DiskSpaceIssues:     diskSpaceIssues,
		NeedsBootstrap:      needsBootstrap,
		Proxies:             proxies,
	}
}
