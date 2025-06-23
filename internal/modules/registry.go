package modules

import (
	"fmt"

	"github.com/kkz6/devtools/internal/types"
)

// Registry manages all available modules
type Registry struct {
	modules map[string]types.Module
	order   []string // Maintain registration order
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]types.Module),
		order:   make([]string, 0),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(module types.Module) {
	info := module.Info()
	if _, exists := r.modules[info.ID]; !exists {
		r.order = append(r.order, info.ID)
	}
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

// List returns information about all registered modules in registration order
func (r *Registry) List() []types.ModuleInfo {
	var infos []types.ModuleInfo
	for _, id := range r.order {
		if module, ok := r.modules[id]; ok {
			infos = append(infos, module.Info())
		}
	}
	return infos
}
