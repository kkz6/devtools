package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/modules"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"
)

func main() {
	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("DevTools %s\n", Version)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Author: Karthick <karthick@gigcodes.com>\n")
		fmt.Printf("Website: https://devkarti.com\n")
		os.Exit(0)
	}

	// Clear screen and show banner
	fmt.Print("\033[H\033[2J")
	ui.ShowBanner()
	
	// Add a personalized welcome message
	welcomeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		MarginLeft(2).
		Bold(true)
	fmt.Println(welcomeStyle.Render("✨ Welcome to DevTools by Karthick!"))
	fmt.Println()
	
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

	// Main loop
	for {
		// Clear screen and show banner
		fmt.Print("\033[H\033[2J")
		ui.ShowBanner()
		
		// Show animated menu and get user selection
		selectedModule, err := ui.ShowAnimatedMenu(registry.List())
		if err != nil {
			if err.Error() == "user exited" {
				fmt.Println("\n👋 Thanks for using DevTools! See you next time.")
				os.Exit(0)
			}
			ui.ShowError(fmt.Sprintf("Error: %v", err))
			continue
		}

		// Execute selected module with loading animation
		module, err := registry.Get(selectedModule)
		if err != nil {
			ui.ShowError(fmt.Sprintf("Error: %v", err))
			continue
		}

		// Clear screen before module execution
		fmt.Print("\033[H\033[2J")
		
		if err := module.Execute(cfg); err != nil {
			// Check if user just wants to go back
			if err == types.ErrNavigateBack {
				continue // Just go back to menu without any message
			}
			ui.ShowError(fmt.Sprintf("Error executing module: %v", err))
			continue
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
		fmt.Println()
		
		// Pause before returning to menu
		fmt.Print("Press Enter to return to main menu...")
		fmt.Scanln()
	}
} 