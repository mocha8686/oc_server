package internal

import (
	"errors"
)

var IdInUse = errors.New("ID in use")

type Registry struct {
	ids map[string]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		ids: make(map[string]struct{}),
	}
}

func (r *Registry) Register(id string) error {
	if _, ok := r.ids[id]; ok {
		return IdInUse
	}
	r.ids[id] = struct{}{}
	return nil
}

func (r *Registry) Unregister(id string) {
	if _, ok := r.ids[id]; !ok {
		return
	}
	delete(r.ids, id)
}
