package datastore

import (
	"errors"
	"sync"
)

type Datastore interface {
	// Would be []byte -> ([]byte, error) in real system instead.
	Get(key string) (interface{}, error)
	Set(key string, val interface{}) error
	Del(key string) error
}

type datastore struct {
	mu      sync.RWMutex
	storage map[string]interface{}
}

func NewDefault() *datastore {
	return &datastore{
		storage: make(map[string]interface{}),
	}
}

func (d *datastore) Get(key string) (interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if val, ok := d.storage[key]; ok {
		return val, nil
	}
	return nil, errors.New("not found")
}

func (d *datastore) Set(key string, val interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.storage[key] = val
	return nil
}

func (d *datastore) Del(key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.storage, key)
	return nil
}
