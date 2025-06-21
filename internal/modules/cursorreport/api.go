package cursorreport

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kkz6/devtools/internal/config"
)

// fetchUsageData fetches current usage data from Cursor API
func fetchUsageData(cfg *config.Config) (*UsageData, error) {
	// In a real implementation, this would make an actual API call
	// For now, we'll return mock data
	
	// Mock API call simulation
	time.Sleep(500 * time.Millisecond)
	
	// Calculate mock data based on current plan
	var tokenLimit, fastLimit, slowLimit int
	switch cfg.Cursor.CurrentPlan {
	case "free":
		tokenLimit = 2000
		fastLimit = 50
		slowLimit = 200
	case "pro":
		tokenLimit = 1000000 // "Unlimited"
		fastLimit = 500
		slowLimit = 10000 // "Unlimited"
	case "business":
		tokenLimit = 1000000 // "Unlimited"
		fastLimit = 2000
		slowLimit = 10000 // "Unlimited"
	}
	
	// Generate realistic usage data
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)
	daysInMonth := periodEnd.Day()
	daysPassed := now.Day()
	
	// Calculate usage based on days passed
	usagePercent := float64(daysPassed) / float64(daysInMonth)
	
	tokensUsed := int(float64(tokenLimit) * usagePercent * 0.7) // 70% usage rate
	apiCalls := int(500 * usagePercent)
	fastRequests := int(float64(fastLimit) * usagePercent * 0.6)
	slowRequests := int(float64(slowLimit) * usagePercent * 0.8)
	
	return &UsageData{
		APICalls:               apiCalls,
		APICallsLimit:          1000,
		APICallsPercentage:     float64(apiCalls) / 1000 * 100,
		TokensUsed:             tokensUsed,
		TokensLimit:            tokenLimit,
		TokensPercentage:       float64(tokensUsed) / float64(tokenLimit) * 100,
		FastRequests:           fastRequests,
		FastRequestsLimit:      fastLimit,
		FastRequestsPercentage: float64(fastRequests) / float64(fastLimit) * 100,
		SlowRequests:           slowRequests,
		SlowRequestsLimit:      slowLimit,
		SlowRequestsPercentage: float64(slowRequests) / float64(slowLimit) * 100,
		PeriodStart:            periodStart,
		PeriodEnd:              periodEnd,
		DaysRemaining:          daysInMonth - daysPassed,
	}, nil
}

// fetchUsageHistory fetches historical usage data
func fetchUsageHistory(cfg *config.Config) (*UsageHistory, error) {
	// Mock API call
	time.Sleep(500 * time.Millisecond)
	
	history := &UsageHistory{
		Days: make([]DayUsage, 30),
	}
	
	// Generate mock historical data
	now := time.Now()
	totalTokens := 0
	totalCalls := 0
	maxTokens := 0
	var peakDay time.Time
	
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		// Simulate varying usage with weekends having less usage
		baseTokens := 50000
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			baseTokens = 20000
		}
		
		// Add some randomness
		tokens := baseTokens + (i * 1000) - (i * i * 50)
		if tokens < 0 {
			tokens = 10000
		}
		
		calls := tokens / 100
		
		history.Days[29-i] = DayUsage{
			Date:     date,
			Tokens:   tokens,
			APICalls: calls,
		}
		
		totalTokens += tokens
		totalCalls += calls
		
		if tokens > maxTokens {
			maxTokens = tokens
			peakDay = date
		}
	}
	
	history.TotalTokens = totalTokens
	history.TotalAPICalls = totalCalls
	history.AvgTokensPerDay = totalTokens / 30
	history.AvgCallsPerDay = totalCalls / 30
	history.PeakDay = peakDay
	
	return history, nil
}

// analyzeCosts analyzes costs across different plans
func analyzeCosts(usage *UsageData, currentPlan string) *CostAnalysis {
	plans := []PlanAnalysis{
		{
			Name:          "Free",
			MonthlyCost:   0,
			IsCurrentPlan: currentPlan == "free",
		},
		{
			Name:          "Pro",
			MonthlyCost:   20,
			IsCurrentPlan: currentPlan == "pro",
		},
		{
			Name:          "Business",
			MonthlyCost:   40,
			IsCurrentPlan: currentPlan == "business",
		},
	}
	
	// Calculate usage costs for each plan
	for i := range plans {
		plans[i].UsageCost = calculateUsageCost(usage, plans[i].Name)
		if plans[i].IsCurrentPlan {
			plans[i].Savings = 0
		} else {
			currentCost := 0.0
			for _, p := range plans {
				if p.IsCurrentPlan {
					currentCost = p.MonthlyCost + p.UsageCost
					break
				}
			}
			plans[i].Savings = currentCost - (plans[i].MonthlyCost + plans[i].UsageCost)
		}
	}
	
	// Find recommended plan
	recommendedPlan := currentPlan
	maxSavings := 0.0
	
	for _, plan := range plans {
		if plan.Savings > maxSavings {
			maxSavings = plan.Savings
			switch plan.Name {
			case "Free":
				recommendedPlan = "free"
			case "Pro":
				recommendedPlan = "pro"
			case "Business":
				recommendedPlan = "business"
			}
		}
	}
	
	// Generate insights
	insights := []string{}
	
	if usage.TokensPercentage > 80 {
		insights = append(insights, "High token usage detected. Consider upgrading for better limits.")
	}
	
	if usage.FastRequestsPercentage > 90 {
		insights = append(insights, "Near fast request limit. This may slow down your workflow.")
	}
	
	if currentPlan == "free" && usage.TokensUsed > 1500 {
		insights = append(insights, "You're approaching the free tier token limit.")
	}
	
	if currentPlan != "free" && usage.TokensPercentage < 10 {
		insights = append(insights, "Low usage detected. You might save money on a lower tier.")
	}
	
	return &CostAnalysis{
		Plans:            plans,
		RecommendedPlan:  recommendedPlan,
		PotentialSavings: maxSavings,
		Insights:         insights,
	}
}

// calculateUsageCost calculates overage costs for a given plan
func calculateUsageCost(usage *UsageData, planName string) float64 {
	// In free plan, no overage is possible (hard limits)
	if planName == "Free" {
		return 0
	}
	
	// For paid plans, calculate any overage costs
	// This is simplified - real implementation would use actual pricing
	overageCost := 0.0
	
	// Example: If exceeding fast requests on Pro plan
	if planName == "Pro" && usage.FastRequests > 500 {
		overage := usage.FastRequests - 500
		overageCost += float64(overage) * 0.01 // $0.01 per extra fast request
	}
	
	return overageCost
}

// makeAPIRequest makes a request to the Cursor API
func makeAPIRequest(cfg *config.Config, endpoint string) ([]byte, error) {
	if cfg.Cursor.APIEndpoint == "" {
		cfg.Cursor.APIEndpoint = "https://api.cursor.sh/v1"
	}
	
	url := fmt.Sprintf("%s%s", cfg.Cursor.APIEndpoint, endpoint)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Cursor.APIKey))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, string(body))
	}
	
	return io.ReadAll(resp.Body)
} 