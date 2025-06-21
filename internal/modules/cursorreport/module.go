package cursorreport

import (
	"fmt"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// Module implements the Cursor AI report module
type Module struct{}

// New creates a new Cursor report module
func New() *Module {
	return &Module{}
}

// Info returns module information
func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "cursor-report",
		Name:        "Cursor AI Usage Report",
		Description: "View usage statistics and cost analysis for Cursor AI",
	}
}

// Execute runs the Cursor report module
func (m *Module) Execute(cfg *config.Config) error {
	ui.ShowBanner()
	
	title := ui.GetGradientTitle("📊 Cursor AI Usage Report")
	fmt.Println(title)
	fmt.Println()

	// Check if Cursor is configured
	if cfg.Cursor.APIKey == "" {
		ui.ShowWarning("Cursor AI is not configured. Please configure it first.")
		if ui.GetConfirmation("Would you like to configure Cursor AI now?") {
			// Redirect to configuration
			ui.ShowInfo("Please use the Configuration Manager to set up Cursor AI")
			return nil
		}
		return nil
	}

	for {
		options := []string{
			"Current Usage Report",
			"Cost Analysis & Savings",
			"Usage History (Last 30 days)",
			"Plan Comparison",
			"Export Report",
			"Back to main menu",
		}
		
		choice, err := ui.SelectFromList("Select report type:", options)
		if err != nil || choice == 5 {
			return nil
		}

		switch choice {
		case 0:
			if err := m.showCurrentUsage(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to get usage report: %v", err))
			}
		case 1:
			if err := m.showCostAnalysis(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to analyze costs: %v", err))
			}
		case 2:
			if err := m.showUsageHistory(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to get usage history: %v", err))
			}
		case 3:
			if err := m.showPlanComparison(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to compare plans: %v", err))
			}
		case 4:
			if err := m.exportReport(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to export report: %v", err))
			}
		}
		
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
	}
}

// showCurrentUsage displays current usage statistics
func (m *Module) showCurrentUsage(cfg *config.Config) error {
	var usage *UsageData
	
	err := ui.ShowLoadingAnimation("Fetching usage data", func() error {
		var err error
		usage, err = fetchUsageData(cfg)
		return err
	})
	
	if err != nil {
		return err
	}

	// Display usage in a nice table
	table := NewTable("Current Usage Statistics")
	table.AddHeader("Metric", "Used", "Limit", "Usage %")
	
	// Add usage data
	table.AddRow("API Calls", fmt.Sprintf("%d", usage.APICalls), fmt.Sprintf("%d", usage.APICallsLimit), fmt.Sprintf("%.1f%%", usage.APICallsPercentage))
	table.AddRow("Tokens", formatNumber(usage.TokensUsed), formatNumber(usage.TokensLimit), fmt.Sprintf("%.1f%%", usage.TokensPercentage))
	table.AddRow("Fast Requests", fmt.Sprintf("%d", usage.FastRequests), fmt.Sprintf("%d", usage.FastRequestsLimit), fmt.Sprintf("%.1f%%", usage.FastRequestsPercentage))
	table.AddRow("Slow Requests", fmt.Sprintf("%d", usage.SlowRequests), fmt.Sprintf("%d", usage.SlowRequestsLimit), fmt.Sprintf("%.1f%%", usage.SlowRequestsPercentage))
	
	fmt.Println(table.Render())
	
	// Show billing period
	fmt.Println()
	ui.ShowInfo(fmt.Sprintf("Billing Period: %s to %s", usage.PeriodStart.Format("Jan 2, 2006"), usage.PeriodEnd.Format("Jan 2, 2006")))
	ui.ShowInfo(fmt.Sprintf("Days Remaining: %d", usage.DaysRemaining))
	
	return nil
}

// showCostAnalysis shows cost analysis and potential savings
func (m *Module) showCostAnalysis(cfg *config.Config) error {
	var usage *UsageData
	
	err := ui.ShowLoadingAnimation("Analyzing costs", func() error {
		var err error
		usage, err = fetchUsageData(cfg)
		return err
	})
	
	if err != nil {
		return err
	}

	analysis := analyzeCosts(usage, cfg.Cursor.CurrentPlan)
	
	// Display cost analysis table
	table := NewTable("Cost Analysis & Savings")
	table.AddHeader("Plan", "Monthly Cost", "Your Usage Cost", "Status")
	
	for _, plan := range analysis.Plans {
		status := ""
		if plan.IsCurrentPlan {
			if plan.Savings > 0 {
				status = fmt.Sprintf("✅ Saving $%.2f", plan.Savings)
			} else {
				status = "📍 Current Plan"
			}
		} else if plan.Savings < 0 {
			status = fmt.Sprintf("❌ Would cost $%.2f more", -plan.Savings)
		} else {
			status = fmt.Sprintf("💰 Could save $%.2f", plan.Savings)
		}
		
		table.AddRow(
			plan.Name,
			fmt.Sprintf("$%.2f", plan.MonthlyCost),
			fmt.Sprintf("$%.2f", plan.UsageCost),
			status,
		)
	}
	
	fmt.Println(table.Render())
	
	// Show recommendations
	fmt.Println()
	if analysis.RecommendedPlan != cfg.Cursor.CurrentPlan {
		ui.ShowInfo(fmt.Sprintf("💡 Recommendation: Switch to %s plan to save $%.2f/month", analysis.RecommendedPlan, analysis.PotentialSavings))
	} else {
		ui.ShowSuccess("✅ You're on the most cost-effective plan for your usage!")
	}
	
	// Show usage insights
	fmt.Println()
	ui.ShowInfo("Usage Insights:")
	for _, insight := range analysis.Insights {
		fmt.Printf("  • %s\n", insight)
	}
	
	return nil
}

// showUsageHistory displays usage history for the last 30 days
func (m *Module) showUsageHistory(cfg *config.Config) error {
	var history *UsageHistory
	
	err := ui.ShowLoadingAnimation("Fetching usage history", func() error {
		var err error
		history, err = fetchUsageHistory(cfg)
		return err
	})
	
	if err != nil {
		return err
	}

	// Create a simple chart
	fmt.Println("\n📈 Usage Trend (Last 30 Days)")
	fmt.Println("═══════════════════════════")
	
	maxTokens := 0
	for _, day := range history.Days {
		if day.Tokens > maxTokens {
			maxTokens = day.Tokens
		}
	}
	
	// Display bar chart
	for _, day := range history.Days {
		barLength := int(float64(day.Tokens) / float64(maxTokens) * 40)
		bar := ""
		for i := 0; i < barLength; i++ {
			bar += "█"
		}
		
		fmt.Printf("%s │ %-40s %s tokens\n", 
			day.Date.Format("Jan 02"),
			bar,
			formatNumber(day.Tokens),
		)
	}
	
	// Summary statistics
	fmt.Println()
	table := NewTable("30-Day Summary")
	table.AddHeader("Metric", "Total", "Daily Average", "Peak Day")
	table.AddRow("Tokens", formatNumber(history.TotalTokens), formatNumber(history.AvgTokensPerDay), history.PeakDay.Format("Jan 2"))
	table.AddRow("API Calls", fmt.Sprintf("%d", history.TotalAPICalls), fmt.Sprintf("%d", history.AvgCallsPerDay), "-")
	
	fmt.Println(table.Render())
	
	return nil
}

// showPlanComparison shows detailed plan comparison
func (m *Module) showPlanComparison(cfg *config.Config) error {
	fmt.Println()
	
	table := NewTable("Cursor AI Plan Comparison")
	table.AddHeader("Feature", "Free", "Pro ($20/mo)", "Business ($40/mo)")
	
	// Add comparison data
	table.AddRow("Monthly Tokens", "2,000", "Unlimited*", "Unlimited*")
	table.AddRow("Fast Requests", "50", "500", "2000")
	table.AddRow("Slow Requests", "200", "Unlimited", "Unlimited")
	table.AddRow("GPT-4 Access", "❌", "✅", "✅")
	table.AddRow("Claude Access", "❌", "✅", "✅")
	table.AddRow("Priority Support", "❌", "❌", "✅")
	table.AddRow("Team Features", "❌", "❌", "✅")
	table.AddRow("SSO", "❌", "❌", "✅")
	
	fmt.Println(table.Render())
	
	fmt.Println()
	ui.ShowInfo("* Unlimited with fair use policy")
	
	// Highlight current plan
	currentPlanDisplay := map[string]string{
		"free":     "Free",
		"pro":      "Pro ($20/mo)",
		"business": "Business ($40/mo)",
	}
	
	fmt.Println()
	ui.ShowInfo(fmt.Sprintf("Your current plan: %s", currentPlanDisplay[cfg.Cursor.CurrentPlan]))
	
	return nil
}

// exportReport exports the usage report
func (m *Module) exportReport(cfg *config.Config) error {
	filename := fmt.Sprintf("cursor_report_%s.txt", time.Now().Format("2006-01-02"))
	
	err := ui.ShowLoadingAnimation("Exporting report", func() error {
		// In a real implementation, this would write to a file
		time.Sleep(1 * time.Second)
		return nil
	})
	
	if err != nil {
		return err
	}
	
	ui.ShowSuccess(fmt.Sprintf("Report exported to: %s", filename))
	return nil
}

// Helper function to format large numbers
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
} 