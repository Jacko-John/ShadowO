package config

type Config struct {
	ListenAddr string `yaml:"listen-addr"`
	Secret     string `yaml:"secret"`
	CertFile   struct {
		Key  string `yaml:"key"`
		Cert string `yaml:"cert"`
	} `yaml:"cert-file"`
}
