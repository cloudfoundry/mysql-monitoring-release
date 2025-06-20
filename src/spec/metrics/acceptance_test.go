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

	When("disk I/O activity occurs", func() {
		AfterEach(func() {
			// Clean up test files
			_, err := bosh.RemoteCommand(deployment, "mysql/0", "sudo rm -f /var/vcap/store/io-test-* /var/vcap/data/io-test-*")
			Expect(err).NotTo(HaveOccurred())
		})

		It("disk I/O metrics are emitted and have reasonable values", func() {
			workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Microsecond, func() {
				instances, err := bosh.Instances(deployment, bosh.MatchByIndexedName("mysql/0"))
				Expect(err).NotTo(HaveOccurred())
				Expect(instances).To(HaveLen(1))
				instanceUUID := instances[0].Id()

				By("Performing intensive concurrent direct I/O operations to generate high IOPS and latency")
				// Generate sustained concurrent I/O load to maximize IOPS measurements

				// Start multiple concurrent write operations in background for sustained IOPS
				_, err = bosh.RemoteCommand(deployment, "mysql/0", `
					# Concurrent persistent disk operations (background)
					sudo dd if=/dev/zero of=/var/vcap/store/io-test-concurrent-1 bs=4K count=10000 oflag=direct 2>/dev/null &
					sudo dd if=/dev/zero of=/var/vcap/store/io-test-concurrent-2 bs=4K count=10000 oflag=direct 2>/dev/null &
					sudo dd if=/dev/zero of=/var/vcap/store/io-test-concurrent-3 bs=8K count=8000 oflag=direct 2>/dev/null &
					sudo dd if=/dev/zero of=/var/vcap/store/io-test-concurrent-4 bs=8K count=8000 oflag=direct 2>/dev/null &
					
					# Concurrent ephemeral disk operations (background)
					sudo dd if=/dev/zero of=/var/vcap/data/io-test-concurrent-1 bs=4K count=15000 oflag=direct 2>/dev/null &
					sudo dd if=/dev/zero of=/var/vcap/data/io-test-concurrent-2 bs=4K count=15000 oflag=direct 2>/dev/null &
					sudo dd if=/dev/zero of=/var/vcap/data/io-test-concurrent-3 bs=8K count=10000 oflag=direct 2>/dev/null &
					sudo dd if=/dev/zero of=/var/vcap/data/io-test-concurrent-4 bs=8K count=10000 oflag=direct 2>/dev/null &
					
					# Wait for some concurrent operations to complete
					wait
				`)
				Expect(err).NotTo(HaveOccurred())

				// Create base files for intensive random read operations
				_, err = bosh.RemoteCommand(deployment, "mysql/0",
					"sudo dd if=/dev/urandom of=/var/vcap/store/io-test-read-base bs=1M count=200 oflag=direct 2>/dev/null")
				Expect(err).NotTo(HaveOccurred())

				_, err = bosh.RemoteCommand(deployment, "mysql/0",
					"sudo dd if=/dev/urandom of=/var/vcap/data/io-test-read-base bs=1M count=400 oflag=direct 2>/dev/null")
				Expect(err).NotTo(HaveOccurred())

				// Start sustained concurrent read/write workload for maximum IOPS
				_, err = bosh.RemoteCommand(deployment, "mysql/0", `
					# Background sustained read operations with different patterns
					sudo dd if=/var/vcap/store/io-test-read-base of=/dev/null bs=4K count=50000 iflag=direct 2>/dev/null &
					sudo dd if=/var/vcap/store/io-test-read-base of=/dev/null bs=4K skip=100 count=40000 iflag=direct 2>/dev/null &
					sudo dd if=/var/vcap/data/io-test-read-base of=/dev/null bs=4K count=80000 iflag=direct 2>/dev/null &
					sudo dd if=/var/vcap/data/io-test-read-base of=/dev/null bs=4K skip=200 count=60000 iflag=direct 2>/dev/null &
					
					# Concurrent small random writes for high IOPS
					for i in {1..6}; do
						sudo dd if=/dev/urandom of=/var/vcap/store/io-test-small-rand-$i bs=4K count=5000 oflag=direct 2>/dev/null &
					done
					
					for i in {1..8}; do
						sudo dd if=/dev/urandom of=/var/vcap/data/io-test-small-rand-$i bs=4K count=7500 oflag=direct 2>/dev/null &
					done
					
					# Let concurrent operations run for sustained I/O
					sleep 10
					
					# Add even more concurrent random I/O while previous operations continue
					for i in {7..10}; do
						sudo dd if=/dev/urandom of=/var/vcap/store/io-test-extra-$i bs=4K count=3000 oflag=direct 2>/dev/null &
					done
					
					for i in {9..14}; do
						sudo dd if=/dev/urandom of=/var/vcap/data/io-test-extra-$i bs=4K count=4000 oflag=direct 2>/dev/null &
					done
					
					# Wait for some operations to complete but leave sustained load
					sleep 15
				`)
				Expect(err).NotTo(HaveOccurred())

				By("Verifying disk I/O metrics are present and have reasonable values")
				Eventually(func() bool {
					// Check for persistent disk I/O metrics
					persistentReadLatency := queryIOMetric(SourceID, instanceUUID, "persistent_disk_read_latency_ms")
					persistentWriteLatency := queryIOMetric(SourceID, instanceUUID, "persistent_disk_write_latency_ms")
					persistentReadIOPS := queryIOMetric(SourceID, instanceUUID, "persistent_disk_read_iops")
					persistentWriteIOPS := queryIOMetric(SourceID, instanceUUID, "persistent_disk_write_iops")

					// Check for ephemeral disk I/O metrics
					ephemeralReadLatency := queryIOMetric(SourceID, instanceUUID, "ephemeral_disk_read_latency_ms")
					ephemeralWriteLatency := queryIOMetric(SourceID, instanceUUID, "ephemeral_disk_write_latency_ms")
					ephemeralReadIOPS := queryIOMetric(SourceID, instanceUUID, "ephemeral_disk_read_iops")
					ephemeralWriteIOPS := queryIOMetric(SourceID, instanceUUID, "ephemeral_disk_write_iops")

					// Print current metric values for debugging
					GinkgoWriter.Printf("=== Disk I/O Metrics Values ===\n")
					GinkgoWriter.Printf("Persistent Disk - Read Latency: %.2f ms, Write Latency: %.2f ms\n",
						persistentReadLatency, persistentWriteLatency)
					GinkgoWriter.Printf("Persistent Disk - Read IOPS: %.2f, Write IOPS: %.2f\n",
						persistentReadIOPS, persistentWriteIOPS)
					GinkgoWriter.Printf("Ephemeral Disk - Read Latency: %.2f ms, Write Latency: %.2f ms\n",
						ephemeralReadLatency, ephemeralWriteLatency)
					GinkgoWriter.Printf("Ephemeral Disk - Read IOPS: %.2f, Write IOPS: %.2f\n",
						ephemeralReadIOPS, ephemeralWriteIOPS)
					GinkgoWriter.Printf("===============================\n")

					// Verify metrics are present (non-negative) and have reasonable values
					// Latency should be >= 0 and < 1000ms (1 second)
					// IOPS should be >= 0 and < 100000 (100k IOPS is very high but reasonable upper bound)
					return persistentReadLatency >= 0 && persistentReadLatency < 1000 &&
						persistentWriteLatency >= 0 && persistentWriteLatency < 1000 &&
						persistentReadIOPS >= 0 && persistentReadIOPS < 10000 &&
						persistentWriteIOPS >= 0 && persistentWriteIOPS < 10000 &&
						ephemeralReadLatency >= 0 && ephemeralReadLatency < 1000 &&
						ephemeralWriteLatency >= 0 && ephemeralWriteLatency < 1000 &&
						ephemeralReadIOPS >= 0 && ephemeralReadIOPS < 10000 &&
						ephemeralWriteIOPS >= 0 && ephemeralWriteIOPS < 10000
				}, "90s", "5s").Should(BeTrue(), "Disk I/O metrics should be present and within reasonable ranges")
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

func queryIOMetric(sourceId, instanceUUID, metricName string) float64 {
	GinkgoHelper()

	promqlMetricName := sanitizeMetricName(fmt.Sprintf("/%s/system/%s", sourceId, metricName))
	promqlQuery := fmt.Sprintf("%s{source_id=%q,index=%q}", promqlMetricName, sourceId, instanceUUID)

	session := cf.Cf("query", promqlQuery).Wait(10 * time.Second)

	var query Query
	err := json.Unmarshal(session.Out.Contents(), &query)
	Expect(err).NotTo(HaveOccurred())

	if query.Data.Result == nil {
		return 0.0
	}

	return query.ValueAsFloat()
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

func (q Query) ValueAsFloat() float64 {
	GinkgoHelper()
	value, err := q.Data.Result[0].Value[1].Float64()
	Expect(err).NotTo(HaveOccurred())
	return value
}
