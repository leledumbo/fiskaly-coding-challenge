package persistence

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
)

type Storage interface {
	// Save Device @data to underlying storage with id @id, may return an error on failure
	Save(id string, data *domain.Device) error
	// Load Device from underlying storage with id @id, may return an error on failure such as no Device with given id exists
	Load(id string) (*domain.Device, error)
	// List all Device-s
	List() []*domain.Device
}
