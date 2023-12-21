package cpu_test

import (
	"errors"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-metrics/cpu"
)

type fakeReadSeeker struct {
	ReadReturns struct {
		Error error
	}
	SeekReturns struct {
		Error error
	}
}

func (f *fakeReadSeeker) Read([]byte) (int, error)       { return 0, f.ReadReturns.Error }
func (f *fakeReadSeeker) Seek(int64, int) (int64, error) { return 0, f.SeekReturns.Error }

var _ = Describe("Stater", func() {
	Describe("GetPercentage", func() {
		var tmpFile *os.File

		BeforeEach(func() {
			var err error
			tmpFile, err = ioutil.TempFile("", "procStatFile")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(tmpFile.Name())
		})

		Context("when data is as we expect", func() {
			var previousProcStatsString, currentProcStatsString string

			BeforeEach(func() {
				previousProcStatsString = "cpu  1229059 1 838002 14163742 781998 0 34211 0\ncpu0 49663 0 40234 104757317 542691 4420 39572 0"
				currentProcStatsString = "cpu  1245566 1 858544 14349528 805918 0 35187 0\ncpu0 49663 0 40234 104757317 542691 4420 39572 0"

				tmpFile.WriteString(previousProcStatsString)
			})

			It("provides average cpu percentage for all cores", func() {
				stater := cpu.New(tmpFile)

				_, err := stater.GetPercentage()
				Expect(err).To(Equal(cpu.ErrNoPreviousData))

				tmpFile.Truncate(0)
				tmpFile.Seek(0, 0)
				tmpFile.WriteString(currentProcStatsString)

				percent, err := stater.GetPercentage()
				Expect(err).NotTo(HaveOccurred())
				Expect(percent).To(Equal(15))
			})
		})

		Context("when proc stats data is not as we expect", func() {
			BeforeEach(func() {
				procStatsString := "cpu1  1229059 1 838002 14163742 781998 0 34211 0\ncpu0 49663 0 40234 104757317 542691 4420 39572 0"

				tmpFile.WriteString(procStatsString)
			})

			It("returns an error", func() {
				stater := cpu.New(tmpFile)

				_, err := stater.GetPercentage()

				Expect(err).To(MatchError("/proc/stat file does not contain data as expected"))
			})

			DescribeTable("when the the data in /proc/stats are not numbers",
				func(procStatsString, expected string) {
					tmpFile.Seek(0, 0)
					tmpFile.WriteString(procStatsString)
					stater := cpu.New(tmpFile)

					_, err := stater.GetPercentage()

					Expect(err).To(MatchError(expected))
				},
				Entry("user", "cpu  % 1 838002 14163742 781998 0 34211 0\n", `error parsing user value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("nice", "cpu  1 % 838002 14163742 781998 0 34211 0\n", `error parsing nice value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("system", "cpu  1 1 % 14163742 781998 0 34211 0\n", `error parsing system value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("idle", "cpu  1 1 838002 % 781998 0 34211 0\n", `error parsing idle value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("iowait", "cpu  1 1 838002 14163742 % 0 34211 0\n", `error parsing iowait value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("irq", "cpu  1 1 838002 14163742 781998 % 34211 0\n", `error parsing irq value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("softirq", "cpu  1 1 838002 14163742 781998 0 % 0\n", `error parsing softirq value: strconv.ParseInt: parsing "%": invalid syntax`),
				Entry("steal", "cpu 1 1 838002 14163742 781998 0 34211 %\n", `error parsing steal value: strconv.ParseInt: parsing "%": invalid syntax`),
			)
		})

		Context("when proc stats cannot be seeked", func() {
			It("returns an error", func() {
				fakeReadSeeker := &fakeReadSeeker{}

				fakeReadSeeker.SeekReturns.Error = errors.New("failed to seek")
				stater := cpu.New(fakeReadSeeker)

				_, err := stater.GetPercentage()

				Expect(err).To(MatchError("failed to seek"))
			})
		})

		Context("when proc stats cannot be read", func() {
			It("returns an error", func() {
				fakeReadSeeker := &fakeReadSeeker{}

				fakeReadSeeker.ReadReturns.Error = errors.New("failed to read")
				stater := cpu.New(fakeReadSeeker)

				_, err := stater.GetPercentage()

				Expect(err).To(MatchError("failed to read /proc/stat file"))
			})
		})

	})
})
