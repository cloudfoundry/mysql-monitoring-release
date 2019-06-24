package acceptance_test

import (
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal/mysql-test-utils/testhelpers"
	"regexp"
	"strconv"
	"time"
)

var _ = Describe("Metrics are received", func() {
	It("correct metrics are emitted within 40s", func() {
		workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
			session := cf.Cf("tail", "-f", "p-mysql")
			Eventually(session.Out, 40 * time.Second).Should(gbytes.Say("/p-mysql/performance/questions:"))
			Eventually(session.Out, 40 * time.Second).Should(gbytes.Say("/p-mysql/galera/wsrep_cluster_status:1"))

			session.Terminate()
		})
	})

	Context("when ephemeral disk (/var/vcap/data) usage increases", func() {
		var (
			args    []string
			session *gexec.Session
			ephemeralDiskUsedMetricRegex string
		)

		BeforeEach(func() {
			ephemeralDiskUsedMetricRegex = `/p-mysql/system/ephemeral_disk_used_percent:([0-9]+)`

		})
		AfterEach(func() {
			args = []string{"ssh", "mysql/0", "-c", "sudo rm /var/vcap/data/pxc-mysql/tmp/file"}
			session = testhelpers.ExecuteBosh(args, 20*time.Second)
		})

		FIt("the corresponding metric increases", func() {
			var (
				availBytes              int
				initialDiskUsePercent   int
				finalDiskUsePercent     int
				expectedDiskUsageChange int
			)

			By("Fetch available space on mysql/0 ephemeral disk", func() {
				args = []string{"ssh", "mysql/0", "-c", "sudo stat -f /var/vcap/data"}
				session = testhelpers.ExecuteBosh(args, 10*time.Second)
				Expect(session.ExitCode()).To(BeZero())
				statOutput := session.Out.Contents()

				blockSizeRegex := `Block size:\s+([0-9]+)`
				blockSize := extractIntMatchingRegex(statOutput, blockSizeRegex)

				availBlocksRegex := `Blocks:\s+Total:\s+[0-9]+\s+Free:\s+[0-9]+\s+Available:\s+([0-9]+)`
				availBlocks := extractIntMatchingRegex(statOutput, availBlocksRegex)
				availBytes = blockSize * availBlocks

				totalBlocksRegex := `Blocks:\s+Total:\s+([0-9]+)`
				totalBlocks := extractIntMatchingRegex(statOutput, totalBlocksRegex)

				expectedDiskUsageChange = availBlocks * 100 / (2 * totalBlocks)

			})
			By("Capturing ephemeral disk usage metric before filling the disk", func() {
				workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
					session := cf.Cf("tail", "-f", "p-mysql")
					Eventually(session.Out, 40*time.Second).Should(gbytes.Say(ephemeralDiskUsedMetricRegex))
					ephDiskMetricsOutput := session.Out.Contents()

					initialDiskUsePercent = extractIntMatchingRegex(ephDiskMetricsOutput, ephemeralDiskUsedMetricRegex)
					session.Terminate()
				})
			})

			By("Writing 50% of available ephemeral disk", func() {
				amountToAllocate := availBytes / 2
				vmCommand := fmt.Sprintf("sudo fallocate -l%d /var/vcap/data/pxc-mysql/tmp/file", amountToAllocate)
				args = []string{"ssh", "mysql/0", "-c", vmCommand}
				session = testhelpers.ExecuteBosh(args, 10*time.Second)
				Expect(session.ExitCode()).To(BeZero())
			})

			By("Capturing the ephemeral disk usage metric after allocating a file", func() {
				workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
					session := cf.Cf("tail", "-f", "p-mysql")
					Eventually(session.Out, 40*time.Second).Should(gbytes.Say(ephemeralDiskUsedMetricRegex))
					out := session.Out.Contents()
					finalDiskUsePercent = extractIntMatchingRegex(out, ephemeralDiskUsedMetricRegex)
					session.Terminate()
				})
			})

			By("Having measured a significant change in ephemeral disk usage", func() {
				emittedDiskUsageChange := finalDiskUsePercent - initialDiskUsePercent
				fmt.Printf("emittedDiskUsageChange: %d", emittedDiskUsageChange)
				fmt.Printf("expectedDiskUsageChange: %d", expectedDiskUsageChange)
				Expect(emittedDiskUsageChange).To(
					BeNumerically(">=", expectedDiskUsageChange-2),
					`Expected the increase in percent ephemeral disk used to close to or greater than the expected percent to write.
Within 2 percent is acceptable because of rounding.`,
					)
			})
		})
	})
})

func extractIntMatchingRegex(source []byte, regexMatch string) int {
	regexMatches := regexp.MustCompile(regexMatch).FindSubmatch(source)
	Expect(len(regexMatches)).To(Equal(2))
	match, err := strconv.Atoi(string(regexMatches[1]))
	Expect(err).NotTo(HaveOccurred())
	return match
}
