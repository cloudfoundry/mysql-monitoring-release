package alert

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/lager"

	uaa_go_client "code.cloudfoundry.org/uaa-go-client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./alertfakes/fake_uaa_client.go --fake-name FakeUAAClient ../vendor/code.cloudfoundry.org/uaa-go-client/client.go Client

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . NotificationsClient
type NotificationsClient interface {
	Email(clientToken string, to string, subject string, html string, kindID string) error
}

type EmailingAlerter struct {
	Logger              lager.Logger
	UAAClient           uaa_go_client.Client
	NotificationsClient NotificationsClient
	ToAddress           string
	SystemDomain        string
	ClusterIdentifier   string
}

func (a *EmailingAlerter) NotUnhealthy(now time.Time) error {
	a.Logger.Debug("No action to take for email alerter", lager.Data{"now": now})
	return nil
}

func (a *EmailingAlerter) Unhealthy(now time.Time) error {
	a.Logger.Debug("Email alerter fetching UAA client token", lager.Data{"now": now})

	forceUpdate := true
	token, err := a.UAAClient.FetchToken(forceUpdate)
	if err != nil {
		return err
	}

	to := a.ToAddress
	subject := fmt.Sprintf("[%s][%s] p-mysql Replication Canary, alert 417", a.SystemDomain, a.ClusterIdentifier)
	html := "{alert-code 417}<br/>This is an e-mail to notify you that the MySQL service's replication canary has detected an unsafe cluster condition in which replication is not performing as expected across all nodes."
	kindID := "p-mysql"

	a.Logger.Debug("Email alerter sending email", lager.Data{"now": now})
	return a.NotificationsClient.Email(token.AccessToken, to, subject, html, kindID)
}
