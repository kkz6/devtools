package githubmanager

import (
	"fmt"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

type Module struct{}

func New() *Module {
	return &Module{}
}

func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "github-manager",
		Name:        "GitHub Repository Manager",
		Description: "Manage GitHub repository settings, actions, and deployments",
	}
}

func (m *Module) Execute(cfg *config.Config) error {
	// Check if GitHub token is configured
	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub token not configured. Please set it in your config file")
	}

	for {
		options := []string{
			"Delete All Action Logs",
			"Delete All Deployments",
			"Back to Main Menu",
		}

		choice, err := ui.SelectFromList("GitHub Repository Manager", options)
		if err != nil {
			return err
		}

		switch choice {
		case 0:
			if err := deleteActionLogs(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to delete action logs: %v", err))
				continue
			}
		case 1:
			if err := deleteDeployments(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to delete deployments: %v", err))
				continue
			}
		case 2:
			return types.ErrNavigateBack
		}
	}
}
