package ui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/database"
	"github.com/cloudfoundry/mysql-diag/disk"
	"github.com/cloudfoundry/mysql-diag/ui"
)

var (
	isCanaryHealthy     bool
	needsBootstrap      bool
	diskSpaceIssues     []disk.DiskSpaceIssue
	messages            []string
	nodeClusterStatuses []*database.NodeClusterStatus
)

var _ = Describe("Reporter", func() {
	BeforeEach(func() {
		isCanaryHealthy = true
		needsBootstrap = false
		diskSpaceIssues = []disk.DiskSpaceIssue{}
		nodeClusterStatuses = []*database.NodeClusterStatus{
			{
				Node: config.MysqlNode{
					Name: "mysql",
					UUID: "c5522e95-1cdc-4930-9242-e3c0a37a3c2a",
				},
				Status: &database.GaleraStatus{
					LocalIndex:  "befe0c28-b5f4",
					LastApplied: 10,
				},
			},
			{
				Node: config.MysqlNode{
					Name: "mysql",
					UUID: "cf85ed2f-3ec1-4cfe-98aa-1d9c56896ce8",
				},
				Status: &database.GaleraStatus{
					LocalIndex:  "8e9483c8-beed",
					LastApplied: 1000,
				},
			},
		}
	})

	Context("when everything is healthy", func() {
		It("remains quiet except for communicating writeable node", func() {
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy:     isCanaryHealthy,
				NeedsBootstrap:      needsBootstrap,
				DiskSpaceIssues:     diskSpaceIssues,
				NodeClusterStatuses: nodeClusterStatuses,
			})

			Expect(len(messages)).To(Equal(1))
		})
	})

	Context("when canary is unhealthy", func() {
		BeforeEach(func() {
			isCanaryHealthy = false
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy:     isCanaryHealthy,
				NeedsBootstrap:      needsBootstrap,
				DiskSpaceIssues:     diskSpaceIssues,
				NodeClusterStatuses: nodeClusterStatuses,
			})
		})

		It("chirps", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] The replication process is unhealthy")))
		})

		It("gives suggestion to download logs", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] Run the bosh logs command:")))
		})

		It("warns not to recreate", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] NOT RECOMMENDED")))
		})
	})

	Context("when needing bootstrap", func() {
		BeforeEach(func() {
			needsBootstrap = true
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy:     isCanaryHealthy,
				NeedsBootstrap:      needsBootstrap,
				DiskSpaceIssues:     diskSpaceIssues,
				NodeClusterStatuses: nodeClusterStatuses,
			})
		})

		It("gives link to bootstrap instructions", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] You must bootstrap the cluster.")))
		})

		It("gives suggestion to download logs", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] Run the bosh logs command:")))
		})

		It("warns not to recreate", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] NOT RECOMMENDED")))
		})

		It("communicates the node to boostrap", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] Bootstrap node: \"mysql/cf85ed2f-3ec1-4cfe-98aa-1d9c56896ce8\"")))

		})
		It("does not communicate the writeable node", func() {
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy: isCanaryHealthy,
				NeedsBootstrap:  needsBootstrap,
				DiskSpaceIssues: diskSpaceIssues,
				NodeClusterStatuses: []*database.NodeClusterStatus{
					{
						Node: config.MysqlNode{
							Name: "mysql",
							UUID: "c5522e95-1cdc-4930-9242-e3c0a37a3c2a",
						},
						Status: &database.GaleraStatus{
							LocalIndex: "befe0c28-b5f4",
						},
					},
					{
						Node: config.MysqlNode{
							Name: "mysql",
							UUID: "cf85ed2f-3ec1-4cfe-98aa-1d9c56896ce8",
						},
						Status: &database.GaleraStatus{
							LocalIndex: "8e9483c8-beed",
						},
					},
				},
			})

			Expect(messages).To(Not(ContainElement(MatchRegexp("NOTE: Proxies will currently attempt to direct traffic to \".*\""))))
		})
	})
	Context("when there are disk space issues", func() {
		BeforeEach(func() {
			diskSpaceIssues = []disk.DiskSpaceIssue{
				{
					DiskType: "Foobar",
					NodeName: "SomeNode",
				},
				{
					DiskType: "Baz",
					NodeName: "OtherNode",
				},
			}
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy:     isCanaryHealthy,
				NeedsBootstrap:      needsBootstrap,
				DiskSpaceIssues:     diskSpaceIssues,
				NodeClusterStatuses: nodeClusterStatuses,
			})
		})

		It("renders a warning for the user", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] Foobar disk usage is very high on node SomeNode" +
				".*Consider re-deploying with larger Foobar disks.")))
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] Baz disk usage is very high on node OtherNode" +
				".*Consider re-deploying with larger Baz disks.")))
		})

		It("does not suggest downloading logs", func() {
			Expect(messages).ToNot(ContainElement(MatchRegexp("\\[CRITICAL\\] Run the bosh logs command")))
		})

		It("warns us to not do silly things", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] NOT RECOMMENDED")))
		})
	})

	Context("when everything is wrong", func() {
		BeforeEach(func() {
			isCanaryHealthy = false
			needsBootstrap = true
			diskSpaceIssues = []disk.DiskSpaceIssue{
				{
					DiskType: "Foobar",
					NodeName: "SomeNode",
				},
				{
					DiskType: "Baz",
					NodeName: "OtherNode",
				},
			}
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy:     isCanaryHealthy,
				NeedsBootstrap:      needsBootstrap,
				DiskSpaceIssues:     diskSpaceIssues,
				NodeClusterStatuses: nodeClusterStatuses,
			})
		})

		It("should not duplicate warning messages", func() {
			Expect(len(messages)).To(Equal(7))
		})
	})

	It("communicates the writeable node", func() {
		messages = ui.Report(ui.ReporterParams{
			IsCanaryHealthy: isCanaryHealthy,
			NeedsBootstrap:  needsBootstrap,
			DiskSpaceIssues: diskSpaceIssues,
			NodeClusterStatuses: []*database.NodeClusterStatus{
				{
					Node: config.MysqlNode{
						Name: "mysql",
						UUID: "c5522e95-1cdc-4930-9242-e3c0a37a3c2a",
					},
					Status: &database.GaleraStatus{
						LocalIndex: "befe0c28-b5f4",
					},
				},
				{
					Node: config.MysqlNode{
						Name: "mysql",
						UUID: "cf85ed2f-3ec1-4cfe-98aa-1d9c56896ce8",
					},
					Status: &database.GaleraStatus{
						LocalIndex: "8e9483c8-beed",
					},
				},
			},
		})

		Expect(messages).To(ContainElement(MatchRegexp("NOTE: Proxies will currently attempt to direct traffic to \"mysql/cf85ed2f-3ec1-4cfe-98aa-1d9c56896ce8\"")))
	})
})
