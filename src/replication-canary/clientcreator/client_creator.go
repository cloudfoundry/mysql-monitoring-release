package clientcreator

import (
	uaagoclient "code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/schema"

	canaryconfig "github.com/cloudfoundry/replication-canary/config"
)

func CreateClient(adminUaaClient uaagoclient.Client, config *canaryconfig.Config) error {
	newUaaClient := &schema.OauthClient{
		ClientId:             config.Notifications.ClientUsername,
		Name:                 "Mysql Monitoring",
		ClientSecret:         config.Notifications.ClientSecret,
		Scope:                []string{},
		ResourceIds:          []string{},
		Authorities:          []string{"notifications.write", "critical_notifications.write", "emails.write"},
		AuthorizedGrantTypes: []string{"client_credentials"},
		AccessTokenValidity:  3600,
		RedirectUri:          []string{},
	}

	_, err := adminUaaClient.RegisterOauthClient(newUaaClient)
	if err != nil && err != uaagoclient.ErrClientAlreadyExists {
		return err
	}
	return nil
}
