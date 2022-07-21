package config

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"

	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/tlsconfig"
	service_config "github.com/pivotal-cf-experimental/service-config"
	"gopkg.in/validator.v2"

	"code.cloudfoundry.org/lager"
)

type Config struct {
	Logger            lager.Logger
	MySQL             MySQL         `yaml:"MySQL" validate:"nonzero"`
	Canary            Canary        `yaml:"Canary" validate:"nonzero"`
	Notifications     Notifications `yaml:"Notifications" validate:"nonzero"`
	Switchboard       Switchboard   `yaml:"Switchboard" validate:"nonzero"`
	WriteReadDelay    int           `yaml:"WriteReadDelay" validate:"nonzero"`
	PollFrequency     int           `yaml:"PollFrequency" validate:"nonzero,min=1"`
	NotifyOnly        bool          `yaml:"NotifyOnly"`
	SkipSSLValidation bool          `yaml:"SkipSSLValidation"`
	BindAddress       string        `yaml:"BindAddress"`
	APIPort           uint          `yaml:"APIPort"`
	TLS               struct {
		Enabled     bool   `yaml:"Enabled"`
		Certificate string `yaml:"Certificate"`
		PrivateKey  string `yaml:"PrivateKey"`
	} `yaml:"TLS"`
}

type MySQL struct {
	ClusterIPs            []string `yaml:"ClusterIPs" validate:"nonzero"`
	Port                  int      `yaml:"Port" validate:"nonzero"`
	GaleraHealthcheckPort int      `yaml:"GaleraHealthcheckPort" validate:"nonzero"`
}

type Canary struct {
	Database string `yaml:"Database" validate:"nonzero"`
	Username string `yaml:"Username" validate:"nonzero"`
	Password string `yaml:"Password" validate:"nonzero"`
}

type Notifications struct {
	AdminClientUsername string `yaml:"AdminClientUsername" validate:"nonzero"`
	AdminClientSecret   string `yaml:"AdminClientSecret" validate:"nonzero"`
	ClientUsername      string `yaml:"ClientUsername" validate:"nonzero"`
	ClientSecret        string `yaml:"ClientSecret" validate:"nonzero"`
	NotificationsDomain string `yaml:"NotificationsDomain" validate:"nonzero"`
	UAADomain           string `yaml:"UAADomain" validate:"nonzero"`
	ToAddress           string `yaml:"ToAddress" validate:"nonzero"`
	SystemDomain        string `yaml:"SystemDomain" validate:"nonzero"`
	ClusterIdentifier   string `yaml:"ClusterIdentifier" validate:"nonzero"`
}

type Switchboard struct {
	URLs     []string `yaml:"URLs" validate:"nonzero"`
	Username string   `yaml:"Username" validate:"nonzero"`
	Password string   `yaml:"Password" validate:"nonzero"`
}

var InvalidDelay = errors.New("WriteReadDelay must be less than the PollFrequency")

func (c Config) Validate() error {
	if c.PollFrequency <= c.WriteReadDelay {
		return InvalidDelay
	}

	return validator.Validate(c)
}

func (c *Config) NetworkListener() (net.Listener, error) {
	address := fmt.Sprintf("%s:%d", c.BindAddress, c.APIPort)

	if !c.TLS.Enabled {
		return net.Listen("tcp", address)
	}

	serverCert, err := tls.X509KeyPair([]byte(c.TLS.Certificate), []byte(c.TLS.PrivateKey))
	if err != nil {
		return nil, err
	}

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentity(serverCert),
	).Server()
	if err != nil {
		return nil, fmt.Errorf("generating tls config: %w", err)
	}

	return tls.Listen("tcp", address, tlsConfig)
}

func NewConfig(osArgs []string) (*Config, error) {
	var rootConfig Config

	binaryName := osArgs[0]
	configurationOptions := osArgs[1:]

	serviceConfig := service_config.New()
	flags := flag.NewFlagSet(binaryName, flag.ExitOnError)

	lagerflags.AddFlags(flags)

	serviceConfig.AddFlags(flags)
	flags.Parse(configurationOptions)

	rootConfig.Logger, _ = lagerflags.NewFromConfig(binaryName, lagerflags.ConfigFromFlags())

	err := serviceConfig.Read(&rootConfig)

	return &rootConfig, err
}
