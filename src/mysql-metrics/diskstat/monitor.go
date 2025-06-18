// Package diskstats provides disk I/O monitoring functionality for Linux systems.
// It reads from /proc/diskstats and /proc/mounts to provide real-time disk statistics
// for filesystem mountpoints, compatible with pt-diskstats output format.
package diskstat

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/blockdevice"
)

// ErrFirstSample is returned by the Sample method when it's the first time
// a given mountpoint is being sampled, as a delta cannot be calculated yet.
var ErrFirstSample = fmt.Errorf("this is the first sample for this mountpoint, no delta available")

// ErrMountPointNotFound is returned when the given path cannot be resolved to
// a known mountpoint.
var ErrMountPointNotFound = fmt.Errorf("could not find a mountpoint for the specified path")

// ProcDiskStatsReader interface abstracts reading /proc/diskstats
type ProcDiskStatsReader interface {
	ProcDiskstats() ([]blockdevice.Diskstats, error)
}

// MountInfoReader interface abstracts reading system mount point information
type MountInfoReader interface {
	GetMounts() ([]*procfs.MountInfo, error)
}

// VolumeMonitor is a monitor for collecting disk I/O statistics
// for block devices based on their filesystem mountpoints.
type VolumeMonitor struct {
	fs         ProcDiskStatsReader
	procfs     MountInfoReader
	prevSample map[string]Sample
	mountCache map[string]string // Cache for mountpoint path -> device name
}

// NewVolumeMonitor creates and initializes a new VolumeMonitor.
func NewVolumeMonitor() (*VolumeMonitor, error) {
	fs, err := blockdevice.NewDefaultFS()
	if err != nil {
		return nil, fmt.Errorf("could not open blockdevice fs: %w", err)
	}
	return &VolumeMonitor{
		fs:         fs,
		procfs:     NewMountReader(),
		prevSample: make(map[string]Sample),
		mountCache: make(map[string]string),
	}, nil
}

// NewVolumeMonitorWithDeps creates a VolumeMonitor with injected dependencies for testing.
func NewVolumeMonitorWithDeps(fs ProcDiskStatsReader, procfs MountInfoReader) *VolumeMonitor {
	return &VolumeMonitor{
		fs:         fs,
		procfs:     procfs,
		prevSample: make(map[string]Sample),
		mountCache: make(map[string]string),
	}
}

// Sample captures the current disk I/O statistics for the device associated
// with the given mountpoint and returns a Delta representing the activity
// since the last call to Sample for the same mountpoint.
func (m *VolumeMonitor) Sample(mountpoint string) (Delta, error) {
	now := time.Now()

	// 1. Resolve the mountpoint to an underlying device name (e.g., "sda").
	deviceName, err := m.resolveDevice(mountpoint)
	if err != nil {
		return Delta{}, err
	}

	// 2. Get all current disk stats.
	allStats, err := m.fs.ProcDiskstats()
	if err != nil {
		return Delta{}, fmt.Errorf("could not read diskstats: %w", err)
	}

	// 3. Find the stats for our specific device.
	var currentStats *blockdevice.Diskstats
	for i := range allStats {
		if allStats[i].DeviceName == deviceName {
			currentStats = &allStats[i]
			break
		}
	}

	if currentStats == nil {
		return Delta{}, fmt.Errorf("could not find stats for device %s", deviceName)
	}

	// 4. Get previous sample and handle the first sample case.
	prevSample, ok := m.prevSample[deviceName]

	// Create current sample
	currentSample := Sample{
		Timestamp: now,
		Stats:     *currentStats,
	}

	// Update state for the next call *before* checking if this is the first sample.
	m.prevSample[deviceName] = currentSample

	if !ok {
		return Delta{}, ErrFirstSample
	}

	// 5. Calculate and return the delta using the SampleDelta function.
	delta := SampleDelta(currentSample, prevSample)
	return delta, nil
}

// SampleMultiple captures disk I/O statistics for multiple mountpoints simultaneously.
// It returns a map of mountpoint to Delta, and any errors encountered combined into a single error.
// Individual errors can be extracted using errors.Unwrap() if needed.
func (m *VolumeMonitor) SampleMultiple(mountpoints []string) (map[string]Delta, error) {
	if len(mountpoints) == 0 {
		return make(map[string]Delta), nil
	}

	results := make(map[string]Delta, len(mountpoints))
	var errs []error

	for _, mountpoint := range mountpoints {
		delta, err := m.Sample(mountpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("mountpoint %s: %w", mountpoint, err))
			continue
		}
		results[mountpoint] = delta
	}

	return results, errors.Join(errs...)
}

// resolveDevice finds the block device name (e.g., "sda", "dm-0") for a given
// filesystem path by checking all system mountpoints. It uses a cache to
// speed up additional lookups.
func (m *VolumeMonitor) resolveDevice(path string) (string, error) {
	// Check cache first
	if dev, ok := m.mountCache[path]; ok {
		return dev, nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path: %w", err)
	}

	// Get all mount points
	mounts, err := m.procfs.GetMounts()
	if err != nil {
		return "", fmt.Errorf("could not get mounts: %w", err)
	}

	// Find the most specific mount point that contains our path
	var bestMount *procfs.MountInfo
	for _, mount := range mounts {
		if strings.HasPrefix(absPath, mount.MountPoint) {
			if bestMount == nil || len(mount.MountPoint) > len(bestMount.MountPoint) {
				bestMount = mount
			}
		}
	}

	if bestMount == nil {
		return "", ErrMountPointNotFound
	}

	// Extract device name (remove /dev/ prefix if present)
	deviceName := strings.TrimPrefix(bestMount.Source, "/dev/")

	// Cache the result
	m.mountCache[absPath] = deviceName

	return deviceName, nil
}
