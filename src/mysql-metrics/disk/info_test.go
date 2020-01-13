package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"github.com/cloudfoundry/mysql-metrics/disk"
	"syscall"
)

type statfs struct {
	callCount          int
	lastPathCalledWith string
	returnsStatfs_t    *syscall.Statfs_t
	err                error
}

func (s *statfs) Statfs(path string, t *syscall.Statfs_t) error {
	s.callCount++
	s.lastPathCalledWith = path

	if s.err != nil {
		return s.err
	}

	*t = *s.returnsStatfs_t

	return nil
}

var _ = Describe("Info", func() {
	It("provides disk space and inode information", func() {
		fakeStatfs := &statfs{
			returnsStatfs_t: &syscall.Statfs_t{
				Bsize:  1024,
				Blocks: 1024,
				Bavail: 512,
				Files:  300,
				Ffree:  200,
			},
		}

		info := disk.NewInfo(fakeStatfs.Statfs)

		bytesFree, bytesTotal, inodesFree, inodesTotal, err := info.Stats("/")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStatfs.callCount).To(Equal(1))

		Expect(bytesTotal).To(Equal(uint64(1048576)))
		Expect(bytesFree).To(Equal(uint64(524288)))
		Expect(inodesFree).To(Equal(uint64(200)))
		Expect(inodesTotal).To(Equal(uint64(300)))
	})

	Context("when the Statfs call fails", func() {
		It("returns an error", func() {
			fakeStatfs := &statfs{
				err: errors.New("failed to stat"),
			}

			info := disk.NewInfo(fakeStatfs.Statfs)

			_, _, _, _, err := info.Stats("/")
			Expect(err).To(MatchError("failed to stat"))
		})
	})
})
