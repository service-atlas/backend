package neo4jrepositories

import (
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Test_mapNodeToTeam(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	later := now.Add(1 * time.Hour).UTC().Truncate(time.Microsecond)

	tests := []struct {
		name        string
		node        neo4j.Node
		wantName    string
		wantId      string
		wantCreated time.Time
		wantUpdated time.Time
		ok          bool
	}{
		{
			name: "all properties present with correct types",
			node: neo4j.Node{Props: map[string]any{
				"name":    "team-a",
				"id":      "abc-123",
				"created": now,
				"updated": later,
			}},
			wantName:    "team-a",
			wantId:      "abc-123",
			wantCreated: now,
			wantUpdated: later,
			ok:          true,
		},
		{
			name: "missing optional properties are zero-valued",
			node: neo4j.Node{Props: map[string]any{
				"name": "only-name",
			}},
			wantName:    "only-name",
			wantId:      "",
			wantCreated: time.Time{},
			wantUpdated: time.Time{},
			ok:          false,
		},
		{
			name: "incorrect types are ignored (leave zero values)",
			node: neo4j.Node{Props: map[string]any{
				"name":    123,          // not a string
				"id":      456,          // not a string
				"created": "2021-01-01", // not time.Time
				"updated": struct{}{},   // not time.Time
			}},
			wantName:    "",
			wantId:      "",
			wantCreated: time.Time{},
			wantUpdated: time.Time{},
			ok:          false,
		},
		{
			name: "extra properties are ignored",
			node: neo4j.Node{Props: map[string]any{
				"name":    "extra",
				"id":      "id-1",
				"created": now,
				"updated": later,
				"foo":     "bar",
			}},
			wantName:    "extra",
			wantId:      "id-1",
			wantCreated: now,
			wantUpdated: later,
			ok:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := MapNodeToTeam(tt.node)

			if ok != tt.ok {
				t.Errorf("Expected ok to be %v, got %v", tt.ok, ok)
			}

			if got.Name != tt.wantName {
				t.Errorf("Name: expected %q, got %q", tt.wantName, got.Name)
			}
			if got.Id != tt.wantId {
				t.Errorf("Id: expected %q, got %q", tt.wantId, got.Id)
			}
			// Created
			if tt.wantCreated.IsZero() {
				if !got.Created.IsZero() {
					t.Errorf("Created: expected zero value, got %v", got.Created)
				}
			} else if !got.Created.Equal(tt.wantCreated) {
				t.Errorf("Created: expected %v, got %v", tt.wantCreated, got.Created)
			}
			// Updated
			if tt.wantUpdated.IsZero() {
				if !got.Updated.IsZero() {
					t.Errorf("Updated: expected zero value, got %v", got.Updated)
				}
			} else if !got.Updated.Equal(tt.wantUpdated) {
				t.Errorf("Updated: expected %v, got %v", tt.wantUpdated, got.Updated)
			}
		})
	}
}

func Test_mapNodeToService(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	later := now.Add(2 * time.Hour).UTC().Truncate(time.Microsecond)

	tests := []struct {
		name            string
		node            neo4j.Node
		wantName        string
		wantDescription string
		wantType        string
		wantId          string
		wantUrl         string
		wantCriticality int
		wantCreated     time.Time
		wantUpdated     time.Time
	}{
		{
			name: "all properties present with correct types",
			node: neo4j.Node{Props: map[string]any{
				"name":        "svc-a",
				"description": "a test service",
				"type":        "api",
				"id":          "svc-123",
				"url":         "https://example.com",
				"created":     now,
				"updated":     later,
				"criticality": 1,
			}},
			wantName:        "svc-a",
			wantDescription: "a test service",
			wantType:        "api",
			wantId:          "svc-123",
			wantUrl:         "https://example.com",
			wantCreated:     now,
			wantUpdated:     later,
			wantCriticality: 1,
		},
		{
			name: "criticality is mapped properly",
			node: neo4j.Node{Props: map[string]any{
				"name":        "svc-a",
				"description": "a test service",
				"type":        "api",
				"id":          "svc-123",
				"url":         "https://example.com",
				"created":     now,
				"updated":     later,
				"criticality": int64(4),
			}},
			wantName:        "svc-a",
			wantDescription: "a test service",
			wantType:        "api",
			wantId:          "svc-123",
			wantUrl:         "https://example.com",
			wantCreated:     now,
			wantUpdated:     later,
			wantCriticality: 4,
		},
		{
			name: "missing optional properties are zero-valued",
			node: neo4j.Node{Props: map[string]any{
				"name": "only-name",
			}},
			wantName:        "only-name",
			wantDescription: "",
			wantType:        "",
			wantId:          "",
			wantUrl:         "",
			wantCreated:     time.Time{},
			wantUpdated:     time.Time{},
			wantCriticality: 0,
		},
		{
			name: "incorrect types are ignored (leave zero values)",
			node: neo4j.Node{Props: map[string]any{
				"name":        123,            // not a string
				"description": []int{1, 2, 3}, // not a string
				"type":        999,            // not a string
				"id":          false,          // not a string
				"url":         struct{}{},     // not a string
				"created":     "yesterday",    // not time.Time
				"updated":     3.14,           // not time.Time
				"criticality": "high",         // not int
			}},
			wantName:        "",
			wantDescription: "",
			wantType:        "",
			wantId:          "",
			wantUrl:         "",
			wantCreated:     time.Time{},
			wantUpdated:     time.Time{},
			wantCriticality: 0,
		},
		{
			name: "extra properties are ignored",
			node: neo4j.Node{Props: map[string]any{
				"name":        "svc-b",
				"description": "desc",
				"type":        "worker",
				"id":          "svc-456",
				"url":         "http://localhost",
				"criticality": 2,
				"created":     now,
				"updated":     later,
				"foo":         "bar",
			}},
			wantName:        "svc-b",
			wantDescription: "desc",
			wantType:        "worker",
			wantId:          "svc-456",
			wantUrl:         "http://localhost",
			wantCriticality: 2,
			wantCreated:     now,
			wantUpdated:     later,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapNodeToService(tt.node)

			if got.Name != tt.wantName {
				t.Errorf("Name: expected %q, got %q", tt.wantName, got.Name)
			}
			if got.Description != tt.wantDescription {
				t.Errorf("Description: expected %q, got %q", tt.wantDescription, got.Description)
			}
			if got.ServiceType != tt.wantType {
				t.Errorf("ServiceType: expected %q, got %q", tt.wantType, got.ServiceType)
			}
			if got.Id != tt.wantId {
				t.Errorf("Id: expected %q, got %q", tt.wantId, got.Id)
			}
			if got.Url != tt.wantUrl {
				t.Errorf("Url: expected %q, got %q", tt.wantUrl, got.Url)
			}
			// Created
			if tt.wantCreated.IsZero() {
				if !got.Created.IsZero() {
					t.Errorf("Created: expected zero value, got %v", got.Created)
				}
			} else if !got.Created.Equal(tt.wantCreated) {
				t.Errorf("Created: expected %v, got %v", tt.wantCreated, got.Created)
			}
			// Updated
			if tt.wantUpdated.IsZero() {
				if !got.Updated.IsZero() {
					t.Errorf("Updated: expected zero value, got %v", got.Updated)
				}
			} else if !got.Updated.Equal(tt.wantUpdated) {
				t.Errorf("Updated: expected %v, got %v", tt.wantUpdated, got.Updated)
			}
			if tt.wantCriticality != got.Criticality {
				t.Errorf("Criticality: expected %d, got %d", tt.wantCriticality, got.Criticality)
			}
		})
	}
}
