package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"

	"github.com/cloudfoundry/replication-canary/alert"
	"github.com/cloudfoundry/replication-canary/canary"
	"github.com/cloudfoundry/replication-canary/clientcreator"
	"github.com/cloudfoundry/replication-canary/config"
	"github.com/cloudfoundry/replication-canary/database"
	"github.com/cloudfoundry/replication-canary/galera"
	"github.com/cloudfoundry/replication-canary/middleware"
	"github.com/cloudfoundry/replication-canary/notifications-client/notificationemailer"
	"github.com/cloudfoundry/replication-canary/switchboard"
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
	appConfig, err := config.NewConfig(os.Args)

	logger := appConfig.Logger

	if err != nil {
		logger.Fatal("Failed to read config", err, lager.Data{
			"config": sanitizeConfig(appConfig),
		})
	}

	err = appConfig.Validate()
	if err != nil {
		logger.Fatal("Failed to validate config", err, lager.Data{
			"config": sanitizeConfig(appConfig),
		})
	}

	logger.Info("Starting replication canary with configuration", lager.Data{
		"mysql": appConfig.MySQL,
	})

	skipSSLCertVerify := appConfig.SkipSSLValidation

	notificationsClient := notificationemailer.NewClient(
		"https://"+appConfig.Notifications.NotificationsDomain,
		skipSSLCertVerify,
		DebugWrapper{logger},
	)

	adminUaaConfig := &uaaconfig.Config{
		ClientName:       appConfig.Notifications.AdminClientUsername,
		ClientSecret:     appConfig.Notifications.AdminClientSecret,
		UaaEndpoint:      "https://" + appConfig.Notifications.UAADomain,
		SkipVerification: skipSSLCertVerify,
	}

	adminUaaClient, err := uaa_go_client.NewClient(
		logger,
		adminUaaConfig,
		clock.NewClock(),
	)
	if err != nil {
		logger.Fatal("Failed to generate a UAA client", err)
	}

	err = clientcreator.CreateClient(adminUaaClient, appConfig)
	if err != nil {
		logger.Fatal("Failed to register UAA client", err)
	}

	uaaConfig := &uaaconfig.Config{
		ClientName:       appConfig.Notifications.ClientUsername,
		ClientSecret:     appConfig.Notifications.ClientSecret,
		UaaEndpoint:      "https://" + appConfig.Notifications.UAADomain,
		SkipVerification: skipSSLCertVerify,
	}

	uaaClient, err := uaa_go_client.NewClient(
		logger,
		uaaConfig,
		clock.NewClock(),
	)
	if err != nil {
		logger.Fatal("Failed to generate a UAA client", err)
	}

	loggingAlerter := &alert.LoggingAlerter{
		Logger: logger,
	}

	emailingAlerter := &alert.EmailingAlerter{
		Logger:              logger,
		UAAClient:           uaaClient,
		NotificationsClient: notificationsClient,
		ToAddress:           appConfig.Notifications.ToAddress,
		SystemDomain:        appConfig.Notifications.SystemDomain,
		ClusterIdentifier:   appConfig.Notifications.ClusterIdentifier,
	}

	aggregateAlerter := alert.AggregateAlerter{
		loggingAlerter,
		emailingAlerter,
	}

	sc := []database.SwitchboardClient{}

	for _, url := range appConfig.Switchboard.URLs {
		switchboardClient := switchboard.NewClient(
			url,
			appConfig.Switchboard.Username,
			appConfig.Switchboard.Password,
			skipSSLCertVerify,
			logger,
		)

		sc = append(sc, switchboardClient)
		switchboardAlerter := &alert.SwitchboardAlerter{
			Logger:            logger,
			SwitchboardClient: switchboardClient,
			NoOp:              appConfig.NotifyOnly,
		}

		aggregateAlerter = append(aggregateAlerter, switchboardAlerter)

	}

	connFactory := database.NewConnectionFactoryFromConfig(appConfig, sc, logger)

	databaseSession := make(map[string]string)
	databaseSession["wsrep_sync_wait"] = "1"

	client := database.NewClient(databaseSession, logger)
	healthchecker := &galera.Client{
		Logger: logger,
	}
	bird := canary.NewCanary(client, healthchecker, time.Duration(appConfig.WriteReadDelay)*time.Second, logger)

	writeConn, err := connFactory.WriteConn()
	if err != nil {
		logger.Fatal("Canary setup failed", err)
	}

	err = bird.Setup(writeConn)
	if err != nil {
		logger.Fatal("Canary setup failed", err)
	}

	coalMiner := canary.NewCoalMiner(
		connFactory,
		bird,
		aggregateAlerter,
		logger,
	)

	go publishApi(coalMiner.StateMachine, appConfig)

	ticker := time.NewTicker(time.Duration(appConfig.PollFrequency) * time.Second)
	logger.Info("ready to sing")
	coalMiner.LetSing(ticker.C)
}

func publishApi(stateMachine canary.StateMachine, cfg *config.Config) {
	insecureStatusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		healthy := stateMachine.GetState() == canary.NotUnhealthy
		fmt.Fprintf(w, "{ \"healthy\": %t }", healthy)
	})

	basicAuth := middleware.NewBasicAuth(cfg.Canary.Username, cfg.Canary.Password)
	secureStatusHandler := basicAuth.Wrap(insecureStatusHandler)

	http.Handle("/api/v1/status", secureStatusHandler)

	bindAddress := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.APIPort)
	fmt.Println(fmt.Sprintf("Listening on: '%s'", bindAddress))

	l, err := cfg.NetworkListener()
	if err != nil {
		panic(err)
	}

	if err := http.Serve(l, nil); err != nil {
		panic(err)
	}
}

func sanitizeConfig(cfg *config.Config) *config.Config {
	cfg.Canary.Password = "REDACTED"

	return cfg
}
