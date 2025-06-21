package modules

import (
	"fmt"

	"github.com/kkz6/devtools/internal/types"
)



// Registry manages all available modules
type Registry struct {
	modules map[string]types.Module
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]types.Module),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(module types.Module) {
	info := module.Info()
	r.modules[info.ID] = module
}

// Get retrieves a module by ID
func (r *Registry) Get(id string) (types.Module, error) {
	module, ok := r.modules[id]
	if !ok {
		return nil, fmt.Errorf("module not found: %s", id)
	}
	return module, nil
}

// List returns information about all registered modules
func (r *Registry) List() []types.ModuleInfo {
	var infos []types.ModuleInfo
	for _, module := range r.modules {
		infos = append(infos, module.Info())
	}
	return infos
} 