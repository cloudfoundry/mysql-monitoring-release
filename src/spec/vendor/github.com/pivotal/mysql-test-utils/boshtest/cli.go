package boshtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/commandreporter"
	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
)

type CLI struct {
	Deployment string
	Timeout    time.Duration
	reporter   *commandreporter.CommandReporter
	cnfPath    string
}

type DeploymentOption func(args *[]string)

type Instance struct {
	Group        string
	IP           string
	Index        string
	UUID         string
	VmCid        string
	ProcessState string
}

func NewCLI(deployment string, timeout time.Duration) *CLI {
	return &CLI{
		Deployment: deployment,
		Timeout:    timeout,
		reporter:   commandreporter.NewCommandReporter(ginkgo.GinkgoWriter),
		cnfPath:    "/var/vcap/jobs/mysql/config/mylogin.cnf",
	}
}

func NewCLIWithCnfPath(deployment string, timeout time.Duration, path string) *CLI {
	return &CLI{
		Deployment: deployment,
		Timeout:    timeout,
		reporter:   commandreporter.NewCommandReporter(ginkgo.GinkgoWriter),
		cnfPath:    path,
	}
}
func (c *CLI) DeleteDeployment() error {
	_, err := c.Run("delete-deployment", "--force")
	return errors.Wrapf(err, "failed to delete deployment %q", c.Deployment)
}

func (c *CLI) Deploy(manifestPath string, options ...DeploymentOption) error {
	args := []string{
		"deploy",
		manifestPath,
	}

	for _, opt := range options {
		opt(&args)
	}

	_, err := c.Run(args...)
	return err
}

func (c *CLI) InstanceAddress(instance string) (string, error) {
	instanceParts := strings.Split(instance, "/")
	instanceGroup := instanceParts[0]

	instances, err := c.InstanceGroupByName(instanceGroup)
	if err != nil {
		return "", err
	}

	for _, inst := range instances {
		if inst.Group+"/"+inst.Index == instance {
			return inst.IP, nil
		}
	}

	return "", errors.New("No matching instance found.")
}

func (c *CLI) InstanceGroupByName(instanceGroup string) ([]Instance, error) {
	output, err := c.Run(
		"instances",
		"--details",
		"--column=IPs",
		"--column=Index",
		"--column=Instance",
		"--column='VM CID'",
		"--column='Process State'",
		"--json",
	)

	if err != nil {
		return nil, err
	}

	var result struct {
		Tables []struct {
			Rows []struct {
				IP           string `json:"ips"`
				Index        string `json:"index"`
				Instance     string `json:"instance"`
				VmCid        string `json:"vm_cid"`
				ProcessState string `json:"process_state"`
			}
		}
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, errors.Wrap(err, "failed to decode bosh instances output")
	}

	var instances []Instance
	for _, row := range result.Tables[0].Rows {
		if strings.HasPrefix(row.Instance, instanceGroup) {
			parts := strings.Split(row.Instance, "/")
			group := parts[0]
			uuid := parts[1]
			instances = append(instances, Instance{
				Group:        group,
				IP:           row.IP,
				Index:        row.Index,
				UUID:         uuid,
				VmCid:        row.VmCid,
				ProcessState: row.ProcessState,
			})
		}
	}

	return instances, nil
}

func (c *CLI) Leader() (*Instance, error) {
	instances, err := c.InstanceGroupByName("mysql")
	if err != nil {
		return nil, err
	}

	var leaders []Instance
	for _, inst := range instances {
		readOnly, err := c.MySQLIsReadOnly("mysql/" + inst.Index)
		if err != nil {
			return nil, err
		}

		if !readOnly {
			leaders = append(leaders, inst)
		}
	}

	switch len(leaders) {
	case 0:
		return nil, errors.New("no writable instance found")
	case 1:
		return &leaders[0], nil
	default:
		return nil, errors.New("multiple writable instances found")
	}
}

func (c *CLI) Follower() (*Instance, error) {
	instances, err := c.InstanceGroupByName("mysql")
	if err != nil {
		return nil, err
	}

	var followers []Instance
	for _, inst := range instances {
		slaveStatus, err := c.MySQLQuery("mysql/"+inst.Index, `SHOW SLAVE STATUS\\G`)
		if err != nil {
			return nil, err
		}

		if strings.Contains(slaveStatus, `Slave_IO_State`) {
			followers = append(followers, inst)
		}
	}

	switch len(followers) {
	case 0:
		return nil, errors.New("no follower instances found")
	case 1:
		return &followers[0], nil
	default:
		return nil, errors.New("multiple follower instances found")
	}
}

func (c *CLI) MySQLIsReadOnly(instance string) (bool, error) {
	readOnlyValue, err := c.MySQLQuery(instance, "SELECT @@global.read_only")

	return readOnlyValue == "1", err
}

func (c *CLI) MySQLExec(instance, sql string) error {
	mysqlCmd := strings.Join([]string{
		"sudo",
		"mysql",
		fmt.Sprintf("--defaults-file=%s", c.cnfPath),
		"-ss", // suppress column names and pretty output
		"--execute='" + sql + "'",
	}, " ")

	_, err := c.Run(
		"ssh",
		instance,
		"--results",
		"--column=Stdout",
		"--command="+mysqlCmd,
	)
	return errors.Wrapf(err, "[instance=%q] failed to run mysql query %q", instance, sql)
}

func (c *CLI) MySQLQuery(instance, sql string) (string, error) {
	mysqlCmd := strings.Join([]string{
		"sudo",
		"mysql",
		fmt.Sprintf("--defaults-file=%s", c.cnfPath),
		"-ss", // suppress column names and pretty output
		`--execute="` + sql + `"`,
	}, " ")

	output, err := c.Run(
		"ssh",
		instance,
		"--results",
		"--column=Stdout",
		"--command="+mysqlCmd,
	)
	return output, errors.Wrapf(err, "[instance=%q] failed to run mysql query %q", instance, sql)
}

func (c *CLI) MySQLSchemaExists(instance, schemaName string) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) = 1 FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'`, schemaName)
	output, err := c.MySQLQuery(instance, query)
	if err != nil {
		return false, err
	}

	return output == "1", nil
}

func (c *CLI) MySQLTableExists(instance, schemaName, tableName string) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) = 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'`, schemaName, tableName)
	output, err := c.MySQLQuery(instance, query)
	if err != nil {
		return false, err
	}

	return output == "1", nil
}

func (c *CLI) Restart(instance string) error {
	_, err := c.Run(
		"restart",
		instance,
	)

	return errors.Wrapf(err, "failed to run bosh restart %q", instance)
}

func (c *CLI) Run(args ...string) (string, error) {
	baseArgs := []string{
		"--non-interactive",
		"--deployment=" + c.Deployment,
	}
	boshArgs := append(baseArgs, args...)
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	command := exec.CommandContext(ctx, "bosh", boshArgs...)
	c.reporter.Report(time.Now(), command)

	var stdoutBuf, stderrBuf bytes.Buffer
	command.Stdout = io.MultiWriter(ginkgo.GinkgoWriter, &stdoutBuf)
	command.Stderr = io.MultiWriter(ginkgo.GinkgoWriter, &stderrBuf)

	if err := command.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = errors.Wrap(err, "command timeout exceeded")
		}

		return stdoutBuf.String(), errors.Wrapf(err, "%q failed", args[0])
	}

	return strings.TrimSpace(stdoutBuf.String()), nil
}

func (c *CLI) RunErrand(errandName, instance string) (string, error) {
	output, err := c.Run("run-errand", errandName, "--instance="+instance)

	return output, errors.Wrapf(err, "error when attempting to run bosh errand %q on instance %q", errandName, instance)
}

func (c *CLI) Start(instance string) error {
	_, err := c.Run("start", instance)
	return errors.Wrapf(err, "failed to run bosh start %s", instance)
}

func (c *CLI) Stop(instance string) error {
	_, err := c.Run("stop", instance)
	return errors.Wrapf(err, "failed to run bosh stop %s", instance)
}

func WithOpsFiles(path ...string) DeploymentOption {
	return func(args *[]string) {
		for _, p := range path {
			*args = append(*args, "--ops-file="+p)
		}
	}
}

func WithVars(varKeyValue ...string) DeploymentOption {
	return func(args *[]string) {
		for _, v := range varKeyValue {
			*args = append(*args, "--var="+v)
		}
	}
}

func WithVarsStore(path string) DeploymentOption {
	return func(args *[]string) {
		*args = append(*args, "--vars-store="+path)
	}
}
