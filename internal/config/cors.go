package config

import (
	"encoding/json"
	"log/slog"
	"slices"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

func getDefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
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

	// Safety check: if AllowedOrigins contains wildcard, we must force AllowCredentials to false
	if slices.Contains(corsConfig.AllowedOrigins, "*") {
		corsConfig.AllowCredentials = false
	}

	return *corsConfig
}
