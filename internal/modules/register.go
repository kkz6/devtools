package modules

import (
	"github.com/kkz6/devtools/internal/modules/bugmanager"
	"github.com/kkz6/devtools/internal/modules/configmanager"
	"github.com/kkz6/devtools/internal/modules/cursorreport"
	"github.com/kkz6/devtools/internal/modules/fluttermanager"
	"github.com/kkz6/devtools/internal/modules/githubmanager"
	"github.com/kkz6/devtools/internal/modules/gitsigning"
	"github.com/kkz6/devtools/internal/modules/releasemanager"
)

// RegisterAll registers all available modules
func RegisterAll(registry *Registry) {
	// 1. Issue Manager (Bug Manager)
	registry.Register(bugmanager.New())

	// 2. Flutter Application Manager
	registry.Register(fluttermanager.New())

	// 3. Release Manager
	registry.Register(releasemanager.New())

	// 4. GitHub Repository Manager
	registry.Register(githubmanager.New())

	// 5. Cursor AI Usage Reporter
	registry.Register(cursorreport.New())

	// 6. Git Commit Signing Setup
	registry.Register(gitsigning.New())

	// 7. Configuration Manager
	registry.Register(configmanager.New())

	// Add more modules here as they are developed
}
