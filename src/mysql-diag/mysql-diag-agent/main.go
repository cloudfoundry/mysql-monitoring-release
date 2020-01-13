package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

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

	pidfile, err := os.Create(c.PidFile)
	if err != nil {
		log.Fatal("Failed to create pidfile: ", err)
	} else {
		ioutil.WriteFile(pidfile.Name(), []byte(strconv.Itoa(os.Getpid())), 0644)
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
		w.Write(responseBytes)
	}

	basicAuth := middleware.NewBasicAuth(c.Username, c.Password)
	http.Handle("/api/v1/info", basicAuth.Wrap(http.HandlerFunc(infoHandler)))

	bindAddress := fmt.Sprintf(":%d", c.Port)
	fmt.Println(fmt.Sprintf("Listening on: '%s'", bindAddress))

	err = http.ListenAndServe(bindAddress, nil)

	log.Fatal(err)
}

type InfoResponse struct {
	Persistent *disk.DiskInfo `json:"persistent"`
	Ephemeral  *disk.DiskInfo `json:"ephemeral"`
}
