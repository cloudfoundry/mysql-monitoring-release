package diskstat

import (
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/blockdevice"
)

var _ = Describe("VolumeMonitor", func() {
	var (
		monitor    *VolumeMonitor
		mockStats  []blockdevice.Diskstats
		mockMounts []*procfs.MountInfo
	)

	BeforeEach(func() {
		// Setup mock disk statistics
		mockStats = []blockdevice.Diskstats{
			{
				Info: blockdevice.Info{DeviceName: "sda1"},
				IOStats: blockdevice.IOStats{
					ReadIOs:         1000,
					ReadMerges:      50,
					ReadSectors:     8000,
					ReadTicks:       2000,
					WriteIOs:        500,
					WriteMerges:     25,
					WriteSectors:    4000,
					WriteTicks:      1000,
					IOsInProgress:   2,
					IOsTotalTicks:   3000,
					WeightedIOTicks: 3500,
				},
			},
			{
				Info: blockdevice.Info{DeviceName: "vda4"},
				IOStats: blockdevice.IOStats{
					ReadIOs:         2000,
					ReadMerges:      100,
					ReadSectors:     16000,
					ReadTicks:       4000,
					WriteIOs:        1000,
					WriteMerges:     50,
					WriteSectors:    8000,
					WriteTicks:      2000,
					IOsInProgress:   1,
					IOsTotalTicks:   6000,
					WeightedIOTicks: 7000,
				},
			},
			{
				Info: blockdevice.Info{DeviceName: "sdc4"},
				IOStats: blockdevice.IOStats{
					ReadIOs:         500,
					ReadMerges:      25,
					ReadSectors:     4000,
					ReadTicks:       1000,
					WriteIOs:        250,
					WriteMerges:     10,
					WriteSectors:    2000,
					WriteTicks:      500,
					IOsInProgress:   0,
					IOsTotalTicks:   1500,
					WeightedIOTicks: 1750,
				},
			},
		}

		// Setup some mock mount information
		mockMounts = []*procfs.MountInfo{
			{
				Source:     "/dev/sda1",
				MountPoint: "/home",
			},
			{
				Source:     "/dev/vda4",
				MountPoint: "/data",
			},
			{
				Source:     "/dev/sdc4",
				MountPoint: "/",
			},
		}

		monitor = createMockVolumeMonitor(mockStats, mockMounts)
	})

	Describe("Sample", func() {
		When("sampling a valid mountpoint for the first time", func() {
			It("should return ErrFirstSample", func() {
				_, err := monitor.Sample("/data")
				Expect(err).To(Equal(ErrFirstSample))
			})

			It("should cache the device resolution", func() {
				_, _ = monitor.Sample("/data")
				Expect(monitor.mountCache["/data"]).To(Equal("vda4"))
			})

			It("should store the initial sample", func() {
				_, _ = monitor.Sample("/data")
				Expect(monitor.prevSample).To(HaveKey("vda4"))
			})
		})

		When("sampling a valid mountpoint for the second time", func() {
			BeforeEach(func() {
				// First sample to initialize
				_, _ = monitor.Sample("/data")

				// Update mock stats for second sample
				mockStats[1].ReadIOs = 2100
				mockStats[1].WriteIOs = 1050
				mockStats[1].ReadSectors = 16800
				mockStats[1].WriteSectors = 8400
				mockStats[1].ReadTicks = 4200
				mockStats[1].WriteTicks = 2100
				mockStats[1].IOsTotalTicks = 6300
				mockStats[1].WeightedIOTicks = 7350

				monitor.fs = &mockDiskstatsReader{diskstats: mockStats}
			})

			It("should return a valid Delta", func() {
				delta, err := monitor.Sample("/data")
				Expect(err).To(BeNil())
				Expect(delta).ToNot(BeNil())
			})

			It("should calculate correct deltas", func() {
				delta, err := monitor.Sample("/data")
				Expect(err).To(BeNil())

				Expect(delta.ReadIOs).To(Equal(uint64(100)))      // 2100 - 2000
				Expect(delta.WriteIOs).To(Equal(uint64(50)))      // 1050 - 1000
				Expect(delta.ReadSectors).To(Equal(uint64(800)))  // 16800 - 16000
				Expect(delta.WriteSectors).To(Equal(uint64(400))) // 8400 - 8000
			})

			It("should have a positive elapsed time", func() {
				time.Sleep(10 * time.Millisecond) // Ensure some time passes
				delta, err := monitor.Sample("/data")
				Expect(err).To(BeNil())
				Expect(delta.Elapsed).To(BeNumerically(">", 0))
			})

			It("should calculate meaningful metrics", func() {
				time.Sleep(100 * time.Millisecond)
				delta, err := monitor.Sample("/data")
				Expect(err).To(BeNil())

				readsPerSec := delta.ReadsPerSecond()
				writesPerSec := delta.WritesPerSecond()

				Expect(readsPerSec).To(BeNumerically(">", 0))
				Expect(writesPerSec).To(BeNumerically(">", 0))
			})
		})

		When("the mountpoint doesn't exist", func() {
			BeforeEach(func() {
				// Create a monitor with no root mount to truly test nonexistent paths
				mockMountsNoRoot := []*procfs.MountInfo{
					{
						Source:     "/dev/sda1",
						MountPoint: "/home",
					},
					{
						Source:     "/dev/vda4",
						MountPoint: "/data",
					},
				}
				monitor = createMockVolumeMonitor(mockStats, mockMountsNoRoot)
			})

			It("should return ErrMountPointNotFound", func() {
				_, err := monitor.Sample("/nonexistent")
				Expect(err).To(Equal(ErrMountPointNotFound))
			})
		})

		When("the device is not found in diskstats", func() {
			BeforeEach(func() {
				// Add a mount that points to a device not in diskstats
				missingMount := &procfs.MountInfo{
					Source:     "/dev/sdb1",
					MountPoint: "/missing",
				}
				mockMounts = append(mockMounts, missingMount)
				monitor = createMockVolumeMonitor(mockStats, mockMounts)
			})

			It("should return an error", func() {
				_, err := monitor.Sample("/missing")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find stats for device sdb1"))
			})
		})

		When("diskstats cannot be read", func() {
			BeforeEach(func() {
				monitor.fs = &mockDiskstatsReader{err: fmt.Errorf("permission denied")}
			})

			It("should return an error", func() {
				_, err := monitor.Sample("/data")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not read diskstats"))
			})
		})

		When("mounts cannot be read", func() {
			BeforeEach(func() {
				monitor.procfs = &mockMountInfoReader{err: fmt.Errorf("permission denied")}
			})

			It("should return an error", func() {
				_, err := monitor.Sample("/data")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not get mounts"))
			})
		})
	})

	Describe("resolveDevice", func() {
		Context("with absolute paths", func() {
			It("should resolve exact mount matches", func() {
				device, err := monitor.resolveDevice("/data")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("vda4"))
			})

			It("should resolve subdirectory paths", func() {
				device, err := monitor.resolveDevice("/data/subdir/file.txt")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("vda4"))
			})

			It("should find the most specific mount", func() {
				// Add a more specific mount
				specificMount := &procfs.MountInfo{
					Source:     "/dev/sda2",
					MountPoint: "/data/specific",
				}
				mockMounts = append(mockMounts, specificMount)
				monitor = createMockVolumeMonitor(mockStats, mockMounts)

				device, err := monitor.resolveDevice("/data/specific/file.txt")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("sda2"))
			})
		})

		Context("with relative paths", func() {
			It("should convert relative paths to absolute using current working directory", func() {
				// Create a mock mount that matches the current working directory or a parent
				// We'll use the root mount for simplicity since any path will match it
				mockMountsForRelative := []*procfs.MountInfo{
					{
						Source:     "vda1",
						MountPoint: "/",
					},
				}
				monitor = createMockVolumeMonitor(mockStats, mockMountsForRelative)

				// Test with a relative path
				device, err := monitor.resolveDevice("relative/path/file.txt")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("vda1"))

				// Verify the path was cached with the absolute version
				expectedAbsolutePath, _ := filepath.Abs("relative/path/file.txt")
				expectedCleanPath := filepath.Clean(expectedAbsolutePath)
				Expect(monitor.mountCache).To(HaveKey(expectedCleanPath))
			})

			It("should handle relative paths when current directory has a parent mount", func() {
				// Find a parent directory that we can use as a mount point
				// We'll use the root directory as a fallback since it always exists
				parentDir := "/"

				// Create a mock mount that matches a parent directory
				mockMountsForParent := []*procfs.MountInfo{
					{
						Source:     "sdb1",
						MountPoint: parentDir,
					},
				}
				monitor = createMockVolumeMonitor(mockStats, mockMountsForParent)

				// Test with a relative path - should resolve to the parent mount
				device, err := monitor.resolveDevice("subdir/file.txt")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("sdb1"))

				// Verify the cache contains the absolute path
				expectedAbsolutePath, _ := filepath.Abs("subdir/file.txt")
				Expect(monitor.mountCache).To(HaveKey(expectedAbsolutePath))
			})
		})

		Context("with caching", func() {
			It("should cache device resolutions", func() {
				// First call
				device1, err1 := monitor.resolveDevice("/data")
				Expect(err1).To(BeNil())
				Expect(device1).To(Equal("vda4"))

				// Verify cache is populated
				Expect(monitor.mountCache["/data"]).To(Equal("vda4"))

				// The second call should use cache (even if we break the mock)
				monitor.procfs = &mockMountInfoReader{err: fmt.Errorf("should not be called")}
				device2, err2 := monitor.resolveDevice("/data")
				Expect(err2).To(BeNil())
				Expect(device2).To(Equal("vda4"))
			})
		})

		Context("with device name cleaning", func() {
			It("should strip /dev/ prefix from device names", func() {
				device, err := monitor.resolveDevice("/home")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("sda1")) // Not "/dev/sda1"
			})
		})
	})

	Describe("edge cases", func() {
		Context("with empty mount points", func() {
			BeforeEach(func() {
				mockMounts = []*procfs.MountInfo{}
				monitor = createMockVolumeMonitor(mockStats, mockMounts)
			})

			It("should return ErrMountPointNotFound", func() {
				_, err := monitor.resolveDevice("/any/path")
				Expect(err).To(Equal(ErrMountPointNotFound))
			})
		})

		Context("with root mount point", func() {
			It("should handle root filesystem correctly", func() {
				device, err := monitor.resolveDevice("/some/deep/path")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("sdc4")) // Should match the root overlay mount
			})
		})

		Context("with path cleaning", func() {
			It("should handle paths with double slashes", func() {
				device, err := monitor.resolveDevice("/data//subdir///file.txt")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("vda4"))
			})

			It("should handle paths with . and ..", func() {
				device, err := monitor.resolveDevice("/data/./subdir/../file.txt")
				Expect(err).To(BeNil())
				Expect(device).To(Equal("vda4"))
			})
		})
	})

	Describe("SampleMultiple", func() {
		Context("with multiple valid mountpoints", func() {
			BeforeEach(func() {
				// Initialize with first samples to avoid ErrFirstSample
				_, _ = monitor.Sample("/home")
				_, _ = monitor.Sample("/data")
				_, _ = monitor.Sample("/")

				// Update mock stats for second sample
				mockStats[0].ReadIOs = 1100 // sda1
				mockStats[0].WriteIOs = 550
				mockStats[1].ReadIOs = 2200 // vda4
				mockStats[1].WriteIOs = 1100
				mockStats[2].ReadIOs = 600 // overlay
				mockStats[2].WriteIOs = 300

				monitor.fs = &mockDiskstatsReader{diskstats: mockStats}
			})

			It("should return deltas for all mountpoints", func() {
				mountpoints := []string{"/home", "/data", "/"}
				results, err := monitor.SampleMultiple(mountpoints)

				Expect(err).To(BeNil())
				Expect(results).To(HaveLen(3))
				Expect(results).To(HaveKey("/home"))
				Expect(results).To(HaveKey("/data"))
				Expect(results).To(HaveKey("/"))

				// Verify deltas are calculated correctly
				Expect(results["/home"].ReadIOs).To(Equal(uint64(100)))  // 1100 - 1000
				Expect(results["/home"].WriteIOs).To(Equal(uint64(50)))  // 550 - 500
				Expect(results["/data"].ReadIOs).To(Equal(uint64(200)))  // 2200 - 2000
				Expect(results["/data"].WriteIOs).To(Equal(uint64(100))) // 1100 - 1000
				Expect(results["/"].ReadIOs).To(Equal(uint64(100)))      // 600 - 500
				Expect(results["/"].WriteIOs).To(Equal(uint64(50)))      // 300 - 250
			})

			It("should handle empty mountpoint list", func() {
				results, err := monitor.SampleMultiple([]string{})
				Expect(err).To(BeNil())
				Expect(results).To(HaveLen(0))
			})

			It("should handle single mountpoint", func() {
				results, err := monitor.SampleMultiple([]string{"/home"})
				Expect(err).To(BeNil())
				Expect(results).To(HaveLen(1))
				Expect(results).To(HaveKey("/home"))
			})
		})

		Context("with first-time sampling", func() {
			It("should return ErrFirstSample for all mountpoints", func() {
				mountpoints := []string{"/home", "/data", "/"}
				results, err := monitor.SampleMultiple(mountpoints)
				Expect(err).ToNot(BeNil())
				Expect(results).To(HaveLen(0))

				// Verify all errors are ErrFirstSample
				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(HaveLen(3))
				for _, individualErr := range individualErrors {
					Expect(individualErr.Error()).To(ContainSubstring("this is the first sample"))
				}
			})
		})

		Context("with mixed success and failure", func() {
			BeforeEach(func() {
				// Initialize only some mountpoints
				_, _ = monitor.Sample("/home")
				_, _ = monitor.Sample("/data")
				// Don't initialize "/" - it will return ErrFirstSample

				// Update mock stats for initialized mountpoints
				mockStats[0].ReadIOs = 1100 // sda1
				mockStats[0].WriteIOs = 550
				mockStats[1].ReadIOs = 2200 // vda4
				mockStats[1].WriteIOs = 1100

				monitor.fs = &mockDiskstatsReader{diskstats: mockStats}
			})

			It("should return partial results and combined errors", func() {
				// Use a truly invalid mountpoint by creating a monitor without the root mount
				mockMountsNoRoot := []*procfs.MountInfo{
					{
						Source:     "/dev/sda1",
						MountPoint: "/home",
					},
					{
						Source:     "/dev/vda4",
						MountPoint: "/data",
					},
				}
				monitor = createMockVolumeMonitor(mockStats, mockMountsNoRoot)

				// Initialize the known mountpoints
				_, _ = monitor.Sample("/home")
				_, _ = monitor.Sample("/data")

				// Update stats
				mockStats[0].ReadIOs = 1100
				mockStats[0].WriteIOs = 550
				mockStats[1].ReadIOs = 2200
				mockStats[1].WriteIOs = 1100
				monitor.fs = &mockDiskstatsReader{diskstats: mockStats}

				mountpoints := []string{"/home", "/data", "/invalid", "/nonexistent"}
				results, err := monitor.SampleMultiple(mountpoints)

				// Should have results for initialized mountpoints
				Expect(results).To(HaveLen(2))
				Expect(results).To(HaveKey("/home"))
				Expect(results).To(HaveKey("/data"))
				Expect(results).ToNot(HaveKey("/invalid"))
				Expect(results).ToNot(HaveKey("/nonexistent"))

				// Should have errors for uninitialized/invalid mountpoints
				Expect(err).ToNot(BeNil())
				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(HaveLen(2))

				// Check error messages
				errorMessages := make([]string, len(individualErrors))
				for i, e := range individualErrors {
					errorMessages[i] = e.Error()
				}
				Expect(errorMessages).To(ContainElement(ContainSubstring("mountpoint /invalid:")))
				Expect(errorMessages).To(ContainElement(ContainSubstring("mountpoint /nonexistent:")))
			})
		})

		Context("with invalid mountpoints", func() {
			It("should return errors for all invalid mountpoints", func() {
				// Create a monitor with no root mount to ensure paths don't resolve
				mockMountsNoRoot := []*procfs.MountInfo{
					{
						Source:     "/dev/sda1",
						MountPoint: "/home",
					},
				}
				monitor = createMockVolumeMonitor(mockStats, mockMountsNoRoot)

				mountpoints := []string{"/invalid1", "/invalid2", "/invalid3"}
				results, err := monitor.SampleMultiple(mountpoints)

				Expect(results).To(HaveLen(0))
				Expect(err).ToNot(BeNil())

				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(HaveLen(3))

				// Verify each error is properly formatted
				for i, individualErr := range individualErrors {
					expectedMountpoint := mountpoints[i]
					Expect(individualErr.Error()).To(ContainSubstring(fmt.Sprintf("mountpoint %s:", expectedMountpoint)))
				}
			})
		})

		Context("with diskstats read error", func() {
			BeforeEach(func() {
				// Initialize first
				_, _ = monitor.Sample("/home")

				// Set up error for subsequent reads
				monitor.fs = &mockDiskstatsReader{err: fmt.Errorf("diskstats read failed")}
			})

			It("should return errors for all mountpoints", func() {
				mountpoints := []string{"/home", "/data"}
				results, err := monitor.SampleMultiple(mountpoints)

				Expect(results).To(HaveLen(0))
				Expect(err).ToNot(BeNil())

				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(HaveLen(2))

				// All errors should contain the diskstats error
				for _, individualErr := range individualErrors {
					Expect(individualErr.Error()).To(ContainSubstring("diskstats read failed"))
				}
			})
		})

		Context("with mount info read error", func() {
			BeforeEach(func() {
				// Set up error for mount reads
				monitor.procfs = &mockMountInfoReader{err: fmt.Errorf("mount info read failed")}
			})

			It("should return errors for all mountpoints", func() {
				mountpoints := []string{"/home", "/data"}
				results, err := monitor.SampleMultiple(mountpoints)

				Expect(results).To(HaveLen(0))
				Expect(err).ToNot(BeNil())

				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(HaveLen(2))

				// All errors should contain the mount info error
				for _, individualErr := range individualErrors {
					Expect(individualErr.Error()).To(ContainSubstring("mount info read failed"))
				}
			})
		})

		Context("error extraction", func() {
			It("should allow extraction of individual errors", func() {
				// Create a monitor with no root mount to ensure paths don't resolve
				mockMountsNoRoot := []*procfs.MountInfo{
					{
						Source:     "/dev/sda1",
						MountPoint: "/home",
					},
				}
				monitor = createMockVolumeMonitor(mockStats, mockMountsNoRoot)

				mountpoints := []string{"/invalid1", "/invalid2"}
				_, err := monitor.SampleMultiple(mountpoints)

				Expect(err).ToNot(BeNil())

				// Extract individual errors
				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(HaveLen(2))

				// Verify each error can be inspected individually
				for i, individualErr := range individualErrors {
					expectedMountpoint := mountpoints[i]
					Expect(individualErr.Error()).To(ContainSubstring(fmt.Sprintf("mountpoint %s:", expectedMountpoint)))
				}
			})

			It("should return nil when no errors occur", func() {
				// Initialize mountpoints first
				_, _ = monitor.Sample("/home")
				_, _ = monitor.Sample("/data")

				// Update stats
				mockStats[0].ReadIOs = 1100
				mockStats[1].ReadIOs = 2200
				monitor.fs = &mockDiskstatsReader{diskstats: mockStats}

				mountpoints := []string{"/home", "/data"}
				results, err := monitor.SampleMultiple(mountpoints)

				Expect(err).To(BeNil())
				Expect(results).To(HaveLen(2))

				individualErrors := unwrapErrors(err)
				Expect(individualErrors).To(BeNil())
			})
		})
	})
})

// Mock implementations for testing
type mockDiskstatsReader struct {
	diskstats []blockdevice.Diskstats
	err       error
}

func (m *mockDiskstatsReader) ProcDiskstats() ([]blockdevice.Diskstats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.diskstats, nil
}

type mockMountInfoReader struct {
	mounts []*procfs.MountInfo
	err    error
}

func (m *mockMountInfoReader) GetMounts() ([]*procfs.MountInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.mounts, nil
}

// Test helper to create a VolumeMonitor with mocked dependencies
func createMockVolumeMonitor(diskstats []blockdevice.Diskstats, mounts []*procfs.MountInfo) *VolumeMonitor {
	return NewVolumeMonitorWithDeps(
		&mockDiskstatsReader{diskstats: diskstats},
		&mockMountInfoReader{mounts: mounts},
	)
}

func unwrapErrors(err error) []error {
	if unwrap, ok := err.(interface{ Unwrap() []error }); ok {
		return unwrap.Unwrap()
	}

	return nil
}
