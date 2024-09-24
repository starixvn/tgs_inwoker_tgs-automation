package config

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigManager struct {
	Loader   ConfigLoader
	Env      EnvGetter
	FilePath string
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Loader:   &FileConfigLoader{},
		Env:      &DefaultEnvGetter{},
		FilePath: os.Getenv("CONFIGPATH"), // 默認路徑從環境變數獲取
	}
}

func (c *ConfigManager) GetConfig(ctx context.Context) (TgsConfig, error) {
	data, err := c.Loader.LoadConfig(ctx, c.FilePath)
	if err != nil {
		return TgsConfig{}, fmt.Errorf("fail to load config: %s, %w", c.FilePath, err)
	}

	var config TgsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return TgsConfig{}, fmt.Errorf("fail to parse config: %w", err)
	}

	return config, nil
}

type ConfigLoader interface {
	LoadConfig(ctx context.Context, path string) ([]byte, error)
}

type FileConfigLoader struct{}

func (f *FileConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

type EnvGetter interface {
	Get(key string) string
}

type DefaultEnvGetter struct{}

func (e *DefaultEnvGetter) Get(key string) string {
	return os.Getenv(key)
}
