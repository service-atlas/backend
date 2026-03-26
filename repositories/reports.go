package repositories

type ServiceRiskReport struct {
	DebtCount      map[string]int64 `json:"debtCount"`
	DependentCount int64            `json:"dependentCount"`
}

type ServiceDebtReport struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Count int64  `json:"count"`
}

// ServiceChangeRisk represents the heuristic change risk classification for a service.
// Risk is one of: "low", "medium", "high". Score is the internal numeric value used to derive Risk.
type ServiceChangeRisk struct {
	Risk  string `json:"risk"`
	Score int    `json:"score,omitempty"`
}

// ComprehensiveServiceRisk represents the risk score for a service, including health and change risk.
type ComprehensiveServiceRisk struct {
	ChangeRisk *ServiceChangeRisk `json:"changeRisk"`
	HealthRisk *ServiceRiskReport `json:"healthRisk"`
}

// ServiceType represents a service type with its type and associated count.
type ServiceType struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}
