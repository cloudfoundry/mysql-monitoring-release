package ui_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "mysql-diag/config"
	. "mysql-diag/diskspaceissue"
	"mysql-diag/ui"
)

var (
	isCanaryHealthy bool
	needsBootstrap  bool
	diskSpaceIssues []DiskSpaceIssue
	messages        []string
	config          *Config
)

var _ = Describe("Reporter", func() {
	BeforeEach(func() {
		isCanaryHealthy = true
		needsBootstrap = false
		diskSpaceIssues = []DiskSpaceIssue{}

		config = &Config{
			Canary: nil,
			Mysql: MysqlConfig{
				Nodes: []MysqlNode{},
			},
		}
	})

	Context("when everything is healthy", func() {
		It("remains quiet", func() {
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy: isCanaryHealthy,
				NeedsBootstrap:  needsBootstrap,
				DiskSpaceIssues: diskSpaceIssues,
			}, config)

			Expect(messages).To(BeEmpty())
		})
	})

	Context("when canary is unhealthy", func() {
		BeforeEach(func() {
			isCanaryHealthy = false
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy: isCanaryHealthy,
				NeedsBootstrap:  needsBootstrap,
				DiskSpaceIssues: diskSpaceIssues,
			}, config)
		})

		It("chirps", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] The replication process is unhealthy")))
		})

		It("gives suggestion to download logs", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] Run the download-logs command:")))
		})

		It("warns not to recreate", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] NOT RECOMMENDED")))
		})
	})

	Context("when needing bootstrap", func() {
		BeforeEach(func() {
			needsBootstrap = true
			messages = ui.Report(ui.ReporterParams{
				IsCanaryHealthy: isCanaryHealthy,
				NeedsBootstrap:  needsBootstrap,
				DiskSpaceIssues: diskSpaceIssues,
			}, config)
		})

		It("gives link to bootstrap instructions", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] You must bootstrap the cluster.")))
		})

		It("gives suggestion to download logs", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[CRITICAL\\] Run the download-logs command:")))
		})

		It("warns not to recreate", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] NOT RECOMMENDED")))
		})
	})

	Context("when there are disk space issues", func() {
		BeforeEach(func() {
			diskSpaceIssues = []DiskSpaceIssue{
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
				IsCanaryHealthy: isCanaryHealthy,
				NeedsBootstrap:  needsBootstrap,
				DiskSpaceIssues: diskSpaceIssues,
			}, config)
		})

		It("renders a warning for the user", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] Foobar disk usage is very high on node SomeNode" +
				".*Consider re-deploying with larger Foobar disks.")))
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] Baz disk usage is very high on node OtherNode" +
				".*Consider re-deploying with larger Baz disks.")))
		})

		It("does not suggest downloading logs", func() {
			Expect(messages).ToNot(ContainElement(MatchRegexp("\\[CRITICAL\\] Run the download-logs command")))
		})

		It("warns us to not do silly things", func() {
			Expect(messages).To(ContainElement(MatchRegexp("\\[WARNING\\] NOT RECOMMENDED")))
		})
	})

	Context("when everything is wrong", func() {
		BeforeEach(func() {
			isCanaryHealthy = false
			needsBootstrap = true
			diskSpaceIssues = []DiskSpaceIssue{
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
				IsCanaryHealthy: isCanaryHealthy,
				NeedsBootstrap:  needsBootstrap,
				DiskSpaceIssues: diskSpaceIssues,
			}, config)
		})

		It("should not duplicate warning messages", func() {
			Expect(len(messages)).To(Equal(6))
		})
	})
})
