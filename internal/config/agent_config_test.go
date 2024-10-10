package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentLoadAppConfig(t *testing.T) {
	t.Run("test load agent config", func(t *testing.T) {
		_, errOpenFile := os.OpenFile("configs/local.yml", os.O_RDONLY, 0666)
		cfg, err := LoadAgentConfig()
		if err != nil {
			assert.Equal(t, err.Error(), errOpenFile.Error())
		} else {
			assert.Equal(t, cfg.ServerURL, "localhost:8080")
		}

	})
}

func TestAgentReadYAML(t *testing.T) {
	t.Run("test agent read yaml", func(t *testing.T) {
		var cfg AgentConfig
		_, errOpenFile := os.OpenFile("configs/local.yml", os.O_RDONLY, 0666)
		err := cfg.ReadYAML("configs/local.yml")
		if err != nil {
			assert.Equal(t, err.Error(), errOpenFile.Error())
		} else {
			assert.Equal(t, cfg.ServerURL, "localhost:8080")
		}

	})
}

// func TestAgentReadServerFlags(t *testing.T) {
// 	t.Run("test read agent flags", func(t *testing.T) {
// 		var cfg AgentConfig
// 		cfg.ReadServerFlags()
// 	})
// }

func TestAgentReadEnvConfig(t *testing.T) {
	t.Run("test read agent env config", func(t *testing.T) {
		var cfg AgentConfig
		err := cfg.ReadEnvConfig()
		assert.NoError(t, err)
	})
}
