package disk

import (
	"syscall"
)

type DiskInfo struct {
	BytesTotal  uint64 `json:"bytes_total"`
	BytesFree   uint64 `json:"bytes_free"`
	InodesTotal uint64 `json:"inodes_total"`
	InodesFree  uint64 `json:"inodes_free"`
}

func GetDiskInfo(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	return &DiskInfo{
		BytesTotal:  stat.Blocks * uint64(stat.Bsize),
		BytesFree:   stat.Bfree * uint64(stat.Bsize),
		InodesTotal: stat.Files,
		InodesFree:  stat.Ffree,
	}, nil
}
