package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	PidFile            string `yaml:"PidFile" validate:"nonzero"`
	Port               uint   `yaml:"Port"`
	Username           string `yaml:"Username"`
	Password           string `yaml:"Password"`
	PersistentDiskPath string `yaml:"PersistentDiskPath"`
	EphemeralDiskPath  string `yaml:"EphemeralDiskPath"`
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
