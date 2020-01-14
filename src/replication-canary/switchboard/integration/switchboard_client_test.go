package integration

import (
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/replication-canary/switchboard"

	"net/http"
	"os"

	"crypto/tls"

	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Switchboard Client Integration Test", func() {
	var (
		client *switchboard.Client

		switchboardURL string

		switchboardUsername string
		switchboardPassword string

		httpClient *http.Client

		testLogger *lagertest.TestLogger
	)

	BeforeEach(func() {
		switchboardDomain, ok := os.LookupEnv("SWITCHBOARD_DOMAIN")
		if !ok {
			switchboardDomain = "0-proxy-p-mysql.bosh-lite.com"
		}

		switchboardUsername, ok = os.LookupEnv("SWITCHBOARD_USERNAME")
		if !ok {
			switchboardUsername = "username"
		}

		switchboardPassword, ok = os.LookupEnv("SWITCHBOARD_PASSWORD")
		if !ok {
			switchboardPassword = "password"
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		testLogger = lagertest.NewTestLogger("switchboard integration test")

		switchboardURL = "https://" + switchboardDomain
		skipSSLCertVerify := true
		client = switchboard.NewClient(
			switchboardURL,
			switchboardUsername,
			switchboardPassword,
			skipSSLCertVerify,
			testLogger,
		)

		client.EnableClusterTraffic()
	})

	AfterEach(func() {
		client.EnableClusterTraffic()
	})

	switchboardTrafficEnabled := func() bool {
		clusterStatusURL := switchboardURL + "/v0/cluster"
		req, err := http.NewRequest("GET", clusterStatusURL, nil)
		Expect(err).NotTo(HaveOccurred())

		req.SetBasicAuth(switchboardUsername, switchboardPassword)

		res, err := httpClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()

		Expect(res.StatusCode).To(BeNumerically("<", http.StatusBadRequest))

		var clusterResponse SwitchboardClusterResponse

		json.NewDecoder(res.Body).Decode(&clusterResponse)
		return (clusterResponse.TrafficEnabled)
	}

	It("correctly turns traffic off/on", func() {
		client.EnableClusterTraffic()

		Expect(switchboardTrafficEnabled()).To(BeTrue())

		client.DisableClusterTraffic()

		Expect(switchboardTrafficEnabled()).To(BeFalse())

		client.EnableClusterTraffic()

		Expect(switchboardTrafficEnabled()).To(BeTrue())
	})
})

type SwitchboardClusterResponse struct {
	TrafficEnabled bool `json:"trafficEnabled"`
}
