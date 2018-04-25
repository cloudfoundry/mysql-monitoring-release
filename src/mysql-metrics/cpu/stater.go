package cpu

import (
	"bufio"
	"errors"
	"io"
)

var ErrNoPreviousData = errors.New("No Previous Data")

type Stater struct {
	procStatFile io.ReadSeeker
	previousStat CPUStat
}

func New(procStatFile io.ReadSeeker) Stater {
	return Stater{
		procStatFile: procStatFile,
	}
}

func (s *Stater) GetPercentage() (int, error) {
	if _, err := s.procStatFile.Seek(0, 0); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(s.procStatFile)
	ok := scanner.Scan()

	if !ok {
		return -1, errors.New("failed to read /proc/stat file")
	}

	text := scanner.Text()
	currentCPUStats, err := NewCPUStat(text)
	if err != nil {
		return -1, err
	}

	previousCPUStats := s.previousStat
	s.previousStat = currentCPUStats

	if previousCPUStats == (CPUStat{}) {
		return -1, ErrNoPreviousData
	}

	prevTotal := previousCPUStats.User + previousCPUStats.System + previousCPUStats.Idle + previousCPUStats.Nice + previousCPUStats.IOWait + previousCPUStats.IRQ + previousCPUStats.SoftIRQ + previousCPUStats.Steal
	currTotal := currentCPUStats.User + currentCPUStats.System + currentCPUStats.Idle + currentCPUStats.Nice + currentCPUStats.IOWait + currentCPUStats.IRQ + currentCPUStats.SoftIRQ + currentCPUStats.Steal

	totald := currTotal - prevTotal
	idled := (currentCPUStats.Idle + currentCPUStats.IOWait) - (previousCPUStats.Idle + previousCPUStats.IOWait)
	cpuPercentageTotal := float64(totald-idled) / float64(totald)

	return int(cpuPercentageTotal * 100), nil
}
