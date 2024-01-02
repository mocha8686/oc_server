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
	logger := slog.With("id", id)

	logger.Debug("Register")

	if _, ok := r.ids[id]; ok {
		logger.Debug("ID in use")
		return IdInUse
	}
	r.ids[id] = struct{}{}
	return nil
}

func (r *Registry) Unregister(id string) {
	logger := slog.With("id", id)

	logger.Debug("Unregister")

	if _, ok := r.ids[id]; !ok {
		logger.Debug("ID not registered")
		return
	}
	delete(r.ids, id)
}
