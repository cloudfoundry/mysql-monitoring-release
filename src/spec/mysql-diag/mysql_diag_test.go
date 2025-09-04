package mysql_diag_test

import (
	"bytes"
	"context"
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
		It("can bootstrap the cluster", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			By("taking the cluster offline")
			stopJob(ctx, "mysql", "galera-init")

			By("emitting diagnostic output", func() {
				output := gbytes.NewBuffer()
				Expect(runMySQLDiag(ctx, withStdout(output))).To(Succeed())
				Expect(output).To(SatisfyAll(
					gbytes.Say(`SEQNO\s+|\s+PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
					gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
					gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
					gbytes.Say(`\s+[0-9]+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
				))
			})
		})
	})

	When("the cluster is online", func() {
		When("mysql is not accepting connections", func() {
			var instances []Instance
			var downInstance string
			BeforeEach(func() {
				var err error
				instances, err = Instances(MatchByInstanceGroup("mysql"))
				Expect(err).NotTo(HaveOccurred())
				downInstance = instances[0].Instance
				By("stopping mysqld on " + downInstance)
				args := []string{"ssh", downInstance, "-c", "sudo kill -s STOP $(pidof mysqld)"}
				session := testhelpers.ExecuteBosh(args, 10*time.Second)

				Expect(session.ExitCode()).To(BeZero())
			})
			AfterEach(func() {
				args := []string{"ssh", downInstance, "sudo monit restart galera-init"}
				session := testhelpers.ExecuteBosh(args, 10*time.Second)
				Expect(session.ExitCode()).To(BeZero())

				Eventually(func() *gbytes.Buffer {
					args = []string{"ssh", downInstance, "--command=\"sudo monit summary | grep galera-init\""}
					session = testhelpers.ExecuteBosh(args, 10*time.Second)
					return session.Out
				}, "2m", "1s").Should(gbytes.Say(`Process 'galera-init'\s+running`))
			})
			It("emits diagnostic output", func() {
				output := gbytes.NewBuffer()
				Expect(runMySQLDiag(context.Background(), withStdout(output))).To(Succeed())
				Expect(output).To(SatisfyAll(
					gbytes.Say(`error retrieving galera status from node `+downInstance),
					gbytes.Say(`SEQNO\s+|\s+PERSISTENT DISK USED\s+\|\s+EPHEMERAL DISK USED`),
					gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`\s+[0-9]+\s+|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`\s+-1+\s+|\s+N/A - ERROR\s+\|\s+ N/A - ERROR\s+\|`),
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

func runMySQLDiag(ctx context.Context, options ...func(*exec.Cmd)) error {
	cmd := exec.CommandContext(ctx, "bosh", "ssh", "mysql-monitor", "--command=mysql-diag")
	cmd.Env = append(os.Environ(), "BOSH_DEPLOYMENT="+os.Getenv("BOSH_DEPLOYMENT"))
	cmd.Stderr = GinkgoWriter
	cmd.Stdout = GinkgoWriter

	for _, option := range options {
		option(cmd)
	}

	GinkgoWriter.Println("$ ", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func withStdout(w io.Writer) func(*exec.Cmd) {
	return func(cmd *exec.Cmd) {
		cmd.Stdout = w
	}
}

func runErrand(ctx context.Context, errandName, instanceGroup string) {
	GinkgoHelper()

	cmd := exec.CommandContext(ctx, "bosh", "run-errand", errandName, "--instance="+instanceGroup)
	cmd.Env = append(os.Environ(), "BOSH_DEPLOYMENT="+os.Getenv("BOSH_DEPLOYMENT"))
	cmd.Stderr = GinkgoWriter
	cmd.Stdout = GinkgoWriter
	GinkgoWriter.Println("$ ", strings.Join(cmd.Args, " "))
	Expect(cmd.Run()).To(Succeed())
}

func stopJob(ctx context.Context, instanceGroup, jobName string) {
	GinkgoHelper()
	GinkgoHelper()

	cmd := exec.CommandContext(ctx, "bosh", "ssh", instanceGroup, "sudo monit stop "+jobName)
	cmd.Env = append(os.Environ(), "BOSH_DEPLOYMENT="+os.Getenv("BOSH_DEPLOYMENT"))
	cmd.Stderr = GinkgoWriter
	cmd.Stdout = GinkgoWriter

	GinkgoWriter.Println("$ ", strings.Join(cmd.Args, " "))
	Expect(cmd.Run()).To(Succeed())

}
