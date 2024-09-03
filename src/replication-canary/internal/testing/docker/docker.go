package docker

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
)

type ContainerSpec struct {
	Image          string
	Ports          []string
	HealthCmd      string
	HealthInterval string
	Env            []string
	Volumes        []string
	Args           []string
}

func Command(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	out := bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(&out, ginkgo.GinkgoWriter)
	cmd.Stderr = ginkgo.GinkgoWriter
	_, _ = fmt.Fprintln(ginkgo.GinkgoWriter, "$", strings.Join(cmd.Args, " "))
	err := cmd.Run()

	return strings.TrimSpace(out.String()), err
}

func RunContainer(spec ContainerSpec) (string, error) {
	containerID, err := CreateContainer(spec)
	if err != nil {
		return "", err
	}

	return containerID, StartContainer(containerID)
}

func StartContainer(name string) error {
	_, err := Command("start", name)
	return err
}

func CreateContainer(spec ContainerSpec) (string, error) {
	args := []string{
		"create",
		"--pull=always",
	}

	if spec.HealthCmd != "" {
		args = append(args, "--health-cmd="+spec.HealthCmd)
	}

	if spec.HealthInterval != "" {
		args = append(args, "--health-interval="+spec.HealthInterval)
	}

	for _, e := range spec.Env {
		args = append(args, "--env="+e)
	}

	for _, v := range spec.Volumes {
		args = append(args, "--volume="+v)
	}

	for _, p := range spec.Ports {
		args = append(args, "--publish="+p)
	}

	args = append(args, spec.Image)
	args = append(args, spec.Args...)

	return Command(args...)
}

func WaitHealthy(container string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return errors.New("timeout waiting for healthy container")
		case <-ticker.C:
			result, err := Command("container", "inspect", "--format={{.State.Status}} {{.State.Health.Status}}", container)
			if err != nil {
				return fmt.Errorf("error inspecting container: %v", err)
			}

			if strings.HasPrefix(result, "exited ") {
				return fmt.Errorf("container exited")
			}

			if result == "running healthy" {
				return nil
			}
		}
	}
}

func RemoveContainer(name string) error {
	_, err := Command("container", "rm", "--force", "--volumes", name)
	return err
}

func CreateVolume(name string) error {
	_, err := Command("volume", "create", name)
	return err
}

func RemoveVolume(name string) error {
	_, err := Command("volume", "remove", "--force", name)
	return err
}

func Copy(src, dst string) error {
	_, err := Command("container", "cp", "--archive", src, dst)
	return err
}

func ContainerPort(containerID, portSpec string) (string, error) {
	hostPort, err := Command("container", "port", containerID, portSpec)
	if err != nil {
		return "", err
	}

	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", err
	}

	return port, nil
}

func MySQLDB(containerName string) (*sql.DB, error) {
	mysqlPort, err := ContainerPort(containerName, "3306/tcp")
	if err != nil {
		return nil, err
	}

	dsn := "root@tcp(127.0.0.1:" + mysqlPort + ")/"

	return sql.Open("mysql", dsn)
}

func Logs(containerID string) error {
	_, _ = fmt.Fprintln(ginkgo.GinkgoWriter, "$ docker logs", containerID)
	cmd := exec.Command("docker", "logs", containerID)
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return cmd.Run()
}
