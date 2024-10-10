package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppLoadAppConfig(t *testing.T) {
	t.Run("test load app config", func(t *testing.T) {
		_, errOpenFile := os.OpenFile("configs/local.yml", os.O_RDONLY, 0666)
		cfg, err := LoadAppConfig()
		if err != nil {
			assert.Equal(t, err.Error(), errOpenFile.Error())
		} else {
			assert.Equal(t, cfg.Addr, "localhost:8080")
		}

	})
}

func TestAppReadYAML(t *testing.T) {
	t.Run("test read yaml", func(t *testing.T) {
		var cfg AppConfig
		_, errOpenFile := os.OpenFile("configs/local.yml", os.O_RDONLY, 0666)
		err := cfg.ReadYAML("configs/local.yml")
		if err != nil {
			assert.Equal(t, err.Error(), errOpenFile.Error())
		} else {
			assert.Equal(t, cfg.Addr, "localhost:8080")
		}

	})
}

func TestAppReadServerFlags(t *testing.T) {
	t.Run("test read server flags", func(t *testing.T) {
		var cfg AppConfig
		cfg.ReadServerFlags()
	})
}

func TestAppReadEnvConfig(t *testing.T) {
	t.Run("test read env config", func(t *testing.T) {
		var cfg AppConfig
		err := cfg.ReadEnvConfig()
		assert.NoError(t, err)
	})
}
