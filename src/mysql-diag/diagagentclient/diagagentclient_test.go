package diagagentclient_test

import (
	"net/http"

	. "github.com/cloudfoundry/mysql-diag/diagagentclient"
	"github.com/cloudfoundry/mysql-diag/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Diagagentclient", func() {
	var (
		server          *ghttp.Server
		host            string
		port            uint
		username        string
		password        string
		diagAgentClient *DiagAgentClient

		responseStatus int
		response       interface{}
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		username = "foo"
		password = "bar"

		host, port = testutil.ParseURL(server.URL())
		diagAgentClient = NewDiagAgentClient(host, port, username, password)

		responseStatus = http.StatusOK
		response = InfoResponse{
			Persistent: DiskInfo{
				BytesTotal:  456,
				BytesFree:   123,
				InodesTotal: 789,
				InodesFree:  567,
			},
			Ephemeral: DiskInfo{
				BytesTotal:  1456,
				BytesFree:   1123,
				InodesTotal: 1789,
				InodesFree:  1567,
			},
		}
	})

	JustBeforeEach(func() {
		server.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/v1/info"),
			ghttp.VerifyBasicAuth(username, password),
			ghttp.RespondWithJSONEncoded(responseStatus, response),
		))
	})

	It("returns info without error", func() {
		info, err := diagAgentClient.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(*info).To(Equal(response))
	})

	Context("when making the request returns an error", func() {
		client := NewDiagAgentClient("notvalid", port, "foo", "bar")

		It("returns an error", func() {
			info, err := client.Info()
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeNil())
		})
	})

	Context("when the response status is not 200", func() {
		BeforeEach(func() {
			responseStatus = http.StatusTeapot
		})

		It("returns an error", func() {
			_, err := diagAgentClient.Info()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the response fails to deserialize", func() {
		BeforeEach(func() {
			response = "not what we expected"
		})

		It("returns an error", func() {
			_, err := diagAgentClient.Info()
			Expect(err).To(HaveOccurred())
		})
	})
})
