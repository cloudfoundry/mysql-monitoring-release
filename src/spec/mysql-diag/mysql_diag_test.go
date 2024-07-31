package mysql_diag_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/mysql-monitoring-release/spec/testhelpers"
)

var _ = Describe("MySQLDiag", Ordered, func() {

	When("the cluster is offline", func() {
		var instances []Instance
		BeforeAll(func() {
			var err error
			instances, err = Instances(MatchByInstanceGroup("mysql"))
			Expect(err).NotTo(HaveOccurred())
		})
		It("can bootstrap the cluster", func() {
			By("taking the cluster offline")
			for _, i := range instances {
				By("taking " + i.Instance + " offline.")
				args := []string{"ssh", i.Instance, "--command=\"sudo monit stop galera-init\""}
				session := testhelpers.ExecuteBosh(args, 10*time.Second)
				Expect(session.ExitCode()).To(BeZero())
			}

			By("emitting diagnostic output")
			args := []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
			session := testhelpers.ExecuteBosh(args, 1*time.Minute)
			Expect(session.ExitCode()).To(BeZero())
			Expect(session).To(SatisfyAll(
				gbytes.Say(`(Checking canary status\.\.\. .*healthy.*)|(Canary not configured)`),
				gbytes.Say(`SEQNO\s+|\s+PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
				gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
				gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
				gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
			))

			By("Finding the node to bootstrap")
			args = []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
			session = testhelpers.ExecuteBosh(args, 1*time.Minute)
			Expect(session.ExitCode()).To(BeZero())
			boostrapNode := ""
			for _, line := range strings.Split(string(session.Out.Contents()), "\n") {
				if strings.Contains(line, "Bootstrap node:") {
					boostrapNode = strings.Split(line, ":")[2]
					boostrapNode = strings.TrimSpace(boostrapNode)
					// remove bold and quotes
					boostrapNode = boostrapNode[1 : len(boostrapNode)-5]
				}
			}

			By("Bootstrapping the node: " + boostrapNode)
			args = []string{"-d", os.Getenv("BOSH_DEPLOYMENT"), "ssh", boostrapNode, `--command="sudo bash -c \"echo -n 'NEEDS_BOOTSTRAP' > /var/vcap/store/pxc-mysql/state.txt\""`}
			session = testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())

			args = []string{"ssh", boostrapNode, "--command=\"sudo monit start galera-init\""}
			session = testhelpers.ExecuteBosh(args, 10*time.Second)
			Expect(session.ExitCode()).To(BeZero())

			Eventually(func() *gbytes.Buffer {
				args = []string{"ssh", boostrapNode, "--command=\"sudo monit summary | grep galera-init\""}
				session = testhelpers.ExecuteBosh(args, 10*time.Second)
				return session.Out
			}, "2m", "1s").Should(gbytes.Say(`Process 'galera-init'\s+running`))

			for _, i := range instances {
				if i.Instance != boostrapNode {
					By("Starting the remaining node: " + i.Instance)
					args := []string{"ssh", i.Instance, "--command=\"sudo monit start galera-init\""}
					session := testhelpers.ExecuteBosh(args, 10*time.Second)
					Expect(session.ExitCode()).To(BeZero())
					Eventually(func() *gbytes.Buffer {
						args = []string{"ssh", i.Instance, "--command=\"sudo monit summary | grep galera-init\""}
						session = testhelpers.ExecuteBosh(args, 10*time.Second)
						return session.Out
					}, "2m", "1s").Should(gbytes.Say(`Process 'galera-init'\s+running`))
				}
			}

			By("emitting diagnostic output of a healthy cluster")
			args = []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
			session = testhelpers.ExecuteBosh(args, 1*time.Minute)
			Expect(session.ExitCode()).To(BeZero())
			Expect(session).To(SatisfyAll(
				gbytes.Say(`(Checking canary status\.\.\. .*healthy.*)|(Canary not configured)`),
				gbytes.Say(`SEQNO\s+|\s+PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
				gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
				gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
				gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
			))
		})
	})
	When("the cluster is online", func() {
		When("mysql is not accepting connections", func() {
			BeforeEach(func() {
				By("stopping mysqld on mysql/0")
				args := []string{"ssh", "mysql/0", "sudo kill -s STOP $(pidof mysqld)"}
				session := testhelpers.ExecuteBosh(args, 10*time.Second)
				Expect(session.ExitCode()).To(BeZero())
			})
			AfterEach(func() {
				args := []string{"ssh", "mysql/0", "sudo monit restart galera-init"}
				session := testhelpers.ExecuteBosh(args, 10*time.Second)
				Expect(session.ExitCode()).To(BeZero())

				Eventually(func() *gbytes.Buffer {
					args = []string{"ssh", "mysql/0", "--command=\"sudo monit summary | grep galera-init\""}
					session = testhelpers.ExecuteBosh(args, 10*time.Second)
					return session.Out
				}, "2m", "1s").Should(gbytes.Say(`Process 'galera-init'\s+running`))
			})
			It("emits diagnostic output", func() {
				args := []string{"ssh", "mysql-monitor", "--command=mysql-diag"}
				session := testhelpers.ExecuteBosh(args, 90*time.Second)
				Expect(session.ExitCode()).To(BeZero())
				Expect(session).To(SatisfyAll(
					gbytes.Say(`(Checking canary status\.\.\. .*healthy.*)|(Canary not configured)`),
					gbytes.Say(`SEQNO\s+|\s+PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
					gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
				))
			})
		})
	})
})

type Instance struct {
	Instance string `json:"instance"`
}

type MatchInstanceFunc func(instance Instance) bool

func MatchByInstanceGroup(name string) MatchInstanceFunc {
	return func(i Instance) bool {
		components := strings.SplitN(i.Instance, "/", 2)
		return components[0] == name
	}
}

func RunWithoutOutput(w io.Writer, name string, args ...string) error {
	defer GinkgoWriter.Println()
	cmd := exec.Command(name, args...)
	cmd.Stderr = GinkgoWriter
	cmd.Stdout = w

	GinkgoWriter.Println("$", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func Instances(matchInstanceFunc MatchInstanceFunc) ([]Instance, error) {
	var output bytes.Buffer

	if err := RunWithoutOutput(&output,
		"bosh",
		"--non-interactive",
		"--tty",
		"instances",
		"--details",
		"--json",
	); err != nil {
		return nil, err
	}

	var result struct {
		Tables []struct {
			Rows []Instance
		}
	}

	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to decode bosh instances output: %v", err)
	}

	var instances []Instance

	for _, row := range result.Tables[0].Rows {
		if matchInstanceFunc(row) {
			instances = append(instances, row)
		}
	}

	return instances, nil
}
