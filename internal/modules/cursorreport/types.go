package cursorreport

import "time"

// UsageData represents current usage statistics
type UsageData struct {
	APICalls               int
	APICallsLimit          int
	APICallsPercentage     float64
	TokensUsed             int
	TokensLimit            int
	TokensPercentage       float64
	FastRequests           int
	FastRequestsLimit      int
	FastRequestsPercentage float64
	SlowRequests           int
	SlowRequestsLimit      int
	SlowRequestsPercentage float64
	PeriodStart            time.Time
	PeriodEnd              time.Time
	DaysRemaining          int
}

// CostAnalysis represents cost analysis results
type CostAnalysis struct {
	Plans            []PlanAnalysis
	RecommendedPlan  string
	PotentialSavings float64
	Insights         []string
}

// PlanAnalysis represents analysis for a specific plan
type PlanAnalysis struct {
	Name          string
	MonthlyCost   float64
	UsageCost     float64
	Savings       float64
	IsCurrentPlan bool
}

// UsageHistory represents historical usage data
type UsageHistory struct {
	Days            []DayUsage
	TotalTokens     int
	TotalAPICalls   int
	AvgTokensPerDay int
	AvgCallsPerDay  int
	PeakDay         time.Time
}

// DayUsage represents usage for a single day
type DayUsage struct {
	Date     time.Time
	Tokens   int
	APICalls int
} 