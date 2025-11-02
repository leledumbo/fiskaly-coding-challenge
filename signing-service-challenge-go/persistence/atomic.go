package persistence

import (
	"sync"
)

// AtomicStorage wraps Storage with per id locking mechanism which allows concurrent access over
// different IDs, but will mutex concurrent access over the same ID
type AtomicStorage struct {
	base  Storage
	locks sync.Map // map[id]*sync.Mutex
}

// Locks id @id so there's only one thread capable of accessing until Unlock with the same id is called
func (s *AtomicStorage) Lock(id string) {
	mu, _ := s.locks.LoadOrStore(id, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
}

// Unlocks id @id, after the call concurrent access to the id will be allowed again
func (s *AtomicStorage) Unlock(id string) {
	if mu, ok := s.locks.Load(id); ok {
		mu.(*sync.Mutex).Unlock()
	}
}

// NewAtomicStorage wraps any Storage with per-ID concurrency protection
func NewAtomicStorage(base Storage) *AtomicStorage {
	return &AtomicStorage{base: base}
}
