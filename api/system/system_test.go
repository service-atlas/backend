package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSystemGetTime(t *testing.T) {
	req, err := http.NewRequest("GET", "/time", nil)
	rw := httptest.NewRecorder()
	GetTime(rw, req)
	if err != nil {
		t.Errorf("GetTime errored with %s", err.Error())
	}
	if rw.Code != http.StatusOK {
		t.Errorf("GetTime errored with %s", string(rune(rw.Code)))
	}
	r_time, t_err := time.Parse("2006-01-02 15:04:05", rw.Body.String())
	if t_err != nil {
		t.Errorf("Time returned by GetTime errored with %s", t_err.Error())
	}
	if !time.Now().After(r_time) {
		t.Errorf("GetTime time is not before current time, return val: %s", rw.Body.String())
	}
}

func TestGetDbAddress(t *testing.T) {
	req, err := http.NewRequest("GET", "/database", nil)
	rw := httptest.NewRecorder()
	err = os.Setenv("NEO4J_URL", "test_url")
	GetDbAddress(rw, req)
	if err != nil {
		t.Errorf("GetDbAddress errored with %s", err.Error())
	}
	if rw.Code != http.StatusOK {
		t.Errorf("GetDbAddress errored with %s", strconv.Itoa(rw.Code))
	}
}

func TestGetDbAddressError(t *testing.T) {
	req, err := http.NewRequest("GET", "/database", nil)
	rw := httptest.NewRecorder()
	_ = os.Unsetenv("NEO4J_URL")
	GetDbAddress(rw, req)
	if err != nil {
		t.Errorf("GetDbAddress errored with %s", err.Error())
	}
	if rw.Code != http.StatusInternalServerError {
		t.Errorf("GetDbAddress errored with %s", strconv.Itoa(rw.Code))
	}
	b := rw.Body.String()
	if !strings.HasPrefix(b, "Could not find environment variable") {
		t.Errorf("GetDbAddress errored with %s", b)
	}
}

func TestGetVersion(t *testing.T) {
	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(req)
	}
	rw := httptest.NewRecorder()
	GetVersion(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("GetVersion errored with %d", rw.Code)
	}

	var resp struct {
		Version string `json:"version"`
	}
	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Version != Version {
		t.Errorf("GetVersion returned unexpected version: %s, expected: %s", resp.Version, Version)
	}
}
