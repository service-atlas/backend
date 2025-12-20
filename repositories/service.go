package repositories

import (
	"errors"
	"net/url"
	"time"
)

type Service struct {
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	ServiceType string    `json:"type"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated,omitempty"`
	Url         string    `json:"url,omitempty"`
	Tier        int       `json:"tier"`
}

func (service *Service) Validate() error {
	switch {
	case service.Name == "":
		return errors.New("service name is required")
	case service.ServiceType == "":
		return errors.New("service type is required")
	}
	//allow url to be optional but validate if passed in
	if service.Url != "" {
		// Validate URL format
		_, err := url.Parse(service.Url)
		if err != nil {
			return errors.New("service url is not a valid URL format")
		}

	}
	if service.Tier == 0 {
		service.Tier = 3
	}
	if service.Tier < 0 || service.Tier > 4 {
		return errors.New("criticality must be between 0 and 4")
	}

	return nil
}
