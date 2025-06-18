package diskstat

import (
	"github.com/prometheus/procfs"
)

// MountReader provides a wrapper around procfs.GetMounts for dependency injection
type MountReader struct{}

// NewMountReader creates a new MountReader instance
func NewMountReader() MountInfoReader {
	return &MountReader{}
}

// GetMounts returns system mount information
func (m *MountReader) GetMounts() ([]*procfs.MountInfo, error) {
	return procfs.GetMounts()
}
