package ui

import (
	"fmt"

	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
	"github.com/cloudfoundry/mysql-diag/msg"
)

type ReporterParams struct {
	IsCanaryHealthy     bool
	NeedsBootstrap      bool
	DiskSpaceIssues     []disk.DiskSpaceIssue
	NodeClusterStatuses []*database.NodeClusterStatus
}

func Report(params ReporterParams) []string {
	messages := []string{}

	if !params.IsCanaryHealthy {
		messages = append(messages, msg.Alert("\n[CRITICAL] The replication process is unhealthy. Writes are disabled."))
	}

	if params.NeedsBootstrap {
		messages = append(messages, msg.Alert("\n[CRITICAL] You must bootstrap the cluster. Follow these instructions: https://docs.vmware.com/en/VMware-SQL-with-MySQL-for-Tanzu-Application-Service/3.2/mysql-for-tas/bootstrapping.html"))
	}

	if !params.IsCanaryHealthy || params.NeedsBootstrap {
		messages = append(messages, msg.Alert("\n[CRITICAL] Run the download-logs command:")+`
$ download-logs -o /tmp/output
For full information about how to download and use the download-logs command see https://community.pivotal.io/s/article/script-to-download-mysql-logs-for-pas-tile-ha-clusters?language=en_US`)
	}

	if !params.IsCanaryHealthy || params.NeedsBootstrap || len(params.DiskSpaceIssues) > 0 {
		for _, diskSpaceIssue := range params.DiskSpaceIssues {
			messages = append(messages, msg.Alert(fmt.Sprintf("\n[WARNING] %s disk usage is very high on node %s. Some fluctuation on the node currently serving "+
				"transactions is normal, due to temporary table usage, but be aware that MySQL needs to have sufficient free space "+
				"to operate. Consider re-deploying with larger %s disks.", diskSpaceIssue.DiskType, diskSpaceIssue.NodeName, diskSpaceIssue.DiskType)))
		}
		messages = append(messages, msg.Alert("\n[WARNING] NOT RECOMMENDED")+`
Do not perform the following unless instructed by Pivotal Support:
- Do not scale down the cluster to one node then scale back. This puts user data at risk.
- Avoid "bosh recreate" and "bosh cck". These options remove logs on the VMs making it harder to diagnose cluster issues.

`)
	}
	name := ""
	minLocalIndex := maxUUID
	for _, status := range params.NodeClusterStatuses {
		if status.Status != nil && status.Status.LocalIndex < minLocalIndex {
			minLocalIndex = status.Status.LocalIndex
			name = fmt.Sprintf("%s/%s", status.Node.Name, status.Node.UUID)
		}
	}
	messages = append(messages, msg.Alert(fmt.Sprintf("NOTE: Proxies will currently attempt to direct traffic to \"%s\"", name)))

	return messages
}
