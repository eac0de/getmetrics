package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type (
	AgentConfig struct {
		ServerURL      string        `env:"ADDRESS" yaml:"addr"`
		PollInterval   time.Duration `yaml:"poll_interval"`
		ReportInterval time.Duration `yaml:"report_interval"`
		SecretKey      string        `env:"KEY"`
		RateLimit      int           `yaml:"rate_limit"`
	}

	EnvAgentConfig struct {
		AgentConfig
		PollInterval   int `env:"POLL_INTERVAL"`
		ReportInterval int `env:"REPORT_INTERVAL"`
	}
)

func NewAgentConfig() *AgentConfig {
	config := new(AgentConfig)
	config.ReadYAML("local_config.yml")
	config.ReadServerFlags()
	config.ReadEnvConfig()
	return config
}

func (c *AgentConfig) ReadYAML(filename string) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("read yaml error(1): %s", err.Error())
	}
	err = yaml.NewDecoder(file).Decode(c)
	if err != nil {
		log.Fatalf("read yaml error(2): %s", err.Error())
	}
}

func (c *AgentConfig) ReadServerFlags() {
	pollInterval := int(c.PollInterval) / int(time.Second)
	reportInterval := int(c.ReportInterval) / int(time.Second)
	flag.StringVar(&c.ServerURL, "a", c.ServerURL, "server address")
	flag.IntVar(&pollInterval, "p", pollInterval, "report interval in seconds")
	flag.IntVar(&reportInterval, "r", reportInterval, "poll interval in seconds")
	flag.StringVar(&c.SecretKey, "k", c.SecretKey, "secret key")
	flag.IntVar(&c.RateLimit, "l", c.RateLimit, "rate limit")
	flag.Parse()
	c.PollInterval = time.Duration(pollInterval) * time.Second
	c.ReportInterval = time.Duration(reportInterval) * time.Second

}

func (c *AgentConfig) ReadEnvConfig() {
	DurationToInt := func(d time.Duration) int {
		return int(d.Seconds())
	}
	envConfig := EnvAgentConfig{
		AgentConfig:    *c,
		PollInterval:   DurationToInt(c.PollInterval),
		ReportInterval: DurationToInt(c.ReportInterval),
	}
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal(err)
	}
	c.ServerURL = envConfig.ServerURL
	c.PollInterval = time.Duration(envConfig.PollInterval) * time.Second
	c.ReportInterval = time.Duration(envConfig.ReportInterval) * time.Second
	c.SecretKey = envConfig.SecretKey
}
