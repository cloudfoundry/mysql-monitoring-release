package config

import (
	"flag"

	"code.cloudfoundry.org/lager/lagerflags"
	"github.com/pivotal-cf-experimental/service-config"
	"gopkg.in/validator.v2"

	"errors"

	"code.cloudfoundry.org/lager"
)

type Config struct {
	PidFile           string `yaml:"PidFile" validate:"nonzero"`
	Logger            lager.Logger
	MySQL             MySQL         `yaml:"MySQL" validate:"nonzero"`
	Canary            Canary        `yaml:"Canary" validate:"nonzero"`
	Notifications     Notifications `yaml:"Notifications" validate:"nonzero"`
	Switchboard       Switchboard   `yaml:"Switchboard" validate:"nonzero"`
	WriteReadDelay    int           `yaml:"WriteReadDelay" validate:"nonzero"`
	PollFrequency     int           `yaml:"PollFrequency" validate:"nonzero,min=1"`
	NotifyOnly        bool          `yaml:"NotifyOnly"`
	SkipSSLValidation bool          `yaml:"SkipSSLValidation"`
	APIPort           uint          `yaml:"APIPort"`
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
