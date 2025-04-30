package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/mysql-monitoring-release/spec/utilities/cmd"
	"sort"
	"strings"

	. "github.com/onsi/gomega"
)

type MatchInstanceFunc func(instance Instance) bool

type Instance struct {
	IP       string `json:"ips"`
	Instance string `json:"instance"`
	Index    string `json:"index"`
	VMCid    string `json:"vm_cid"`
}

func (i *Instance) Id() string {
	components := strings.SplitN(i.Instance, "/", 2)
	Expect(components).To(HaveLen(2))
	return components[1]
}

// MatchByIndexedName matches by comparing the provided name to INSTANCE-GROUP/INDEX
func MatchByIndexedName(name string) MatchInstanceFunc {
	return func(i Instance) bool {
		components := strings.SplitN(i.Instance, "/", 2)
		instanceGroup := components[0]
		return instanceGroup+"/"+i.Index == name
	}
}

func Instances(deploymentName string, matchInstanceFunc MatchInstanceFunc) ([]Instance, error) {
	var output bytes.Buffer

	fmt.Printf("deploymentName: %s\n", deploymentName)
	if err := cmd.RunWithoutOutput(&output,
		"bosh",
		"--non-interactive",
		"--tty",
		"--deployment="+deploymentName,
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

	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].Index < instances[j].Index
	})

	return instances, nil
}

func RemoteCommand(deploymentName, instanceSpec, cmdString string) (string, error) {
	var output bytes.Buffer
	err := cmd.RunWithoutOutput(&output,
		"bosh",
		"--deployment="+deploymentName,
		"ssh",
		instanceSpec,
		"--column=Stdout",
		"--results",
		"--command="+cmdString,
	)
	return strings.TrimSpace(output.String()), err
}
