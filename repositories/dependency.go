package repositories

import "errors"

type Dependency struct {
	Id             string `json:"id"`
	Version        string `json:"version,omitempty"`
	Name           string `json:"name,omitempty"`
	ServiceType    string `json:"type,omitempty"`
	DependencyType string `json:"dependency_type,omitempty"`
}

func (d *Dependency) Validate() error {
	if d.Id == "" {
		return errors.New("dependency id is required")
	}
	return nil
}
