package handlers

import "fmt"

// Registry holds all available honeypot modules and provides
// lookup by name.
type Registry struct {
	modules map[string]Module
}

// NewRegistry creates an empty module registry.
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]Module),
	}
}

// Register adds a module to the registry. Panics if a module
// with the same name is already registered.
func (r *Registry) Register(m Module) {
	name := m.Name()
	if _, exists := r.modules[name]; exists {
		panic(fmt.Sprintf("module already registered: %s", name))
	}
	r.modules[name] = m
}

// Get returns a module by name.
func (r *Registry) Get(name string) (Module, bool) {
	m, ok := r.modules[name]
	return m, ok
}

// Enabled returns all modules whose names appear in the given list.
// Unknown names are silently skipped.
func (r *Registry) Enabled(names []string) []Module {
	var modules []Module
	for _, name := range names {
		if m, ok := r.modules[name]; ok {
			modules = append(modules, m)
		}
	}
	return modules
}

// All returns every registered module.
func (r *Registry) All() []Module {
	modules := make([]Module, 0, len(r.modules))
	for _, m := range r.modules {
		modules = append(modules, m)
	}
	return modules
}

// Names returns the names of all registered modules.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}
