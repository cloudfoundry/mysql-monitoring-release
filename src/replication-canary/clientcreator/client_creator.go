package clientcreator

import (
	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/schema"
	canary_config "replication-canary/config"
)

func CreateClient(adminUaaClient uaa_go_client.Client, config *canary_config.Config) error {
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
	if err != nil && err != uaa_go_client.ErrClientAlreadyExists {
		return err
	}
	return nil
}
