package disk

import "syscall"

type Info struct {
	statfs func(string, *syscall.Statfs_t) error
}

func NewInfo(statfs func(string, *syscall.Statfs_t) error) Info {
	return Info{
		statfs: statfs,
	}
}

func (i Info) Stats(path string) (bytesFree, bytesTotal, inodesFree, inodesTotal uint64, err error) {
	statfs_t := &syscall.Statfs_t{}

	if err = i.statfs(path, statfs_t); err != nil {
		return
	}

	bytesFree = uint64(statfs_t.Bsize) * statfs_t.Bavail
	bytesTotal = uint64(statfs_t.Bsize) * statfs_t.Blocks
	inodesFree = statfs_t.Ffree
	inodesTotal = statfs_t.Files

	return
}
