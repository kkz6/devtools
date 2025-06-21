package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/modules"
	"github.com/kkz6/devtools/internal/ui"
)

func main() {
	// Clear screen and show banner
	fmt.Print("\033[H\033[2J")
	ui.ShowBanner()
	time.Sleep(500 * time.Millisecond)

	// Load configuration with animation
	var cfg *config.Config
	err := ui.ShowLoadingAnimation("Loading configuration", func() error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			log.Printf("Warning: Could not load config: %v", err)
			cfg = config.New()
		}
		return nil
	})
	if err != nil {
		os.Exit(1)
	}

	// Register all available modules
	registry := modules.NewRegistry()
	modules.RegisterAll(registry)

	// Show animated menu and get user selection
	selectedModule, err := ui.ShowAnimatedMenu(registry.List())
	if err != nil {
		if err.Error() == "user exited" {
			fmt.Println("\n👋 Thanks for using DevTools! See you next time.")
			os.Exit(0)
		}
		ui.ShowError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}

	// Execute selected module with loading animation
	module, err := registry.Get(selectedModule)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}

	// Clear screen before module execution
	fmt.Print("\033[H\033[2J")
	
	if err := module.Execute(cfg); err != nil {
		ui.ShowError(fmt.Sprintf("Error executing module: %v", err))
		os.Exit(1)
	}

	// Save configuration after execution
	err = ui.ShowLoadingAnimation("Saving configuration", func() error {
		return config.Save(cfg)
	})
	if err != nil {
		log.Printf("Warning: Could not save config: %v", err)
	}

	// Show completion message
	fmt.Println()
	ui.ShowSuccess("✨ Task completed successfully!")
} 