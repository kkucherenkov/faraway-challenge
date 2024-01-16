package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

type Service struct {
	Port     string `yaml:"port" envconfig:"SERVICE_PORT"`
	Host     string `yaml:"host" envconfig:"SERVICE_HOST"`
	DataFile string `yaml:"dataFile" envconfig:"DATA_FILE"`
}

type Cache struct {
	Port string `yaml:"cachePort" envconfig:"CACHE_PORT"`
	Host string `yaml:"cacheHost" envconfig:"CACHE_HOST"`
}

type Pow struct {
	Complexity    int   `yaml:"complexity" envconfig:"POW_COMPLEXITY"`
	Expiration    int64 `yaml:"expiration" envconfig:"POW_EXPIRATION"`
	MaxIterations int   `yaml:"maxIterations" envconfig:"POW_MAX_ITERATIONS"`
}

type Config struct {
	Service Service `yaml:"service"`
	Cache   Cache   `yaml:"cache"`
	Pow     Pow     `yaml:"pow"`
}

func Load(configFile string) (*Config, error) {
	config := Config{}
	var err error
	err = readFile(configFile, &config)
	err = readEnv(&config)
	return &config, err
}

func readFile(configFile string, cfg *Config) error {
	f, err := os.Open(configFile)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func readEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return err
	}
	return nil
}
