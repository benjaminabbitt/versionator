package vcs

import (
	"sync"
)

// VCSRegistry manages available version control systems
type VCSRegistry struct {
	systems map[string]VersionControlSystem
	mutex   sync.RWMutex
}

var registry = &VCSRegistry{
	systems: make(map[string]VersionControlSystem),
}

// RegisterVCS registers a version control system
func (r *VCSRegistry) RegisterVCS(vcs VersionControlSystem) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.systems[vcs.Name()] = vcs
}

// UnregisterVCS unregisters a version control system
func (r *VCSRegistry) UnregisterVCS(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.systems, name)
}

// GetActiveVCS returns the first VCS that detects it's in a repository
func (r *VCSRegistry) GetActiveVCS() VersionControlSystem {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, vcs := range r.systems {
		if vcs.IsRepository() {
			return vcs
		}
	}
	return nil
}

// GetVCS returns a specific VCS by name
func (r *VCSRegistry) GetVCS(name string) VersionControlSystem {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.systems[name]
}

// ListVCS returns all registered VCS names
func (r *VCSRegistry) ListVCS() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var names []string
	for name := range r.systems {
		names = append(names, name)
	}
	return names
}

// Global functions for easy access
func RegisterVCS(vcs VersionControlSystem) {
	registry.RegisterVCS(vcs)
}

func UnregisterVCS(name string) {
	registry.UnregisterVCS(name)
}

func GetActiveVCS() VersionControlSystem {
	return registry.GetActiveVCS()
}

func GetVCS(name string) VersionControlSystem {
	return registry.GetVCS(name)
}

func ListVCS() []string {
	return registry.ListVCS()
}
