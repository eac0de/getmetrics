package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type HTTPServerConfig struct {
	Addr            string        `env:"ADDRESS" yaml:"addr"`
	LogLevel        string        `env:"LOG_LEVEL" yaml:"log_level"`
	StoreInterval   time.Duration `yaml:"store_interval"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH" yaml:"file_storage_path"`
	Restore         bool          `env:"RESTORE" yaml:"restore"`
}

type EnvHTTPServerConfig struct {
	HTTPServerConfig
	StoreInterval int `env:"STORE_INTERVAL"`
}

func NewHTTPServerConfig() *HTTPServerConfig {
	config := new(HTTPServerConfig)
	config.ReadYAML("local_config.yml")
	fmt.Println(*config)
	config.ReadServerFlags()
	fmt.Println(*config)
	config.ReadEnvConfig()
	fmt.Println(*config)
	return config
}

func (c *HTTPServerConfig) ReadYAML(filename string) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("read yaml error(1): %s", err.Error())
	}
	err = yaml.NewDecoder(file).Decode(c)
	if err != nil {
		log.Fatalf("read yaml error(2): %s", err.Error())
	}
}

func (c *HTTPServerConfig) ReadServerFlags() {
	storeInterval := int(c.StoreInterval) / int(time.Second)
	flag.StringVar(&c.Addr, "a", c.Addr, "server address")
	flag.StringVar(&c.LogLevel, "ll", c.LogLevel, "server log level")
	flag.IntVar(&storeInterval, "i", storeInterval, "server store interval")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "server file restore path")
	flag.BoolVar(&c.Restore, "r", c.Restore, "server restore")
	flag.Parse()
	c.StoreInterval = time.Duration(storeInterval) * time.Second

}

func (c *HTTPServerConfig) ReadEnvConfig() {
	DurationToInt := func(d time.Duration) int {
		return int(d.Seconds())
	}
	envConfig := EnvHTTPServerConfig{
		HTTPServerConfig: *c,
		StoreInterval:    DurationToInt(c.StoreInterval),
	}
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal(err)
	}
	c.Addr = envConfig.Addr
	c.LogLevel = envConfig.LogLevel
	c.FileStoragePath = envConfig.FileStoragePath
	c.Restore = envConfig.Restore
	c.StoreInterval = time.Duration(envConfig.StoreInterval) * time.Second
}
