//go:build linux
// +build linux

package cpu_test

import (
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/mysql-metrics/cpu"
)

type cpuBurner struct {
	close chan struct{}
}

func (c *cpuBurner) burnup() {
	for {
		select {
		case <-c.close:
			return
		default:
			math.Sqrt(float64(time.Now().UnixNano()))
		}
	}

}

var _ = Describe("GetPercentage integration test", func() {
	var (
		burner cpuBurner
		stater cpu.Stater
	)

	BeforeEach(func() {
		burner = cpuBurner{
			close: make(chan struct{}),
		}

		f, err := os.Open("/proc/stat")
		Expect(err).NotTo(HaveOccurred())

		stater = cpu.New(f)
	})

	AfterEach(func() {
		close(burner.close)
	})

	It("determines the cpu utilization on linux", func() {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		cpus := r.Intn(runtime.NumCPU()) + 1

		requiredPercent := (float64(cpus) / float64(runtime.NumCPU())) * 100

		for i := 0; i < cpus; i++ {
			go burner.burnup()
		}

		Eventually(func() int {
			percent, err := stater.GetPercentage()
			if err != cpu.ErrNoPreviousData && err != nil {
				Fail(err.Error())
			}
			return percent
		}).Should(BeNumerically(">=", requiredPercent))
	})
})
