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
