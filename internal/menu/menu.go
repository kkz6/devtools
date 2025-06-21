package menu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kkz6/devtools/internal/types"
)

// Show displays an interactive menu and returns the selected module ID
func Show(modules []types.ModuleInfo) (string, error) {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║          DevTools Manager              ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println("\nAvailable tools:")
	fmt.Println()

	// Display modules
	for i, module := range modules {
		fmt.Printf("  %d. %s\n", i+1, module.Name)
		fmt.Printf("     %s\n", module.Description)
		fmt.Println()
	}

	fmt.Println("  0. Exit")
	fmt.Println()

	// Get user selection
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Select a tool (0-" + strconv.Itoa(len(modules)) + "): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		if choice == 0 {
			fmt.Println("Goodbye!")
			os.Exit(0)
		}

		if choice < 1 || choice > len(modules) {
			fmt.Println("Invalid selection. Please try again.")
			continue
		}

		return modules[choice-1].ID, nil
	}
} 