package repositories

import (
	"errors"
	"service-atlas/internal"
)

type Dependency struct {
	Id              string `json:"id"`
	Version         string `json:"version,omitempty"`
	Name            string `json:"name,omitempty"`
	ServiceType     string `json:"type,omitempty"`
	InteractionType string `json:"interaction_type,omitempty"`
}

func (d *Dependency) Validate() error {
	if d.Id == "" {
		return errors.New("dependency id is required")
	}
	if d.InteractionType == "" {
		d.InteractionType = "data"
	}
	if !internal.InteractionType.IsMember(d.InteractionType) {
		return errors.New("invalid interaction type")
	}
	return nil
}
