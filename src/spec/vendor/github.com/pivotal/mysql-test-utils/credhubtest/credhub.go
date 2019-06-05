package credhubtest

import (
	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"github.com/pkg/errors"
)

func NewCredhubClient(credhubServer, clientName, clientSecret, caCert string) (*credhub.CredHub, error) {
	uaaCreds := auth.UaaClientCredentials(clientName, clientSecret)

	chClient, err := credhub.New(credhubServer,
		credhub.CaCerts(caCert),
		credhub.SkipTLSValidation(true),
		credhub.Auth(uaaCreds),
	)

	return chClient, err
}

func GetPassword(client *credhub.CredHub, partialName string) (string, error) {
	results, err := client.FindByPartialName(partialName)
	if err != nil {
		return "", err
	}

	if len(results.Credentials) != 1 {
		return "", errors.Errorf("expected to find exactly one %s, but found %d", partialName, len(results.Credentials))
	}

	credentialName := results.Credentials[0].Name

	pw, err := client.GetLatestPassword(credentialName)
	if err != nil {
		return "", err
	}

	return string(pw.Value), nil
}
