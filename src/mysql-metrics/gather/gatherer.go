package gather

import "strconv"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DatabaseClient
type DatabaseClient interface {
	ShowGlobalStatus() (map[string]string, error)
	ShowGlobalVariables() (map[string]string, error)
	ShowSlaveStatus() (map[string]string, error)
	HeartbeatStatus() (map[string]string, error)
	ServicePlansDiskAllocated() (map[string]string, error)
	IsAvailable() bool
	IsFollower() (bool, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Stater
type Stater interface {
	Stats(path string) (bytesFree, bytesTotal, inodesFree, inodesTotal uint64, err error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CpuStater
type CpuStater interface {
	GetPercentage() (int, error)
}

type Gatherer struct {
	client          DatabaseClient
	stater          Stater
	cpuStater       CpuStater
	previousQueries int
}

func NewGatherer(client DatabaseClient, stater Stater, cpuStater CpuStater) *Gatherer {
	return &Gatherer{
		client:          client,
		stater:          stater,
		cpuStater:       cpuStater,
		previousQueries: -1,
	}
}

func (g Gatherer) BrokerStats() (map[string]string, error) {
	return g.client.ServicePlansDiskAllocated()
}
func (g Gatherer) CPUStats() (map[string]string, error) {
	percentage, err := g.cpuStater.GetPercentage()
	if err != nil {
		return nil, err
	}
	return map[string]string{"cpu_utilization_percent": strconv.Itoa(percentage)}, err
}
func (g Gatherer) DiskStats() (map[string]string, error) {
	bytesFreePersistent, bytesTotalPersistent, inodesFreePersistent, inodesTotalPersistent, err := g.stater.Stats("/var/vcap/store")
	if err != nil {
		return nil, err
	}

	bytesFreeEphemeral, bytesTotalEphemeral, inodesFreeEphemeral, inodesTotalEphemeral, err := g.stater.Stats("/var/vcap/data")
	if err != nil {
		return nil, err
	}

	persistentDiskUsedBytes := bytesTotalPersistent - bytesFreePersistent
	ephemeralDiskUsedBytes := bytesTotalEphemeral - bytesFreeEphemeral
	persistentInodesUsed := inodesTotalPersistent - inodesFreePersistent
	ephemeralInodesUsed := inodesTotalEphemeral - inodesFreeEphemeral
	return map[string]string{
		"persistent_disk_used":                strconv.FormatUint(persistentDiskUsedBytes/1024, 10),
		"persistent_disk_free":                strconv.FormatUint(bytesFreePersistent/1024, 10),
		"persistent_disk_used_percent":        strconv.FormatUint(g.calculateWholePercent(persistentDiskUsedBytes, bytesTotalPersistent), 10),
		"persistent_disk_inodes_used":         strconv.FormatUint(persistentInodesUsed, 10),
		"persistent_disk_inodes_free":         strconv.FormatUint(inodesFreePersistent, 10),
		"persistent_disk_inodes_used_percent": strconv.FormatUint(g.calculateWholePercent(persistentInodesUsed, inodesTotalPersistent), 10),
		"ephemeral_disk_used":                 strconv.FormatUint(ephemeralDiskUsedBytes/1024, 10),
		"ephemeral_disk_free":                 strconv.FormatUint(bytesFreeEphemeral/1024, 10),
		"ephemeral_disk_used_percent":         strconv.FormatUint(g.calculateWholePercent(ephemeralDiskUsedBytes, bytesTotalEphemeral), 10),
		"ephemeral_disk_inodes_used":          strconv.FormatUint(ephemeralInodesUsed, 10),
		"ephemeral_disk_inodes_free":          strconv.FormatUint(inodesFreeEphemeral, 10),
		"ephemeral_disk_inodes_used_percent":  strconv.FormatUint(g.calculateWholePercent(ephemeralInodesUsed, inodesTotalEphemeral), 10),
	}, nil
}

func (Gatherer) calculateWholePercent(numerator, denominator uint64) uint64 {
	numeratorFloat := float64(numerator)
	denominatorFloat := float64(denominator)
	return uint64((numeratorFloat / denominatorFloat) * 100)
}

func (g Gatherer) IsDatabaseAvailable() bool {
	return g.client.IsAvailable()
}

func (g Gatherer) IsDatabaseFollower() (bool, error) {
	return g.client.IsFollower()
}

func (g *Gatherer) DatabaseMetadata() (globalStatus map[string]string, globalVariables map[string]string, err error) {
	globalStatus, err = g.client.ShowGlobalStatus()
	if err != nil {
		return nil, nil, err
	}

	globalVariables, err = g.client.ShowGlobalVariables()
	if err != nil {
		return nil, nil, err
	}

	currentQueries := -1

	if currentQueriesString, ok := globalStatus["queries"]; ok {
		var err error
		if currentQueries, err = strconv.Atoi(currentQueriesString); err != nil {
			globalStatus["queries_delta"] = "0"
		}
	} else {
		globalStatus["queries_delta"] = "0"
	}

	if g.previousQueries != -1 {
		if currentQueries-g.previousQueries >= 0 {
			globalStatus["queries_delta"] = strconv.Itoa(currentQueries - g.previousQueries)
		}
	}

	g.previousQueries = currentQueries

	return
}

func (g Gatherer) FollowerMetadata() (slaveStatus map[string]string, heartbeatStatus map[string]string, err error) {
	slaveStatus, err = g.client.ShowSlaveStatus()
	if err != nil {
		return nil, nil, err
	}

	heartbeatStatus, err = g.client.HeartbeatStatus()
	if err != nil {
		return slaveStatus, nil, err
	}

	return
}
