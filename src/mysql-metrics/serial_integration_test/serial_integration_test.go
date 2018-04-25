package serial_integration_test

// Running the metrics main() function in parallel tests causes test pollution when reading the metrics with dropsonde

import (
	"fmt"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	lock           sync.RWMutex
	receivedEvents []eventTracker
	udpListener    net.PacketConn
)

var _ = Describe("mysql-metrics", func() {
	var (
		configFilepath  string
		tempDir         string
		password        string
		username        string
		err             error
		session         *gexec.Session
		metricFrequency int
	)

	BeforeEach(func() {
		if env, ok := os.LookupEnv("MYSQL_USER"); ok {
			username = env
		} else {
			panic("Missing environment variable: MYSQL_USER")
		}
		if env, ok := os.LookupEnv("MYSQL_PASSWORD"); ok {
			password = env
		} else {
			panic("Missing environment variable: MYSQL_PASSWORD")
		}

		metricFrequency = 1
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		configString := fmt.Sprintf(`{
			"host":"%s",
			"username":"%s",
			"password":"%s",
			"metrics_frequency":%d,
			"origin":"my_custom_origin",
			"emit_leader_follower_metrics": true,
			"emit_mysql_metrics": true,
			"emit_galera_metrics": true,
			"emit_disk_metrics": true
		}`, "localhost", username, password, metricFrequency)
		Expect(err).NotTo(HaveOccurred())
		configFilepath = filepath.Join(tempDir, "metric-config.yml")

		err = ioutil.WriteFile(configFilepath, []byte(configString), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

	})
	AfterEach(func() {
		err = os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session.Interrupt()).Should(gexec.Exit())
	})

	runMainWithArgs := func(args ...string) {
		args = append(
			args,
			"-c", configFilepath,
		)

		_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", args)
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(metricsBinPath, args...)
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	}

	// Shamelessly stolen from: https://github.com/cloudfoundry/dropsonde/blob/master/integration_test/dropsonde_end_to_end_test.go
	Describe("end-to-end metrics test", func() {
		origin := []string{"my_custom_origin"}

		BeforeEach(func() {
			var err error

			udpListener, err = net.ListenPacket("udp4", ":3457")
			Expect(err).ToNot(HaveOccurred())

			go listenForEvents(origin)
			runMainWithArgs()
		})

		AfterEach(func() {
			udpListener.Close()
		})

		It("emits metrics", func() {
			expectedEvent := eventTracker{
				eventType: "ValueMetric", name: "/my_custom_origin/system/persistent_disk_inodes_used_percent",
			}

			Eventually(func() []eventTracker {
				lock.Lock()
				defer lock.Unlock()
				return receivedEvents
			}, 5, .001).Should(ContainElement(expectedEvent))
		})
	})
})

func listenForEvents(origin []string) {
	for {
		buffer := make([]byte, 1024)
		n, _, err := udpListener.ReadFrom(buffer)
		if err != nil {
			return
		}

		if n == 0 {
			panic("Received empty packet")
		}
		envelope := new(events.Envelope)
		err = proto.Unmarshal(buffer[0:n], envelope)
		if err != nil {
			panic(err)
		}

		var eventId = envelope.GetEventType().String()

		tracker := eventTracker{eventType: eventId}

		if envelope.GetEventType() == events.Envelope_ValueMetric {
			tracker.name = envelope.GetValueMetric().GetName()
		} else {
			panic("Unexpected message type: " + envelope.GetEventType().String())
		}

		if envelope.GetOrigin() != strings.Join(origin, "/") {
			panic("origin not as expected")
		}

		func() {
			lock.Lock()
			defer lock.Unlock()
			receivedEvents = append(receivedEvents, tracker)
		}()
	}
}

type eventTracker struct {
	eventType string
	name      string
}
