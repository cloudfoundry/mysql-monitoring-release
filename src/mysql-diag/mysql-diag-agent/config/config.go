package config

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"

	"code.cloudfoundry.org/tlsconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	BindAddress        string `yaml:"BindAddress"`
	Port               uint   `yaml:"Port"`
	Username           string `yaml:"Username"`
	Password           string `yaml:"Password"`
	PersistentDiskPath string `yaml:"PersistentDiskPath"`
	EphemeralDiskPath  string `yaml:"EphemeralDiskPath"`

	TLS struct {
		Enabled     bool   `yaml:"Enabled"`
		Certificate string `yaml:"Certificate"`
		PrivateKey  string `yaml:"PrivateKey"`
	} `yaml:"TLS"`
}

func LoadFromFile(filepath string) (*Config, error) {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var c *Config
	err = yaml.Unmarshal(contents, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c Config) GetPersistentDiskPath() string {
	if c.PersistentDiskPath == "" {
		return "/var/vcap/store"
	} else {
		return c.PersistentDiskPath
	}
}

func (c Config) GetEphemeralDiskPath() string {
	if c.EphemeralDiskPath == "" {
		return "/var/vcap/data"
	} else {
		return c.EphemeralDiskPath
	}
}

func (c *Config) NetworkListener() (net.Listener, error) {
	address := fmt.Sprintf("%s:%d", c.BindAddress, c.Port)

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
