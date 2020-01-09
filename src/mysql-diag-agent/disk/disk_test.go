package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-diag-agent/disk"
)

var _ = Describe("GetDiskInfo", func() {

	It("provides disk space and inode information", func() {
		info, err := disk.GetDiskInfo("/")

		Expect(err).NotTo(HaveOccurred())

		Expect(info.BytesFree).ToNot(BeZero())
		Expect(info.BytesTotal).ToNot(BeZero())
		Expect(info.BytesFree).Should(BeNumerically("<", info.BytesTotal))

		// Concourse containers always report 0 for free/used inodes (!!)
		//Expect(info.InodesFree).ToNot(BeZero())
		//Expect(info.InodesTotal).ToNot(BeZero())
		//Expect(info.InodesFree).Should(BeNumerically("<", info.InodesTotal))
	})

	Context("when given a bad path", func() {
		It("returns an error", func() {
			_, err := disk.GetDiskInfo("/doesnotexist")

			Expect(err).To(HaveOccurred())
		})
	})
})
