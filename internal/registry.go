package internal

import "sync"

type Registry[T any] struct {
	items map[string]T
	mutex sync.RWMutex
}

func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{
		items: make(map[string]T),
	}
}

func (r *Registry[T]) Register(key string, item T) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.items[key] = item
}

func (r *Registry[T]) Remove(key string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.items, key)
}

func (r *Registry[T]) Get(key string) T {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.items[key]
}

func (r *Registry[T]) Items() map[string]T {
	return r.items
}
