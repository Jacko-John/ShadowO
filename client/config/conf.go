package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RemoteAddr string `yaml:"remote-addr"`
	Secret     string `yaml:"secret"`
	Mappings   []struct {
		Name       string `yaml:"name"`
		RemotePort string `yaml:"remote-port"`
		LocalAddr  string `yaml:"local-addr"`
	} `yaml:"mappings"`
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
