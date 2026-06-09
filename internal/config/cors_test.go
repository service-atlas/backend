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
	t.Setenv("CORS_CONFIG", `{"AllowedOrigins":["https://example.com","https://api.example.com"],"AllowedMethods":["GET","OPTIONS"],"AllowCredentials":true}`)

	got := GetCORSConfig()

	want := CORSConfig{
		AllowedOrigins:   []string{"https://example.com", "https://api.example.com"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowCredentials: true,
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

func TestGetCORSConfig_AllowCredentialsLogic(t *testing.T) {
	tests := []struct {
		name string
		json string
		want bool
	}{
		{
			name: "wildcard origin forces credentials false",
			json: `{"AllowedOrigins":["*"],"AllowCredentials":true}`,
			want: false,
		},
		{
			name: "wildcard origin with others forces credentials false",
			json: `{"AllowedOrigins":["https://example.com","*"],"AllowCredentials":true}`,
			want: false,
		},
		{
			name: "explicit origins allow credentials true",
			json: `{"AllowedOrigins":["https://example.com"],"AllowCredentials":true}`,
			want: true,
		},
		{
			name: "explicit origins can default credentials to false",
			json: `{"AllowedOrigins":["https://example.com"]}`,
			want: false,
		},
		{
			name: "explicit origins can explicitly set credentials to false",
			json: `{"AllowedOrigins":["https://example.com"],"AllowCredentials":false}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CORS_CONFIG", tt.json)
			got := GetCORSConfig()
			if got.AllowCredentials != tt.want {
				t.Errorf("AllowCredentials = %v, want %v", got.AllowCredentials, tt.want)
			}
		})
	}
}
