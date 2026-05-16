package system

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"service-atlas/internal/config"
	"time"
)

var Version = "dev"

func GetTime(rw http.ResponseWriter, _ *http.Request) {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := io.WriteString(rw, now)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
}

func GetVersion(rw http.ResponseWriter, _ *http.Request) {
	resp := struct {
		Version string `json:"version"`
	}{
		Version: Version,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(b)
}

func GetDbAddress(rw http.ResponseWriter, _ *http.Request) {
	url := config.GetConfigValue("neo4j_url")

	if url == "" {
		http.Error(rw, "Could not find environment variable", http.StatusInternalServerError)
		return
	}
	_, err := io.WriteString(rw, url)
	if err != nil {
		log.Println(err)
	}
}
