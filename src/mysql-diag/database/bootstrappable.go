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

func isHealthy(status *GaleraStatus) bool {
	return status != nil && (status.LocalState == "Synced" || status.LocalState == "Donor/Desynced")
}
