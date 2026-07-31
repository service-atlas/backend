package services

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"service-atlas/repositories"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPOSTSuccess(t *testing.T) {

	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any
				m = append(m, map[string]any{
					"id":   "1",
					"name": "test",
					"url":  "http://test.com",
					"type": "service",
				})
				return m
			},
			Err: nil,
		},
	}

	service := repositories.Service{
		Name:        "MockService",
		ServiceType: "Internal",
		Description: "Unit test service",
		Url:         "http://test.com",
	}
	serviceJson, err := json.Marshal(&service)
	req, err := http.NewRequest("POST", "/hello/world?name=test", io.NopCloser(strings.NewReader(string(serviceJson))))

	rw := httptest.NewRecorder()
	handler.CreateService(rw, req)
	if err != nil {
		t.Errorf("Service POST errored with %s", err.Error())
	}
	if rw.Code != http.StatusCreated {
		t.Errorf("Service POST errored with %s", strconv.Itoa(rw.Code))
	}
	returnedService := &repositories.Service{}
	err = json.Unmarshal(rw.Body.Bytes(), &returnedService)
	switch {
	case err != nil:
		t.Errorf("Service POST errored parsing return body with %s", err.Error())
	case returnedService.Id != "1":
		t.Errorf("Service POST errored with %s", returnedService.Id)
	case returnedService.Name != "MockService":
		t.Errorf("Service POST errored with Name inconsistency %s", returnedService.Name)
	case returnedService.Description != "Unit test service":
		t.Errorf("Service POST errored with Description inconsistency %s", returnedService.Description)
	case returnedService.ServiceType != "Internal":
		t.Errorf("Service POST errored with ServiceType inconsistency %s", returnedService.ServiceType)
	case returnedService.Url != "http://test.com":
		t.Errorf("Service POST errored with Url inconsistency %s", returnedService.Url)
	case returnedService.Created.IsZero():
		t.Errorf("Service POST errored: Created time is zero")
	case returnedService.Updated.IsZero():
		t.Errorf("Service POST errored: Updated time is zero")
	case time.Since(returnedService.Created) > 5*time.Second:
		t.Errorf("Service POST errored: Created time is not approximately now: %v", returnedService.Created)
	case time.Since(returnedService.Updated) > 5*time.Second:
		t.Errorf("Service POST errored: Updated time is not approximately now: %v", returnedService.Updated)
	}

}

func TestPOSTError(t *testing.T) {
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any
				m = append(m, map[string]any{
					"id":   "1",
					"name": "test",
					"url":  "http://test.com",
					"type": "service",
				})
				return m
			},
			Err: errors.New("test error"),
		},
	}

	service := repositories.Service{
		Name:        "MockService",
		ServiceType: "Internal",
		Description: "Unit test service",
		Url:         "http://test.com",
	}
	serviceJson, err := json.Marshal(&service)
	req, err := http.NewRequest("POST", "/hello/world?name=test", io.NopCloser(strings.NewReader(string(serviceJson))))
	if err != nil {
		panic(err)
	}
	rw := httptest.NewRecorder()
	handler.CreateService(rw, req)
	if rw.Code != http.StatusInternalServerError {
		t.Errorf("Service POST errored with %s", strconv.Itoa(rw.Code))
	}
}

func TestPOSTInvalidBody(t *testing.T) {
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any
				m = append(m, map[string]any{
					"id": "1",
				})
				return m
			},
			Err: nil,
		},
	}
	req, err := http.NewRequest("POST", "/hello/world?name=test", io.NopCloser(strings.NewReader("some text")))

	rw := httptest.NewRecorder()
	handler.CreateService(rw, req)
	if err != nil {
		t.Errorf("Service POST errored with %s", err.Error())
	}

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Service POST errored with %s", strconv.Itoa(rw.Code))
	}
}
