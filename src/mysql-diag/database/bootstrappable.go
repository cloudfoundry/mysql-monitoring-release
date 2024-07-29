package database

func DoWeNeedBootstrap(gs []*GaleraStatus) bool {
	if len(gs) == 0 {
		return false
	}

	for _, st := range gs {
		if isHealthy(st) {
			return false
		}
	}

	return true
}

// returns true if the cluster needs bootstrap
func CheckClusterBootstrapStatus(rows map[string]*NodeClusterStatus) bool {
	statuses := make([]*GaleraStatus, 0, len(rows))
	for _, row := range rows {
		statuses = append(statuses, row.Status)
	}

	if DoWeNeedBootstrap(statuses) {
		return true
	} else {
		return false
	}
}

func isHealthy(status *GaleraStatus) bool {
	return status != nil && (status.LocalState == "Synced" || status.LocalState == "Donor/Desynced")
}
