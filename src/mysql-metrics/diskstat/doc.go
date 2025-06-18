/*
Package diskstats provides disk I/O monitoring functionality for Linux systems.

It reads from /proc/diskstats and /proc/mounts to provide real-time disk statistics
for filesystem mountpoints, with output compatible with Percona Toolkit's pt-diskstats.

# Basic Usage

Create a VolumeMonitor and sample disk statistics:

	monitor, err := diskstats.NewVolumeMonitor()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize baseline (first sample returns ErrFirstSample)
	monitor.Sample("/")

	// Wait for some I/O activity
	time.Sleep(2 * time.Second)

	// Get disk statistics delta
	delta, err := monitor.Sample("/")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Reads/sec: %.2f\n", delta.ReadsPerSecond())
	fmt.Printf("Writes/sec: %.2f\n", delta.WritesPerSecond())

# Multiple Mountpoints

Monitor multiple mountpoints simultaneously:

	mountpoints := []string{"/", "/var", "/home"}
	results, errors := monitor.SampleMultiple(mountpoints)

	for mountpoint, delta := range results {
		fmt.Printf("%s: %.2f reads/sec\n", mountpoint, delta.ReadsPerSecond())
	}

# Working with Results

Process the returned Delta values to extract specific metrics:

	for mountpoint, delta := range results {
		fmt.Printf("%s: %.2f reads/sec, %.2f writes/sec\n",
			mountpoint, delta.ReadsPerSecond(), delta.WritesPerSecond())
		fmt.Printf("  Read throughput: %.2f MiB/s\n", delta.ReadMiBPerSec())
		fmt.Printf("  Write throughput: %.2f MiB/s\n", delta.WriteMiBPerSec())
		fmt.Printf("  Device utilization: %.1f%%\n", delta.BusyPercent())
	}

# Error Handling

The library defines several specific errors:

- ErrFirstSample: Returned on the first call to Sample() for a mountpoint
- ErrMountPointNotFound: The specified path has no corresponding mountpoint

# Requirements

This library requires:
- Linux operating system (uses /proc/diskstats and /proc/mounts)
- Appropriate permissions to read /proc files
- Go 1.24+
*/
package diskstat
