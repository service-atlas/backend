package config

import (
	"encoding/json"
	"log/slog"
)

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
}

func getDefaultCORSConfig() CORSConfig {
	return CORSConfig{AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"}}
}

func GetCORSConfig() CORSConfig {
	configStr := GetConfigValue("cors_config")
	if configStr == "" {
		return getDefaultCORSConfig()
	}
	corsConfig := &CORSConfig{}
	err := json.Unmarshal([]byte(configStr), corsConfig)
	if err != nil {
		slog.Warn("Error parsing CORS config: ", slog.Any("error", err))
		return getDefaultCORSConfig()
	}
	return *corsConfig
}
