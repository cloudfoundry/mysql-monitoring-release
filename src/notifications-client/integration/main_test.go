package main_test

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"notifications-client/notificationemailer/integration/mailinator"
)

func runMainWithArgs(args ...string) *gexec.Session {
	command := exec.Command(binPath, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

var _ = Describe("NotificationsClient", func() {
	Describe("Executable", func() {
		var (
			args         []string
			toAddressArg string
			subject      string
			bodyHTML     string
			kindID       string

			notificationsDomainArg    string
			uaaDomainArg              string
			uaaAdminClientUsernameArg string
			uaaAdminClientSecretArg   string
			uaaClientUsernameArg      string
			uaaClientSecretArg        string
			skipSSLCertVerifyArg      bool
		)

		BeforeEach(func() {
			uuid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())

			args = []string{}
			toAddressArg = recipientEmail
			subject = "fake-subject" + uuid.String()
			bodyHTML = "fake-body"
			kindID = "fake-kind-id"

			notificationsDomainArg = notificationsDomain
			uaaDomainArg = uaaDomain
			uaaAdminClientUsernameArg = uaaAdminClientUsername
			uaaAdminClientSecretArg = uaaAdminClientSecret
			uaaClientUsernameArg = uaaClientUsername
			uaaClientSecretArg = uaaClientSecret
			skipSSLCertVerifyArg = skipSSLCertVerify
		})

		JustBeforeEach(func() {
			args = append(args, fmt.Sprintf("--toAddress=%s", toAddressArg))
			args = append(args, fmt.Sprintf("--subject=%s", subject))
			args = append(args, fmt.Sprintf("--bodyHTML=%s", bodyHTML))
			args = append(args, fmt.Sprintf("--kindID=%s", kindID))
			args = append(args, fmt.Sprintf("--notificationsDomain=%s", notificationsDomainArg))
			args = append(args, fmt.Sprintf("--uaaDomain=%s", uaaDomainArg))
			args = append(args, fmt.Sprintf("--uaaClientUsername=%s", uaaClientUsernameArg))
			args = append(args, fmt.Sprintf("--uaaClientSecret=%s", uaaClientSecretArg))
			args = append(args, fmt.Sprintf("--uaaAdminClientUsername=%s", uaaAdminClientUsernameArg))
			args = append(args, fmt.Sprintf("--uaaAdminClientSecret=%s", uaaAdminClientSecretArg))
			args = append(args, fmt.Sprintf("--skipSSLCertVerify=%t", skipSSLCertVerifyArg))
		})

		It("sends an email", func() {
			mailinatorClient := mailinator.NewClient(mailinatorToken)

			session := runMainWithArgs(args...)
			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			recipientEmailFragment := strings.Split(recipientEmail, "@")[0]

			hasMailinatorReceivedEmail := func() bool {
				res, err := mailinatorClient.GetMessageList(recipientEmail)
				Expect(err).NotTo(HaveOccurred())

				found := false
				for _, message := range res.Messages {
					matchedSubject, err := regexp.MatchString(".*"+subject, message.Subject)
					Expect(err).NotTo(HaveOccurred())

					if matchedSubject && (message.To == recipientEmailFragment) {
						found = true
						break
					}
				}

				return found
			}

			Eventually(
				hasMailinatorReceivedEmail,
				time.Duration(testTimeout)*time.Second,
				time.Duration(testTimeout/20)*time.Second,
			).Should(BeTrue(), fmt.Sprintf("Did not find message with Subject: %s, To: %s", subject, recipientEmail))
		})

		Context("when to address argument is empty", func() {
			BeforeEach(func() {
				toAddressArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when subject argument is empty", func() {
			BeforeEach(func() {
				subject = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when body html argument is empty", func() {
			BeforeEach(func() {
				bodyHTML = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when kind id argument is empty", func() {
			BeforeEach(func() {
				kindID = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when notifications domain argument is empty", func() {
			BeforeEach(func() {
				notificationsDomainArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when uaa domain argument is empty", func() {
			BeforeEach(func() {
				uaaDomainArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when uaa admin username argument is empty", func() {
			BeforeEach(func() {
				uaaAdminClientUsernameArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when uaa admin secret argument is empty", func() {
			BeforeEach(func() {
				uaaAdminClientSecretArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when uaa username argument is empty", func() {
			BeforeEach(func() {
				uaaClientUsernameArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when uaa secret argument is empty", func() {
			BeforeEach(func() {
				uaaClientSecretArg = ""
			})

			It("exits with error", func() {
				session := runMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})
})
