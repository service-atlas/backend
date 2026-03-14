package repositories

import (
	"errors"
	"net/url"
	"service-atlas/internal"
	"strings"
	"time"
)

type Service struct {
	Id               string    `json:"id,omitempty"`
	Name             string    `json:"name"`
	ServiceType      string    `json:"type"`
	Description      string    `json:"description"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated,omitempty"`
	Url              string    `json:"url,omitempty"`
	Tier             int       `json:"tier"`
	ArchitectureRole string    `json:"architecture_role,omitempty"`
	Exposure         string    `json:"exposure,omitempty"`
	ImpactDomain     []string  `json:"impact_domain,omitempty"`
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
		return errors.New("tier must be between 0 and 4")
	}
	if service.ArchitectureRole != "" {
		service.ArchitectureRole = strings.ToLower(service.ArchitectureRole)
		if !internal.ArchitectureRole.IsMember(service.ArchitectureRole) {
			return errors.New("invalid architecture role")
		}
	}
	if service.Exposure != "" {
		service.Exposure = strings.ToLower(service.Exposure)
		if !internal.Exposure.IsMember(service.Exposure) {
			return errors.New("invalid exposure")
		}
	}
	if len(service.ImpactDomain) > 0 {
		for i := range service.ImpactDomain {
			service.ImpactDomain[i] = strings.ToLower(service.ImpactDomain[i])
			if !internal.ImpactDomain.IsMember(service.ImpactDomain[i]) {
				return errors.New("invalid impact domain")
			}
		}
	}

	return nil
}
