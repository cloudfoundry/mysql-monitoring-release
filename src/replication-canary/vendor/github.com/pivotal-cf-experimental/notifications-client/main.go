package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"

	"code.cloudfoundry.org/uaa-go-client"
	uaa_config "code.cloudfoundry.org/uaa-go-client/config"
	"code.cloudfoundry.org/uaa-go-client/schema"
	"github.com/pivotal-cf-experimental/notifications-client/notificationemailer"
)

var (
	toAddress              = flag.String("toAddress", "", "email address to send to")
	subject                = flag.String("subject", "", "subject of email")
	bodyHTML               = flag.String("bodyHTML", "", "body of email in HTML")
	kindID                 = flag.String("kindID", "", "a key to identify the type of email to be sent")
	notificationsDomain    = flag.String("notificationsDomain", "", "notifications domain e.g. notifications.bosh-lite.com")
	uaaDomain              = flag.String("uaaDomain", "", "uaa Domain")
	uaaAdminClientUsername = flag.String("uaaAdminClientUsername", "", "uaa Admin Client username")
	uaaAdminClientSecret   = flag.String("uaaAdminClientSecret", "", "uaa Admin Client secret")
	uaaClientUsername      = flag.String("uaaClientUsername", "", "uaa client username")
	uaaClientSecret        = flag.String("uaaClientSecret", "", "uaa client secret")
	skipSSLCertVerify      = flag.Bool("skipSSLCertVerify", false, "skip validation of SSL certificates")
)

type DebugWrapper struct {
	logger lager.Logger
}

func (d DebugWrapper) Debug(action string, message map[string]interface{}) {
	data := lager.Data{}

	for k, v := range message {
		data[k] = v
	}

	d.logger.Debug(action, data)
}

func main() {
	flag.Parse()

	logger := lager.NewLogger("Notifications Emailer")
	outSink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
	logger.RegisterSink(outSink)

	if toAddress == nil || *toAddress == "" {
		log.Fatal("toAddress must be provided and not empty")
	}

	if subject == nil || *subject == "" {
		log.Fatal("subject must be provided and not empty")
	}

	if bodyHTML == nil || *bodyHTML == "" {
		log.Fatal("bodyHTML must be provided and not empty")
	}

	if kindID == nil || *kindID == "" {
		log.Fatal("kindID must be provided and not empty")
	}

	if notificationsDomain == nil || *notificationsDomain == "" {
		log.Fatal("notificationsDomain must be provided and not empty")
	}

	if uaaAdminClientUsername == nil || *uaaAdminClientUsername == "" {
		log.Fatal("uaaAdminClientUsername must be provided and not empty")
	}

	if uaaAdminClientSecret == nil || *uaaAdminClientSecret == "" {
		log.Fatal("uaaAdminClientSecret must be provided and not empty")
	}

	if uaaDomain == nil || *uaaDomain == "" {
		log.Fatal("uaaDomain must be provided and not empty")
	}

	if uaaClientUsername == nil || *uaaClientUsername == "" {
		log.Fatal("uaaClientUsername must be provided and not empty")
	}

	if uaaClientSecret == nil || *uaaClientSecret == "" {
		log.Fatal("uaaClientSecret must be provided and not empty")
	}

	logger.Info("Running with config", lager.Data{
		"toAddress":              toAddress,
		"subject":                subject,
		"kindID":                 kindID,
		"notificationsDomain":    notificationsDomain,
		"uaaDomain":              uaaDomain,
		"uaaAdminClientUsername": uaaAdminClientUsername,
		"uaaClientUsername":      uaaClientUsername,
	})

	notificationsClient := notificationemailer.NewClient(
		fmt.Sprintf("https://%s", *notificationsDomain),
		*skipSSLCertVerify,
		DebugWrapper{logger},
	)

	adminUaaConfig := &uaa_config.Config{
		ClientName:       *uaaAdminClientUsername,
		ClientSecret:     *uaaAdminClientSecret,
		UaaEndpoint:      fmt.Sprintf("https://%s", *uaaDomain),
		SkipVerification: *skipSSLCertVerify,
	}

	logger.Info("adminUAACONFIG:", lager.Data{
		"clientName":       adminUaaConfig.ClientName,
		"UAAEndpoint":      adminUaaConfig.UaaEndpoint,
		"SkipVerification": adminUaaConfig.SkipVerification,
	})

	adminUaaClient, err := uaa_go_client.NewClient(
		logger,
		adminUaaConfig,
		clock.NewClock(),
	)
	if err != nil {
		logger.Fatal("Failed to generate a UAA client", err)
	}

	newUaaClient := &schema.OauthClient{
		ClientId:             *uaaClientUsername,
		Name:                 "Mysql Monitoring",
		ClientSecret:         *uaaClientSecret,
		Scope:                []string{},
		ResourceIds:          []string{},
		Authorities:          []string{"notifications.write", "critical_notifications.write", "emails.write"},
		AuthorizedGrantTypes: []string{"client_credentials"},
		AccessTokenValidity:  3600,
		RedirectUri:          []string{},
	}

	_, err = adminUaaClient.RegisterOauthClient(newUaaClient)
	if err != nil && err != uaa_go_client.ErrClientAlreadyExists {
		logger.Fatal("Failed to register UAA client", err)
	}

	uaaConfig := &uaa_config.Config{
		ClientName:       *uaaClientUsername,
		ClientSecret:     *uaaClientSecret,
		UaaEndpoint:      fmt.Sprintf("https://%s", *uaaDomain),
		SkipVerification: *skipSSLCertVerify,
	}

	uaaClient, err := uaa_go_client.NewClient(logger, uaaConfig, clock.NewClock())
	if err != nil {
		logger.Fatal("Failed to generate a UAA client", err)
	}

	forceUpdate := true
	clientToken, err := uaaClient.FetchToken(forceUpdate)
	if err != nil {
		logger.Fatal("Failed to generate a UAA client", err)
	}

	err = notificationsClient.Email(clientToken.AccessToken, *toAddress, *subject, *bodyHTML, *kindID)
	if err != nil {
		panic(err)
	}

	logger.Info("Successfully sent email")

	os.Exit(0)
}
