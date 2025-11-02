package persistence

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
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

// Fulfill Storage interface so it can be used as a Storage, too

func (s *AtomicStorage) Load(id string) (*domain.Device, error) {
	return s.base.Load(id)
}

func (s *AtomicStorage) Save(id string, data *domain.Device) error {
	return s.base.Save(id, data)
}

func (s *AtomicStorage) List() []*domain.Device {
	return s.base.List()
}

// NewAtomicStorage wraps any Storage with per-ID concurrency protection
func NewAtomicStorage(base Storage) *AtomicStorage {
	return &AtomicStorage{base: base}
}
