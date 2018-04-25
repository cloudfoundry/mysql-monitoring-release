package disk

import (
	"fmt"
	"mysql-diag/config"
	"mysql-diag/diagagentclient"
	. "mysql-diag/diskspaceissue"
	"mysql-diag/msg"
	"mysql-diag/ui"
	"os"
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

func PercentUsed(total uint64, free uint64) uint{
	used := total - free
	if total == 0 {
		return 100
	}
	return uint(used * 100 / total)
}

func CheckDiskStatus(mysqlConfig config.MysqlConfig) []DiskSpaceIssue {
	if mysqlConfig.Agent == nil {
		fmt.Println("Agent not configured, skipping disk check")
		return nil
	}

	rows := getDisks(mysqlConfig)

	if isAnyInfoPresent(rows) {
		diskInfoTable := ui.NewDiskInfoTable(os.Stdout)

		for _, row := range rows {
			n := row.Node
			diskInfoTable.Add(n.Host, n.Name, n.UUID, row.Info)
		}

		diskInfoTable.Render()

		return ValidateCapacity(rows, mysqlConfig.Threshold)
	} else {
		fmt.Println(msg.Alert("Unable to gather disk usage information, moving on. Run bosh vms --vitals for this information."))
		return nil
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

func getDisks(mysqlConfig config.MysqlConfig) []NodeDiskInfo {
	channel := make(chan NodeDiskInfo, len(mysqlConfig.Nodes))

	for _, n := range mysqlConfig.Nodes {
		n := n

		go func() {
			intro := fmt.Sprintf("Checking disk status of %s/%s at %s... ", n.Name, n.UUID, n.Host)
			fmt.Println(intro)

			client := diagagentclient.NewDiagAgentClient(n.Host, mysqlConfig.Agent.Port, mysqlConfig.Agent.Username, mysqlConfig.Agent.Password)
			info, err := client.Info()
			if err != nil {
				msg.PrintfErrorIntro(intro, "%v", err)
			} else {
				fmt.Println(intro + "done")
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
