package types

import "github.com/kkz6/devtools/internal/config"

// ModuleInfo contains metadata about a module
type ModuleInfo struct {
	ID          string
	Name        string
	Description string
}

// Module represents a tool module that can be executed
type Module interface {
	Execute(cfg *config.Config) error
	Info() ModuleInfo
} 