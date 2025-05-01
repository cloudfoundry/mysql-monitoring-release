package acceptance_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/cloudfoundry/mysql-monitoring-release/spec/utilities/bosh"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Metrics are received", func() {
	It("correct metrics are emitted within 40s", func() {
		workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
			session := cf.Cf("tail", "-f", SourceID)
			Eventually(session.Out, 40*time.Second).Should(gbytes.Say("/" + SourceID + "/performance/questions:"))
			Eventually(session.Out, 40*time.Second).Should(gbytes.Say("/" + SourceID + "/galera/wsrep_cluster_status:1"))

			session.Terminate()
		})
	})

	It("has unique wsrep_local_index values for each node", func() {
		workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
			session := cf.Cf("tail", "-f", SourceID, "--name-filter=/"+SourceID+"/galera/wsrep_local_index")
			Eventually(session, 40*time.Second).Should(SatisfyAll(
				gbytes.Say("/"+SourceID+"/galera/wsrep_local_index:0"),
				gbytes.Say("/"+SourceID+"/galera/wsrep_local_index:1"),
				gbytes.Say("/"+SourceID+"/galera/wsrep_local_index:2"),
			))
			session.Terminate()
		})
	})

	Context("when ephemeral disk (/var/vcap/data) usage increases", func() {
		AfterEach(func() {
			_, err := bosh.RemoteCommand(deployment, "mysql/0", "sudo rm /var/vcap/data/file")
			Expect(err).NotTo(HaveOccurred())
		})

		It("the corresponding metric increases", func() {
			var (
				availBytes              int
				initialDiskUsePercent   int64
				expectedDiskUsageChange int
			)

			By("Fetch available space on mysql/0 ephemeral disk", func() {
				statOutput, err := bosh.RemoteCommand(deployment, "mysql/0", "sudo stat -f /var/vcap/data")
				Expect(err).NotTo(HaveOccurred())

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
					instances, err := bosh.Instances(deployment, bosh.MatchByIndexedName("mysql/0"))
					Expect(err).NotTo(HaveOccurred())
					Expect(instances).To(HaveLen(1))
					instanceUUID := instances[0].Id()

					Eventually(func() int64 {
						initialDiskUsePercent = queryEphemeralDisk(SourceID, instanceUUID)
						return initialDiskUsePercent
					}, "60s", "5s").Should(BeNumerically(">=", 0))
				})
			})

			By("Writing 50% of available ephemeral disk", func() {
				amountToAllocate := availBytes / 2
				vmCommand := fmt.Sprintf("sudo fallocate -l%d /var/vcap/data/file", amountToAllocate)
				_, err := bosh.RemoteCommand(deployment, "mysql/0", vmCommand)
				Expect(err).NotTo(HaveOccurred())
			})

			By("Capturing the ephemeral disk usage metric after allocating a file", func() {
				workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
					instances, err := bosh.Instances(deployment, bosh.MatchByIndexedName("mysql/0"))
					Expect(err).NotTo(HaveOccurred())
					Expect(instances).To(HaveLen(1))
					instanceUUID := instances[0].Id()

					Eventually(func() int64 {
						finalDiskUsePercent := queryEphemeralDisk(SourceID, instanceUUID)

						return finalDiskUsePercent - initialDiskUsePercent
					}, "60s", "5s").Should(BeNumerically(">=", expectedDiskUsageChange-2),
						`Expected the increase in percent ephemeral disk used to close to or greater than the expected percent to write.
Within 2 percent is acceptable because of rounding.`,
					)
				})
			})
		})
	})
})

func extractIntMatchingRegex(source string, regexMatch string) int {
	regexMatches := regexp.MustCompile(regexMatch).FindStringSubmatch(source)
	Expect(len(regexMatches)).To(Equal(2))
	match, err := strconv.Atoi(regexMatches[1])
	Expect(err).NotTo(HaveOccurred())
	return match
}

func queryEphemeralDisk(sourceId, instanceUUID string) int64 {
	GinkgoHelper()

	promqlMetricName := sanitizeMetricName(fmt.Sprintf("/%s/system/ephemeral_disk_used_percent", sourceId))
	promqlQuery := fmt.Sprintf("%s{source_id=%q,index=%q}", promqlMetricName, sourceId, instanceUUID)

	session := cf.Cf("query", promqlQuery).Wait(10 * time.Second)

	var query Query
	err := json.Unmarshal(session.Out.Contents(), &query)
	Expect(err).NotTo(HaveOccurred())

	if query.Data.Result == nil {
		return 0
	}

	return query.ValueAsInt()
}

// github.com/cloudfoundry/log-cache-release/src/internal/promql/promql.go
func sanitizeMetricName(name string) string {
	GinkgoHelper()
	// Forcefully convert all invalid separators to underscores
	// First character: Match the if it's NOT A-z or underscore ^[^A-z_]
	// All others: Match if they're NOT alphanumeric or understore [\W_]+?
	var re = regexp.MustCompile(`^[^A-z_]|[\W_]+?`)
	return re.ReplaceAllString(name, "_")
}

type Query struct {
	Data struct {
		Result []struct {
			Value []json.Number `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (q Query) ValueAsInt() int64 {
	GinkgoHelper()
	value, err := q.Data.Result[0].Value[1].Int64()
	Expect(err).NotTo(HaveOccurred())
	return value
}
