package config

import (
	"encoding/json"
	"log/slog"
)

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func getDefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}
}

func GetCORSConfig() CORSConfig {
	configStr := GetConfigValue("cors_config")
	slog.Debug("CORS config: ", slog.String("config", configStr))
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
