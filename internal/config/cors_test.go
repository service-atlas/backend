package config

import (
	"reflect"
	"testing"
)

func TestGetCORSConfig_DefaultWhenMissing(t *testing.T) {
	t.Setenv("CORS_CONFIG", "")

	got := GetCORSConfig()
	want := getDefaultCORSConfig()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("default CORS config mismatch. got=%v want=%v", got, want)
	}
}

func TestGetCORSConfig_ParsesValidJSON(t *testing.T) {
	t.Setenv("CORS_CONFIG", `{"AllowedOrigins":["https://example.com","https://api.example.com"],"AllowedMethods":["GET","OPTIONS"]}`)

	got := GetCORSConfig()

	want := CORSConfig{
		AllowedOrigins: []string{"https://example.com", "https://api.example.com"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parsed CORS config mismatch.\n got=%v\nwant=%v", got, want)
	}
}

func TestGetCORSConfig_InvalidJSONFallsBackToDefault(t *testing.T) {
	t.Setenv("CORS_CONFIG", `{"AllowedOrigins": ["https://example.com",]}`) // invalid JSON (trailing comma)

	got := GetCORSConfig()
	want := getDefaultCORSConfig()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("invalid JSON should fall back to default. got=%v want=%v", got, want)
	}
}
