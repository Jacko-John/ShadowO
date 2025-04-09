package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr string `yaml:"listen-addr"`
	Secret     string `yaml:"secret"`
	CertFile   struct {
		Key  string `yaml:"key"`
		Cert string `yaml:"cert"`
	} `yaml:"cert-file"`
}

func (c *Config) String() string {
	return fmt.Sprintf("listen-addr: %s | secret: %s | cert-files: %v", c.ListenAddr, c.Secret, c.CertFile)
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
