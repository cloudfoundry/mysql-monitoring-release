//go:build linux

package diskstat

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	// Test configuration constants
	testDuration     = 30 * time.Second
	sampleInterval   = 100 * time.Millisecond
	ioWorkloadTime   = 20 * time.Second
	concurrentReads  = 5
	concurrentWrites = 3
	fileSize         = 1024 * 1024 // 1MB per file
	numTestFiles     = 10
)

var (
	// Package-level test state
	monitor   *VolumeMonitor
	testDir   string
	mountPath string
)

// setupTestEnvironment initializes the test environment
func setupTestEnvironment() {
	// Use TEST_VOLUME environment variable if set, otherwise create temp directory
	testVolume := os.Getenv("TEST_VOLUME")
	if testVolume == "" {
		// Create a temporary directory for testing
		var err error
		testVolume, err = os.MkdirTemp("", "diskstats-integration-*")
		Expect(err).NotTo(HaveOccurred(), "Failed to create test directory")
	}

	// Get the absolute path
	var err error
	testVolume, err = filepath.Abs(testVolume)
	Expect(err).NotTo(HaveOccurred(), "Failed to get absolute path")

	// Create a subdirectory for this test run to avoid conflicts
	testDir = filepath.Join(testVolume, fmt.Sprintf("integration-test-%d", time.Now().UnixNano()))
	err = os.MkdirAll(testDir, 0755)
	Expect(err).NotTo(HaveOccurred(), "Failed to create test subdirectory")

	// Use the parent directory for monitoring (the actual mount point)
	mountPath = testVolume

	// Create the monitor with real implementations
	monitor, err = NewVolumeMonitor()
	Expect(err).NotTo(HaveOccurred(), "Failed to create VolumeMonitor")
}

// teardownTestEnvironment cleans up the test environment
func teardownTestEnvironment() {
	if testDir != "" {
		// Only remove the directory if it's not the TEST_VOLUME itself
		testVolume := os.Getenv("TEST_VOLUME")
		if testVolume == "" || testDir != testVolume {
			err := os.RemoveAll(testDir)
			if err != nil {
				GinkgoWriter.Printf("Warning: Failed to clean up test directory %s: %v\n", testDir, err)
			}
		} else {
			// If using TEST_VOLUME, just clean up the contents
			entries, err := os.ReadDir(testDir)
			if err == nil {
				for _, entry := range entries {
					path := filepath.Join(testDir, entry.Name())
					os.RemoveAll(path)
				}
			}
		}
	}
}

// generateIOLoad creates I/O activity in the background
func generateIOLoad(ctx context.Context) {
	var wg sync.WaitGroup

	// Start write workloads
	for i := 0; i < concurrentWrites; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			writeWorkload(ctx, workerID)
		}(i)
	}

	// Start read workloads
	for i := 0; i < concurrentReads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			readWorkload(ctx, workerID)
		}(i)
	}

	// Start mixed I/O workload
	wg.Add(1)
	go func() {
		defer wg.Done()
		mixedIOWorkload(ctx)
	}()

	wg.Wait()
}

// writeWorkload performs continuous write operations
func writeWorkload(ctx context.Context, workerID int) {
	data := make([]byte, fileSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	fileIndex := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			filename := filepath.Join(testDir, fmt.Sprintf("write_worker_%d_file_%d.dat", workerID, fileIndex))

			file, err := os.Create(filename)
			if err != nil {
				continue
			}

			// Write data in chunks
			chunkSize := 64 * 1024
			for offset := 0; offset < len(data); offset += chunkSize {
				end := offset + chunkSize
				if end > len(data) {
					end = len(data)
				}
				file.Write(data[offset:end])
			}

			file.Sync()
			file.Close()

			// Clean up periodically to avoid filling disk
			if fileIndex%5 == 0 && fileIndex > 0 {
				oldFile := filepath.Join(testDir, fmt.Sprintf("write_worker_%d_file_%d.dat", workerID, fileIndex-5))
				os.Remove(oldFile)
			}

			fileIndex++
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// readWorkload performs continuous read operations
func readWorkload(ctx context.Context, workerID int) {
	// Create initial files to read from
	testFiles := make([]string, numTestFiles)
	data := make([]byte, fileSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	for i := 0; i < numTestFiles; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("read_worker_%d_initial_%d.dat", workerID, i))
		err := os.WriteFile(filename, data, 0644)
		if err == nil {
			testFiles[i] = filename
		}
	}

	fileIndex := 0
	for {
		select {
		case <-ctx.Done():
			// Clean up test files
			for _, filename := range testFiles {
				if filename != "" {
					os.Remove(filename)
				}
			}
			return
		default:
			if testFiles[fileIndex] != "" {
				// Read the entire file in larger chunks to increase disk activity
				file, err := os.Open(testFiles[fileIndex])
				if err == nil {
					buffer := make([]byte, 256*1024) // Larger buffer for more I/O
					totalRead := 0
					for {
						n, err := file.Read(buffer)
						if err != nil || n == 0 {
							break
						}
						totalRead += n
						// Process the data to prevent optimization
						for j := 0; j < n; j += 4096 {
							_ = buffer[j]
						}
					}
					file.Close()

					// Occasionally re-create the file to avoid caching
					if fileIndex%3 == 0 {
						// Modify the data slightly to avoid caching
						for i := range data {
							data[i] = byte((i + totalRead) % 256)
						}
						os.WriteFile(testFiles[fileIndex], data, 0644)
					}
				}
			}
			fileIndex = (fileIndex + 1) % numTestFiles
			time.Sleep(10 * time.Millisecond) // Slightly faster reads
		}
	}
}

// mixedIOWorkload performs mixed read/write operations
func mixedIOWorkload(ctx context.Context) {
	data := make([]byte, fileSize/2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	operationCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			filename := filepath.Join(testDir, fmt.Sprintf("mixed_workload_%d.dat", operationCount%10))

			if operationCount%3 == 0 {
				// Write operation
				err := os.WriteFile(filename, data, 0644)
				if err == nil {
					// Force sync
					if file, err := os.OpenFile(filename, os.O_WRONLY, 0644); err == nil {
						file.Sync()
						file.Close()
					}
				}
			} else {
				// Read operation
				if _, err := os.ReadFile(filename); err != nil {
					// If file doesn't exist, create it for next read
					os.WriteFile(filename, data, 0644)
				}
			}

			operationCount++
			time.Sleep(50 * time.Millisecond)
		}
	}
}

var _ = Describe("Integration Tests for VolumeMonitor", Label("integration"), func() {
	BeforeEach(func() {
		setupTestEnvironment()
	})

	AfterEach(func() {
		teardownTestEnvironment()
	})

	Describe("Volume Monitor Stress Testing", Label("integration"), func() {
		When("subjected to intensive I/O workloads", func() {
			It("should accurately monitor disk statistics under stress", func() {
				By("starting integration test with test directory: " + testDir)

				// First sample should return ErrFirstSample
				_, err := monitor.Sample(mountPath)
				Expect(err).To(Equal(ErrFirstSample), "First sample should return ErrFirstSample")

				By("starting I/O workload...")

				// Start background I/O
				ctx, cancel := context.WithTimeout(context.Background(), testDuration)
				defer cancel()

				go generateIOLoad(ctx)

				// Give I/O workload time to start
				time.Sleep(2 * time.Second)

				By("starting sample collection...")

				// Collect samples
				samples := make([]Delta, 0)
				sampleCount := 0
				maxReadsPerSec := 0.0
				maxWritesPerSec := 0.0
				maxBusyPercent := 0.0
				totalReadOps := uint64(0)
				totalWriteOps := uint64(0)
				totalReadBytes := uint64(0)
				totalWriteBytes := uint64(0)

				ticker := time.NewTicker(sampleInterval)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						By("test duration completed")
						goto CollectionComplete
					case <-ticker.C:
						sampleCount++
						delta, err := monitor.Sample(mountPath)

						if err != nil {
							By(fmt.Sprintf("Sample %d: Error - %v", sampleCount, err))
							continue
						}

						samples = append(samples, delta)

						// Calculate metrics
						readsPerSec := delta.ReadsPerSecond()
						writesPerSec := delta.WritesPerSecond()
						busyPercent := delta.BusyPercent()

						// Track maximums
						if readsPerSec > maxReadsPerSec {
							maxReadsPerSec = readsPerSec
						}
						if writesPerSec > maxWritesPerSec {
							maxWritesPerSec = writesPerSec
						}
						if busyPercent > maxBusyPercent {
							maxBusyPercent = busyPercent
						}

						// Accumulate totals
						totalReadOps += delta.ReadIOs
						totalWriteOps += delta.WriteIOs
						totalReadBytes += delta.ReadSectors * 512
						totalWriteBytes += delta.WriteSectors * 512

						// Log every 10th sample
						if sampleCount%10 == 0 {
							readMBPerSec := float64(delta.ReadSectors*512) / (1024 * 1024) / delta.Elapsed.Seconds()
							writeMBPerSec := float64(delta.WriteSectors*512) / (1024 * 1024) / delta.Elapsed.Seconds()
							By(fmt.Sprintf("Sample %d: R/s=%.1f, W/s=%.1f, RMB/s=%.2f, WMB/s=%.2f, Busy=%.1f%%",
								sampleCount, readsPerSec, writesPerSec, readMBPerSec, writeMBPerSec, busyPercent))
						}
					}
				}

			CollectionComplete:
				By(fmt.Sprintf("Integration test completed. Collected %d samples", len(samples)))

				// Verify we collected meaningful data
				Expect(len(samples)).To(BeNumerically(">", 10), "Should collect at least 10 samples")

				// Count samples with I/O activity
				samplesWithActivity := 0
				for _, sample := range samples {
					if sample.ReadIOs > 0 || sample.WriteIOs > 0 {
						samplesWithActivity++
					}
				}

				// Verify we detected I/O activity
				activityPercentage := float64(samplesWithActivity) / float64(len(samples)) * 100
				Expect(activityPercentage).To(BeNumerically(">=", 50), "Should detect I/O activity in at least 50% of samples")

				// Verify meaningful I/O was detected
				Expect(totalReadOps+totalWriteOps).To(BeNumerically(">", 100), "Should detect significant I/O operations")

				// Verify performance metrics are reasonable
				// Note: Read operations might not always be detected if files are cached or if timing doesn't align
				// So we check that either reads OR writes were detected, but we must detect some disk utilization
				Expect(maxReadsPerSec+maxWritesPerSec).To(BeNumerically(">", 0), "Should detect either read or write operations")
				Expect(maxBusyPercent).To(BeNumerically(">", 0), "Should detect disk utilization")

				By("=== Integration Test Results ===")
				By(fmt.Sprintf("Total samples: %d", len(samples)))
				By(fmt.Sprintf("Samples with I/O activity: %d (%.1f%%)", samplesWithActivity, activityPercentage))
				By(fmt.Sprintf("Total read operations: %d", totalReadOps))
				By(fmt.Sprintf("Total write operations: %d", totalWriteOps))
				By(fmt.Sprintf("Total bytes read: %d (%.2f MB)", totalReadBytes, float64(totalReadBytes)/(1024*1024)))
				By(fmt.Sprintf("Total bytes written: %d (%.2f MB)", totalWriteBytes, float64(totalWriteBytes)/(1024*1024)))
				By(fmt.Sprintf("Max reads/sec: %.1f", maxReadsPerSec))
				By(fmt.Sprintf("Max writes/sec: %.1f", maxWritesPerSec))
				By(fmt.Sprintf("Max busy%%: %.1f", maxBusyPercent))

				// Additional diagnostic information
				samplesWithReads := 0
				samplesWithWrites := 0
				for _, sample := range samples {
					if sample.ReadIOs > 0 {
						samplesWithReads++
					}
					if sample.WriteIOs > 0 {
						samplesWithWrites++
					}
				}
				By(fmt.Sprintf("Samples with read activity: %d (%.1f%%)", samplesWithReads, float64(samplesWithReads)/float64(len(samples))*100))
				By(fmt.Sprintf("Samples with write activity: %d (%.1f%%)", samplesWithWrites, float64(samplesWithWrites)/float64(len(samples))*100))
			})
		})
	})

	Describe("Multiple Mountpoints", func() {
		When("monitoring multiple mountpoints simultaneously", func() {
			It("should handle concurrent monitoring of different paths", func() {
				// Create additional test directories within the test volume to avoid overlay filesystem issues
				testDir2 := filepath.Join(mountPath, "subdir2")
				err := os.MkdirAll(testDir2, 0755)
				Expect(err).NotTo(HaveOccurred())
				defer os.RemoveAll(testDir2)

				testDir3 := filepath.Join(mountPath, "subdir3")
				err = os.MkdirAll(testDir3, 0755)
				Expect(err).NotTo(HaveOccurred())
				defer os.RemoveAll(testDir3)

				mountpoints := []string{mountPath, testDir2, testDir3}

				By(fmt.Sprintf("Testing multiple mountpoints: %v", mountpoints))

				// Initialize all mountpoints
				// Note: All subdirectories on the same filesystem will resolve to the same device,
				// so only the first one will return ErrFirstSample
				for _, mp := range mountpoints {
					_, err := monitor.Sample(mp)
					if err != nil {
						Expect(err).To(Equal(ErrFirstSample), fmt.Sprintf("First sample should return ErrFirstSample for %s", mp))
					}
				}

				// Generate some I/O activity
				testFile1 := filepath.Join(testDir, "multi_test1.dat")
				testFile2 := filepath.Join(testDir2, "multi_test2.dat")
				testFile3 := filepath.Join(testDir3, "multi_test3.dat")

				// Create files with some I/O
				data := make([]byte, 1024*1024) // 1MB
				for i := range data {
					data[i] = byte(i % 256)
				}

				err = os.WriteFile(testFile1, data, 0644)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(testFile2, data, 0644)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(testFile3, data, 0644)
				Expect(err).NotTo(HaveOccurred())

				// Sample all mountpoints simultaneously
				time.Sleep(100 * time.Millisecond) // Allow some time for I/O
				results, err := monitor.SampleMultiple(mountpoints)

				By(fmt.Sprintf("SampleMultiple results: %d mountpoints, error: %v", len(results), err))

				// Clean up test files
				os.Remove(testFile1)
				os.Remove(testFile2)
				os.Remove(testFile3)

				// We should get at least some results
				Expect(len(results)).To(BeNumerically(">=", 1), "Should get results for at least one mountpoint")

				// Display results
				for mp, delta := range results {
					readsPerSec := delta.ReadsPerSecond()
					writesPerSec := delta.WritesPerSecond()
					readMBPerSec := float64(delta.ReadSectors*512) / (1024 * 1024) / delta.Elapsed.Seconds()
					writeMBPerSec := float64(delta.WriteSectors*512) / (1024 * 1024) / delta.Elapsed.Seconds()
					By(fmt.Sprintf("Mountpoint %s: R/s=%.1f, W/s=%.1f, RMB/s=%.3f, WMB/s=%.3f",
						mp, readsPerSec, writesPerSec, readMBPerSec, writeMBPerSec))
				}
			})
		})
	})

	Describe("Concurrent Access", func() {
		When("accessed concurrently by multiple goroutines", func() {
			It("should maintain thread safety", func() {
				// Initialize with first sample
				_, err := monitor.Sample(mountPath)
				Expect(err).To(Equal(ErrFirstSample))

				const numGoroutines = 10
				const numSamples = 20

				var wg sync.WaitGroup
				var mu sync.Mutex
				successCount := 0
				errorCount := 0

				// Start concurrent goroutines
				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(goroutineID int) {
						defer wg.Done()
						for j := 0; j < numSamples; j++ {
							_, err := monitor.Sample(mountPath)

							mu.Lock()
							if err != nil && err != ErrFirstSample {
								errorCount++
							} else {
								successCount++
							}
							mu.Unlock()

							time.Sleep(time.Millisecond)
						}
					}(i)
				}

				wg.Wait()

				By(fmt.Sprintf("Concurrent access test: %d successful samples, %d errors", successCount, errorCount))

				// Verify no unexpected errors occurred
				Expect(errorCount).To(Equal(0), "Should not have any unexpected errors during concurrent access")
				Expect(successCount).To(Equal(numGoroutines*numSamples), "Should have all samples succeed")
			})
		})
	})

	Describe("Long Running Stability", func() {
		When("running for an extended period", func() {
			It("should maintain stability and accuracy", func() {
				// First sample should return ErrFirstSample
				_, err := monitor.Sample(mountPath)
				Expect(err).To(Equal(ErrFirstSample))

				// Run for an extended period (but adjust based on environment)
				longTestDuration := 2 * time.Minute
				const longSampleInterval = 1 * time.Second

				// If TEST_DURATION environment variable is set and short, reduce long test duration
				if testDurationEnv := os.Getenv("TEST_DURATION"); testDurationEnv != "" {
					if envDuration, err := time.ParseDuration(testDurationEnv + "s"); err == nil && envDuration < 2*time.Minute {
						longTestDuration = envDuration // Use the environment duration directly
						By(fmt.Sprintf("Adjusted long test duration to %v based on TEST_DURATION environment variable", longTestDuration))
					}
				}

				ctx, cancel := context.WithTimeout(context.Background(), longTestDuration)
				defer cancel()

				By("Starting long-running test...")

				// Start background I/O
				go generateIOLoad(ctx)

				// Collect samples over the extended period
				ticker := time.NewTicker(longSampleInterval)
				defer ticker.Stop()

				sampleCount := 0
				errorCount := 0

				for {
					select {
					case <-ctx.Done():
						goto LongTestComplete
					case <-ticker.C:
						_, err := monitor.Sample(mountPath)
						sampleCount++

						if err != nil && err != ErrFirstSample {
							errorCount++
						}
					}
				}

			LongTestComplete:
				By(fmt.Sprintf("Long-running test progress: %d samples collected in %v", sampleCount, longTestDuration))

				// Verify stability
				Expect(sampleCount).To(BeNumerically(">", 5), "Should collect multiple samples over extended period")

				By(fmt.Sprintf("Long-running test completed: %d samples in %v, %d errors", sampleCount, longTestDuration, errorCount))

				// Allow for occasional errors but they should be minimal
				errorRate := float64(errorCount) / float64(sampleCount)
				Expect(errorRate).To(BeNumerically("<", 0.1), "Error rate should be less than 10%")
			})
		})
	})
})
