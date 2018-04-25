package cpu

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type CPUStat struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
}

func NewCPUStat(stats string) (CPUStat, error) {
	cpuStatsStrings := strings.Fields(stats)
	if cpuStatsStrings[0] != "cpu" {
		return CPUStat{}, errors.New("/proc/stat file does not contain data as expected")
	}
	user, err := strconv.ParseInt(cpuStatsStrings[1], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing user value: %v", err)
	}
	nice, err := strconv.ParseInt(cpuStatsStrings[2], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing nice value: %v", err)
	}
	system, err := strconv.ParseInt(cpuStatsStrings[3], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing system value: %v", err)
	}
	idle, err := strconv.ParseInt(cpuStatsStrings[4], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing idle value: %v", err)
	}
	iowait, err := strconv.ParseInt(cpuStatsStrings[5], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing iowait value: %v", err)
	}
	irq, err := strconv.ParseInt(cpuStatsStrings[6], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing irq value: %v", err)
	}
	softirq, err := strconv.ParseInt(cpuStatsStrings[7], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing softirq value: %v", err)
	}
	steal, err := strconv.ParseInt(cpuStatsStrings[8], 10, 64)
	if err != nil {
		return CPUStat{}, fmt.Errorf("error parsing steal value: %v", err)
	}
	return CPUStat{
		User:    uint64(user),
		Nice:    uint64(nice),
		System:  uint64(system),
		Idle:    uint64(idle),
		IOWait:  uint64(iowait),
		IRQ:     uint64(irq),
		SoftIRQ: uint64(softirq),
		Steal:   uint64(steal),
	}, nil

}
