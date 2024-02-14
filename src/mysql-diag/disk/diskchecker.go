package disk

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/diagagentclient"
	. "github.com/cloudfoundry/mysql-diag/diskspaceissue"
	"github.com/cloudfoundry/mysql-diag/msg"
	"github.com/cloudfoundry/mysql-diag/ui"
)

func ValidateCapacity(nodeDiskInfos []NodeDiskInfo, threshold *config.ThresholdConfig) []DiskSpaceIssue {
	issues := []DiskSpaceIssue{}

	for _, nodeDiskInfo := range nodeDiskInfos {
		if nodeDiskInfo.Info == nil {
			continue
		}

		ephemeralPercentBytesUsed := PercentUsed(nodeDiskInfo.Info.Ephemeral.BytesTotal, nodeDiskInfo.Info.Ephemeral.BytesFree)

		ephemeralPercentInodesUsed := PercentUsed(nodeDiskInfo.Info.Ephemeral.InodesTotal, nodeDiskInfo.Info.Ephemeral.InodesFree)

		persistentPercentBytesUsed := PercentUsed(nodeDiskInfo.Info.Persistent.BytesTotal, nodeDiskInfo.Info.Persistent.BytesFree)

		persistentPercentInodesUsed := PercentUsed(nodeDiskInfo.Info.Persistent.InodesTotal, nodeDiskInfo.Info.Persistent.InodesFree)

		if ephemeralPercentBytesUsed > threshold.DiskUsedWarningPercent ||
			ephemeralPercentInodesUsed > threshold.DiskInodesUsedWarningPercent {
			issues = append(issues, DiskSpaceIssue{
				DiskType: "Ephemeral",
				NodeName: nodeDiskInfo.Node.Name + "/" + nodeDiskInfo.Node.UUID,
			})
		}

		if persistentPercentBytesUsed > threshold.DiskUsedWarningPercent ||
			persistentPercentInodesUsed > threshold.DiskInodesUsedWarningPercent {
			issues = append(issues, DiskSpaceIssue{
				DiskType: "Persistent",
				NodeName: nodeDiskInfo.Node.Name + "/" + nodeDiskInfo.Node.UUID,
			})
		}
	}

	return issues
}

func PercentUsed(total uint64, free uint64) uint {
	used := total - free
	if total == 0 {
		return 100
	}
	return uint(used * 100 / total)
}

func CheckDiskStatus(nodeList []NodeDiskInfo, t *config.ThresholdConfig) []DiskSpaceIssue {
	if isAnyInfoPresent(nodeList) {
		return ValidateCapacity(nodeList, t)
	} else {
		return nil
	}
}

func RenderDiskTable(nodeList []NodeDiskInfo) {
	if isAnyInfoPresent(nodeList) {
		diskInfoTable := ui.NewDiskInfoTable(os.Stdout)

		for _, row := range nodeList {
			n := row.Node
			diskInfoTable.Add(n.Name, n.UUID, row.Info)
		}

		diskInfoTable.Render()
	} else {
		fmt.Println(msg.Alert("Unable to gather disk usage information, moving on. Run bosh vms --vitals for this information."))
	}
}

func isAnyInfoPresent(infos []NodeDiskInfo) bool {
	for _, info := range infos {
		if info.Info != nil {
			return true
		}
	}
	return false
}

func GetNodesDiskInfo(mysqlConfig config.MysqlConfig) []NodeDiskInfo {
	channel := make(chan NodeDiskInfo, len(mysqlConfig.Nodes))

	for _, n := range mysqlConfig.Nodes {
		n := n

		go func() {
			client := diagagentclient.NewDiagAgentClient(*mysqlConfig.Agent)
			address := fmt.Sprintf("%s:%d", n.Host, mysqlConfig.Agent.Port)
			info, err := client.Info(address, mysqlConfig.Agent.TLS.Enabled)
			if err != nil {
				msg.PrintfErrorIntro("", "error retrieving disk info: %v", err)
			}

			channel <- NodeDiskInfo{Node: n, Info: info}
		}()
	}

	var nodeDiskInfos []NodeDiskInfo
	for i := 0; i < len(mysqlConfig.Nodes); i++ {
		ns := <-channel
		nodeDiskInfos = append(nodeDiskInfos, ns)
	}

	return nodeDiskInfos
}
