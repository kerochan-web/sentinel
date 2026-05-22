package itsm

import (
	"sync"
)

// StateStore abstracts our persistence layers so we can swap out drivers seamlessly.
type StateStore interface {
	Get(serviceName string) (*incidentTracker, error)
	Save(serviceName string, tracker *incidentTracker) error
	Delete(serviceName string) error
}

// InMemStore implements StateStore using a thread-safe map.
type InMemStore struct {
	mu   sync.RWMutex
	data map[string]*incidentTracker
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		data: make(map[string]*incidentTracker),
	}
}

func (s *InMemStore) Get(serviceName string) (*incidentTracker, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[serviceName], nil
}

func (s *InMemStore) Save(serviceName string, tracker *incidentTracker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[serviceName] = tracker
	return nil
}

func (s *InMemStore) Delete(serviceName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, serviceName)
	return nil
}
