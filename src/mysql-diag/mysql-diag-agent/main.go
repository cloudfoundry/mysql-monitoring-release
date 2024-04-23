package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudfoundry/mysql-diag/mysql-diag-agent/config"
	"github.com/cloudfoundry/mysql-diag/mysql-diag-agent/disk"
	"github.com/cloudfoundry/mysql-diag/mysql-diag-agent/middleware"
)

const (
	defaultConfigPath = "/var/vcap/jobs/mysql-diag-agent/config/mysql-diag-agent.yml"
)

var (
	configFilepath = flag.String("c", defaultConfigPath, "location of config file")
)

func main() {
	flag.Parse()

	c, err := config.LoadFromFile(*configFilepath)
	if err != nil {
		log.Fatal(err)
	}

	infoHandler := func(w http.ResponseWriter, req *http.Request) {
		ephemeral, err := disk.GetDiskInfo(c.GetEphemeralDiskPath())
		if err != nil {
			panic(err)
		}

		persistent, err := disk.GetDiskInfo(c.GetPersistentDiskPath())
		if err != nil {
			panic(err)
		}

		response := InfoResponse{Ephemeral: ephemeral, Persistent: persistent}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBytes)
	}

	router := http.NewServeMux()
	basicAuth := middleware.NewBasicAuth(c.Username, c.Password)
	router.Handle("/api/v1/info", basicAuth.Wrap(http.HandlerFunc(infoHandler)))

	bindAddress := fmt.Sprintf(":%d", c.Port)
	fmt.Printf("Listening on: '%s'\n", bindAddress)

	listener, err := c.NetworkListener()
	if err != nil {
		log.Fatalf("Failed to create network listener: %v", err)
	}

	err = http.Serve(listener, router)
	log.Fatal(err)
}

type InfoResponse struct {
	Persistent *disk.DiskInfo `json:"persistent"`
	Ephemeral  *disk.DiskInfo `json:"ephemeral"`
}
