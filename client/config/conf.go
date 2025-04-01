package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RemoteAddr string `yaml:"remote-addr"`
	Secret     string `yaml:"secret"`
	SkipVerify bool   `yaml:"skip-cert-verify"`
	Mappings   []struct {
		Name       string `yaml:"name"`
		RemotePort string `yaml:"remote-port"`
		LocalAddr  string `yaml:"local-addr"`
	} `yaml:"mappings"`
}

func (c *Config) String() string {
	return fmt.Sprintf("remote-addr: %s | secret: %s | skip-cert-verify: %t | mappings: %v", c.RemoteAddr, c.Secret, c.SkipVerify, c.Mappings)
}

var conf Config

func Init(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		return err
	}

	return nil
}

func Get() *Config {
	return &conf
}
