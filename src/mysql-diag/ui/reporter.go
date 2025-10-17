package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/cloudfoundry/mysql-diag/data"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
	"github.com/cloudfoundry/mysql-diag/msg"
)

type ReporterParams struct {
	NeedsBootstrap      bool
	DiskSpaceIssues     []disk.DiskSpaceIssue
	NodeClusterStatuses []*database.NodeClusterStatus
	Proxies             []data.Proxy
}

func Report(params ReporterParams) []string {
	messages := []string{}

	if params.NeedsBootstrap {
		slices.SortStableFunc(params.NodeClusterStatuses, func(i, j *database.NodeClusterStatus) int {
			return cmp.Compare(i.Status.LastApplied, j.Status.LastApplied)
		})
		bootstrapNode := fmt.Sprintf("%s/%s", params.NodeClusterStatuses[len(params.NodeClusterStatuses)-1].Node.Name, params.NodeClusterStatuses[len(params.NodeClusterStatuses)-1].Node.UUID)
		messages = append(messages, msg.Alert("\n[CRITICAL] You must bootstrap the cluster. Follow these instructions: https://docs.vmware.com/en/VMware-SQL-with-MySQL-for-Tanzu-Application-Service/3.2/mysql-for-tas/bootstrapping.html"))
		messages = append(messages, msg.Alert(fmt.Sprintf("\n[CRITICAL] Bootstrap node: \"%s\"", bootstrapNode)))
	}

	if params.NeedsBootstrap {
		messages = append(messages, msg.Alert("\n[CRITICAL] Run the bosh logs command: targeting each of the VMs in your VMware SQL with MySQL for TAS cluster, proxies, and jumpbox to retrieve the VM logs."))
	}

	if params.NeedsBootstrap || len(params.DiskSpaceIssues) > 0 {
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
	if !params.NeedsBootstrap {
		var activeBackends []string
		for _, proxy := range params.Proxies {
			for _, backend := range proxy.Backends {
				if backend.Active {
					activeBackends = append(activeBackends, backend.Name)
				}
			}
		}
		slices.Sort(activeBackends)
		activeBackends = slices.Compact(activeBackends)
		if len(activeBackends) > 1 {
			messages = append(messages, msg.Error(fmt.Sprintf("MISMATCHED ACTIVE BACKENDS:\n%s\n", strings.Join(activeBackends, "\n"))))
		} else if len(activeBackends) == 1 {
			messages = append(messages, msg.Alert(fmt.Sprintf("NOTE: Proxies will currently attempt to direct traffic to \"%s\"", activeBackends[0])))
		}
	}

	return messages
}
