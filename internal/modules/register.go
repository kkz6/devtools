package modules

import (
	"github.com/kkz6/devtools/internal/modules/configmanager"
	"github.com/kkz6/devtools/internal/modules/cursorreport"
	"github.com/kkz6/devtools/internal/modules/gitsigning"
)

// RegisterAll registers all available modules
func RegisterAll(registry *Registry) {
	// Register Configuration Manager module
	registry.Register(configmanager.New())
	
	// Register Git Signing module
	registry.Register(gitsigning.New())
	
	// Register Cursor AI Report module
	registry.Register(cursorreport.New())
	
	// Add more modules here as they are developed
} 