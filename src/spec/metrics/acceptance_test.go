package acceptance_test

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/mysql-monitoring-release/spec/utilities/bosh"
	"regexp"
	"strconv"
	"strings"
	"time"

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
				initialDiskUsePercent   int
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

					promQL := `_` + strings.ReplaceAll(SourceID, "-", "_") + `_system_ephemeral_disk_used_percent{source_id="` + SourceID + `",index="` + instances[0].Id() + `"}`

					var query Query
					Eventually(func() int {
						session := cf.Cf("query", promQL).Wait(10 * time.Second)

						err := json.Unmarshal(session.Out.Contents(), &query)
						Expect(err).NotTo(HaveOccurred())

						if query.Data.Result == nil {
							return 0
						}

						return query.ValueAsInt()
					}, "60s", "5s").Should(BeNumerically(">=", 0))

					initialDiskUsePercent = query.ValueAsInt()
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

					promQL := `_` + strings.ReplaceAll(SourceID, "-", "_") + `_system_ephemeral_disk_used_percent{source_id="` + SourceID + `",index="` + instances[0].Id() + `"}`

					var query Query
					Eventually(func() int {
						session := cf.Cf("query", promQL).Wait(10 * time.Second)

						err := json.Unmarshal(session.Out.Contents(), &query)
						Expect(err).NotTo(HaveOccurred())

						if query.Data.Result == nil {
							return 0
						}

						finalDiskUsePercent := query.ValueAsInt()

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

type Query struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Deployment string `json:"deployment"`
				Index      string `json:"index"`
				Ip         string `json:"ip"`
				Job        string `json:"job"`
				Origin     string `json:"origin"`
				SourceId   string `json:"source_id"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (q Query) ValueAsInt() int {
	value, ok := q.Data.Result[0].Value[1].(string)
	Expect(ok).To(BeTrue())

	valueAsInt, err := strconv.Atoi(value)
	Expect(err).NotTo(HaveOccurred())
	return valueAsInt
}
