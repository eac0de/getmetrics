package config

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Addr            string        `env:"ADDRESS" yaml:"addr"`
	LogLevel        string        `env:"LOG_LEVEL" yaml:"log_level"`
	StoreInterval   time.Duration `yaml:"store_interval"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH" yaml:"file_storage_path"`
	Restore         bool          `env:"RESTORE" yaml:"restore"`
	DatabaseDSN     string        `env:"DATABASE_DSN" yaml:"database_dsn"`
	SecretKey       string        `env:"KEY"`
}

type EnvAppConfig struct {
	AppConfig
	StoreInterval int `env:"STORE_INTERVAL"`
}

func LoadAppConfig() (*AppConfig, error) {
	config := new(AppConfig)
	err := config.ReadYAML("configs/local.yml")
	if err != nil {
		return nil, err
	}
	config.ReadServerFlags()
	err = config.ReadEnvConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *AppConfig) ReadYAML(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	err = yaml.NewDecoder(file).Decode(c)
	if err != nil {
		return err
	}
	return nil
}

func (c *AppConfig) ReadServerFlags() {
	storeInterval := int(c.StoreInterval) / int(time.Second)
	flag.StringVar(&c.Addr, "a", c.Addr, "server address")
	flag.StringVar(&c.LogLevel, "ll", c.LogLevel, "server log level")
	flag.IntVar(&storeInterval, "i", storeInterval, "server store interval")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "server file restore path")
	flag.BoolVar(&c.Restore, "r", c.Restore, "server restore")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "db address")
	flag.StringVar(&c.SecretKey, "k", c.SecretKey, "secret key")
	flag.Parse()
	c.StoreInterval = time.Duration(storeInterval) * time.Second

}

func (c *AppConfig) ReadEnvConfig() error {
	DurationToInt := func(d time.Duration) int {
		return int(d.Seconds())
	}
	envConfig := EnvAppConfig{
		AppConfig:     *c,
		StoreInterval: DurationToInt(c.StoreInterval),
	}
	err := env.Parse(&envConfig)
	if err != nil {
		return err
	}
	c.Addr = envConfig.Addr
	c.LogLevel = envConfig.LogLevel
	c.FileStoragePath = envConfig.FileStoragePath
	c.Restore = envConfig.Restore
	c.StoreInterval = time.Duration(envConfig.StoreInterval) * time.Second
	c.DatabaseDSN = envConfig.DatabaseDSN
	c.SecretKey = envConfig.SecretKey
	return nil
}
