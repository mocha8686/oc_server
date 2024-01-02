package internal

import (
	"errors"
	"log/slog"
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
	slog.Debug("Register", "id", id)

	if _, ok := r.ids[id]; ok {
		slog.Debug("ID in use", "id", id)
		return IdInUse
	}
	r.ids[id] = struct{}{}
	return nil
}

func (r *Registry) Unregister(id string) {
	slog.Debug("Unregister", "id", id)

	if _, ok := r.ids[id]; !ok {
		slog.Debug("ID not registered", "id", id)
		return
	}
	delete(r.ids, id)
}
