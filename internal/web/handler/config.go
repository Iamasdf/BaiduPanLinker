package handler

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
)

type ServerConfig struct {
	EnableWeb bool `json:"enable_web"`
	EnableAPI bool `json:"enable_api"`
	WebPort   int  `json:"web_port"`
}

func GetServerConfigPath() string {
	return filepath.Join(pcsconfig.GetConfigDir(), "server.json")
}

func LoadServerConfig() (*ServerConfig, error) {
	configPath := GetServerConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ServerConfig{
				EnableWeb: true,
				EnableAPI: true,
				WebPort:   8080,
			}, nil
		}
		return nil, err
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.WebPort == 0 {
		config.WebPort = 8080
	}

	return &config, nil
}

func SaveServerConfig(config *ServerConfig) error {
	configPath := GetServerConfigPath()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
