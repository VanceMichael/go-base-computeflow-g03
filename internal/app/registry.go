package app

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"sync"
)

type Registry struct {
	mu    sync.RWMutex
	ports map[string]domain.Port
}

func NewRegistry() *Registry { return &Registry{ports: map[string]domain.Port{}} }
func (r *Registry) Put(port domain.Port) error {
	if port.ID == "" || port.Code == "" {
		return fmt.Errorf("%w: incomplete port", domain.ErrInvalid)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.ports[port.ID]; ok && old.Code != port.Code {
		return fmt.Errorf("%w: port identity changed", domain.ErrConflict)
	}
	r.ports[port.ID] = port
	return nil
}
func (r *Registry) Get(id string) (domain.Port, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.ports[id]
	return p, ok
}
func (r *Registry) List() []domain.Port {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Port, 0, len(r.ports))
	for _, p := range r.ports {
		out = append(out, p)
	}
	return out
}
