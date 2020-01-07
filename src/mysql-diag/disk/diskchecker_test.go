package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/config"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/diagagentclient"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag/disk"
)

var (
	thresholdConfig *config.ThresholdConfig
)

type Params struct {
	nodeName, uuid                                                                     string
	ephemeralBytesFree, ephemeralInodesFree, persistentBytesFree, persistentInodesFree int
}

func createNodeDiskInfos(params Params) []disk.NodeDiskInfo {
	return []disk.NodeDiskInfo{
		{
			Node: config.MysqlNode{
				Host: "Host",
				Name: params.nodeName,
				UUID: params.uuid,
			},
			Info: &diagagentclient.InfoResponse{
				Persistent: diagagentclient.DiskInfo{
					BytesTotal:  uint64(100),
					BytesFree:   uint64(params.persistentBytesFree),
					InodesTotal: uint64(100),
					InodesFree:  uint64(params.persistentInodesFree),
				},
				Ephemeral: diagagentclient.DiskInfo{
					BytesTotal:  uint64(100),
					BytesFree:   uint64(params.ephemeralBytesFree),
					InodesTotal: uint64(100),
					InodesFree:  uint64(params.ephemeralInodesFree),
				},
			},
		},
	}
}

var _ = Describe("DiskChecker", func() {
	BeforeEach(func() {
		thresholdConfig = &config.ThresholdConfig{
			DiskUsedWarningPercent:       90,
			DiskInodesUsedWarningPercent: 80,
		}
	})

	Describe("#ValidateCapacity", func() {
		Context("When the node could not retrieve disk info", func() {
			It("does not erroneously blame disk space", func() {
				nodeDiskInfos := []disk.NodeDiskInfo{
					{
						Node: config.MysqlNode{
							Host: "Host",
							Name: "Name",
							UUID: "UUID",
						},
					},
				}
				diskSpaceIssues := disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(diskSpaceIssues).To(BeEmpty())
			})
		})

		Context("when an ephemeral disk does not have enough space", func() {
			It("gets returned as a DiskSpaceIssue", func() {
				nodeName := "NodeName"
				uuid := "uuid"
				nodeDiskInfos := createNodeDiskInfos(Params{
					nodeName:             nodeName,
					uuid:                 uuid,
					persistentBytesFree:  100,
					persistentInodesFree: 100,
					ephemeralBytesFree:   9,
					ephemeralInodesFree:  21,
				})
				diskSpaceIssues := disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(1))
				diskSpaceIssue := diskSpaceIssues[0]
				Expect(diskSpaceIssue.DiskType).To(Equal("Ephemeral"))
				Expect(diskSpaceIssue.NodeName).To(Equal(nodeName + "/" + uuid))

				nodeDiskInfos = createNodeDiskInfos(Params{
					nodeName:             nodeName,
					uuid:                 uuid,
					persistentBytesFree:  100,
					persistentInodesFree: 100,
					ephemeralBytesFree:   100,
					ephemeralInodesFree:  19,
				})
				diskSpaceIssues = disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(1))
				diskSpaceIssue = diskSpaceIssues[0]
				Expect(diskSpaceIssue.DiskType).To(Equal("Ephemeral"))
				Expect(diskSpaceIssue.NodeName).To(Equal(nodeName + "/" + uuid))

				nodeDiskInfos = createNodeDiskInfos(Params{
					nodeName:             nodeName,
					uuid:                 uuid,
					persistentBytesFree:  100,
					persistentInodesFree: 100,
					ephemeralBytesFree:   9,
					ephemeralInodesFree:  9,
				})
				diskSpaceIssues = disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(1))
				diskSpaceIssue = diskSpaceIssues[0]
				Expect(diskSpaceIssue.DiskType).To(Equal("Ephemeral"))
				Expect(diskSpaceIssue.NodeName).To(Equal(nodeName + "/" + uuid))
			})
		})

		Context("when an ephemeral disk has enough space", func() {
			It("does not get returned as a DiskSpaceIssue", func() {
				nodeDiskInfos := createNodeDiskInfos(Params{
					nodeName:             "nodeName",
					uuid:                 "uuid",
					persistentBytesFree:  100,
					persistentInodesFree: 100,
					ephemeralBytesFree:   10,
					ephemeralInodesFree:  20,
				})
				diskSpaceIssues := disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(diskSpaceIssues).To(BeEmpty())
			})
		})

		Context("when a persistent disk does not have enough space", func() {
			It("gets returned as a DiskSpaceIssue", func() {
				nodeName := "Different Node Name"
				uuid := "uuie"
				nodeDiskInfos := createNodeDiskInfos(Params{
					nodeName:             nodeName,
					uuid:                 uuid,
					persistentBytesFree:  9,
					persistentInodesFree: 100,
					ephemeralBytesFree:   100,
					ephemeralInodesFree:  100,
				})
				diskSpaceIssues := disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(1))
				diskSpaceIssue := diskSpaceIssues[0]
				Expect(diskSpaceIssue.DiskType).To(Equal("Persistent"))
				Expect(diskSpaceIssue.NodeName).To(Equal(nodeName + "/" + uuid))

				nodeDiskInfos = createNodeDiskInfos(Params{
					nodeName:             nodeName,
					uuid:                 uuid,
					persistentBytesFree:  100,
					persistentInodesFree: 19,
					ephemeralBytesFree:   100,
					ephemeralInodesFree:  100,
				})
				diskSpaceIssues = disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(1))
				diskSpaceIssue = diskSpaceIssues[0]
				Expect(diskSpaceIssue.DiskType).To(Equal("Persistent"))
				Expect(diskSpaceIssue.NodeName).To(Equal(nodeName + "/" + uuid))

				nodeDiskInfos = createNodeDiskInfos(Params{
					nodeName:             nodeName,
					uuid:                 uuid,
					persistentBytesFree:  9,
					persistentInodesFree: 9,
					ephemeralBytesFree:   100,
					ephemeralInodesFree:  100,
				})
				diskSpaceIssues = disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(1))
				diskSpaceIssue = diskSpaceIssues[0]
				Expect(diskSpaceIssue.DiskType).To(Equal("Persistent"))
				Expect(diskSpaceIssue.NodeName).To(Equal(nodeName + "/" + uuid))
			})
		})

		Context("when a persistent disk has enough space", func() {
			It("does not get returned as a DiskSpaceIssue", func() {
				nodeDiskInfos := createNodeDiskInfos(Params{
					nodeName:             "nodeName",
					uuid:                 "uuid",
					persistentBytesFree:  10,
					persistentInodesFree: 20,
					ephemeralBytesFree:   100,
					ephemeralInodesFree:  100,
				})
				diskSpaceIssues := disk.ValidateCapacity(nodeDiskInfos, thresholdConfig)

				Expect(diskSpaceIssues).To(BeEmpty())
			})
		})

		Context("when multiple nodes do not have enough space", func() {
			It("returns the each node as a DiskSpaceIssue", func() {
				node1DiskInfos := createNodeDiskInfos(Params{
					nodeName:             "node1",
					uuid:                 "uuid1",
					persistentBytesFree:  9,
					persistentInodesFree: 19,
					ephemeralBytesFree:   100,
					ephemeralInodesFree:  100,
				})

				node2DiskInfos := createNodeDiskInfos(Params{
					nodeName:             "node2",
					uuid:                 "uuid2",
					persistentBytesFree:  100,
					persistentInodesFree: 100,
					ephemeralBytesFree:   9,
					ephemeralInodesFree:  19,
				})

				allNodeDiskInfos := append(node1DiskInfos, node2DiskInfos...)

				diskSpaceIssues := disk.ValidateCapacity(allNodeDiskInfos, thresholdConfig)

				Expect(len(diskSpaceIssues)).To(Equal(2))
				issue1 := diskSpaceIssues[0]
				Expect(issue1.NodeName).To(Equal("node1/uuid1"))
				Expect(issue1.DiskType).To(Equal("Persistent"))

				issue2 := diskSpaceIssues[1]
				Expect(issue2.NodeName).To(Equal("node2/uuid2"))
				Expect(issue2.DiskType).To(Equal("Ephemeral"))
			})
		})
	})
})
