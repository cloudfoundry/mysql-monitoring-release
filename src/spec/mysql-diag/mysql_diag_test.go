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

	"github.com/cloudfoundry/mysql-monitoring-release/spec/utilities/bosh"
)

var _ = Describe("MySQLDiag", Ordered, func() {
	var deploymentName = os.Getenv("BOSH_DEPLOYMENT")

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
					gbytes.Say(`INSTANCE\s+\|\s+STATE\s+\|\s+CLUSTER STATUS`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+N/A - ERROR\s+\|\s+N/A - ERROR\s+\|`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+N/A - ERROR\s+\|\s+N/A - ERROR\s+\|`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+N/A - ERROR\s+\|\s+N/A - ERROR\s+\|`),
				))
			})

			By("monitoring the bootstrap process in the background")
			go func() {
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				for range ticker.C {
					select {
					case <-ctx.Done():
						return
					default:
						_ = runMySQLDiag(ctx, withStdout(io.Discard))
					}
				}
			}()

			By("Bootstrapping the cluster again")
			runErrand(ctx, "bootstrap", "mysql/0")

			By("emitting diagnostic output of a healthy cluster", func() {
				output := gbytes.NewBuffer()
				Expect(runMySQLDiag(ctx, withStdout(output))).To(Succeed())
				Expect(output).To(SatisfyAll(
					gbytes.Say(`INSTANCE\s+\|\s+STATE\s+\|\s+CLUSTER STATUS`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
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
				Expect(bosh.RemoteCommand(deploymentName, downInstance, `sudo kill -s STOP $(pidof mysqld)`)).
					Error().NotTo(HaveOccurred())
				DeferCleanup(func() {
					Expect(bosh.RemoteCommand(deploymentName, downInstance, `sudo kill -s CONT $(pidof mysqld)`)).
						Error().NotTo(HaveOccurred())
					Expect(bosh.RemoteCommand(deploymentName, downInstance, `sudo monit restart galera-init`)).
						Error().NotTo(HaveOccurred())
					Eventually(func() string {
						output, _ := bosh.RemoteCommand(deploymentName, downInstance, `sudo monit summary | grep galera-init`)
						return output
					}, "2m", "1s").Should(MatchRegexp(`Process 'galera-init'\s+running`))
				})
			})

			It("emits diagnostic output", func() {
				output := gbytes.NewBuffer()
				Expect(runMySQLDiag(context.Background(), withStdout(output))).To(Succeed())
				Expect(output).To(SatisfyAll(
					gbytes.Say(`error retrieving galera status from node `+downInstance),
					gbytes.Say(`INSTANCE\s+\|\s+STATE\s+\|\s+CLUSTER STATUS`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
					gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+N/A - ERROR\s+\|\s+N/A - ERROR\s+\|`),
				))
			})
		})

		When("a node initiates SST while mysql-diag is running", func() {
			BeforeEach(func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				By("continuously running mysql-diag in the background")
				go func() {
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					for range ticker.C {
						select {
						case <-ctx.Done():
							return
						default:
							_ = runMySQLDiag(ctx, withStdout(io.Discard))
						}
					}
				}()
			})

			It("completes SST and rejoins the cluster", func() {
				var downInstance string
				instances, err := Instances(MatchByInstanceGroup("mysql"))
				Expect(err).NotTo(HaveOccurred())
				downInstance = instances[0].Instance

				By("stopping galera-init on " + downInstance)
				Expect(bosh.RemoteCommand(deploymentName, downInstance, `sudo monit stop galera-init`)).
					Error().NotTo(HaveOccurred())
				Eventually(func() string {
					output, _ := bosh.RemoteCommand(deploymentName, downInstance, `sudo monit summary | grep galera-init`)
					return output
				}, "2m", "5s").Should(MatchRegexp(`Process 'galera-init'\s+not monitored`))
				DeferCleanup(func() {
					Expect(bosh.RemoteCommand(deploymentName, downInstance, `sudo monit restart galera-init`)).
						Error().NotTo(HaveOccurred())
					Eventually(func() string {
						output, _ := bosh.RemoteCommand(deploymentName, downInstance, `sudo monit summary | grep galera-init`)
						return output
					}, "2m", "1s").Should(MatchRegexp(`Process 'galera-init'\s+running`))
				})

				By("purging state information to trigger SST upon node restart")
				_, err = bosh.RemoteCommand(deploymentName, downInstance, `sudo bash -c 'rm -rf /var/vcap/store/pxc-mysql/grastate.dat'`)
				Expect(err).NotTo(HaveOccurred())

				By("(re)starting galera-init on " + downInstance)
				Eventually(func() error {
					_, e := bosh.RemoteCommand(deploymentName, downInstance, `sudo monit start galera-init`)
					return e
				}, "2m", "5s").Should(Succeed(), "galera-init failed to start")

				By("seeing " + downInstance + " has rejoined the cluster successfully")
				Eventually(func() string {
					output, _ := bosh.RemoteCommand(deploymentName, downInstance, `sudo monit summary | grep galera-init`)
					return output
				}, "2m", "5s").Should(MatchRegexp(`Process 'galera-init'\s+running`))

				By("seeing diagnostic output showing the previously-stopped node is synced")
				Eventually(func(g Gomega) {
					output := gbytes.NewBuffer()
					g.Expect(runMySQLDiag(context.Background(), withStdout(output))).To(Succeed())
					g.Expect(output).To(SatisfyAll(
						gbytes.Say(`INSTANCE\s+\|\s+STATE\s+\|\s+CLUSTER STATUS`),
						gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
						gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
						gbytes.Say(`mysql/[0-9a-f-]+\s+\|\s+Synced\s+\|\s+Primary\s+\|`),
					))
				}, "5m", "5s").Should(Succeed())
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
	GinkgoHelper()

	cmd := exec.CommandContext(ctx, "bosh", "ssh", "mysql-monitor", "--command=mysql-diag")
	cmd.Env = append(os.Environ(), "BOSH_DEPLOYMENT="+os.Getenv("BOSH_DEPLOYMENT"))
	cmd.Stderr = GinkgoWriter
	cmd.Stdout = GinkgoWriter

	for _, option := range options {
		option(cmd)
	}

	GinkgoWriter.Println("$", strings.Join(cmd.Args, " "))
	GinkgoWriter.Println()

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
