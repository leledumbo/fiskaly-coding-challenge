package persistence

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
)

type Storage interface {
	Save(id string, data *domain.Device) error
	Load(id string) (*domain.Device, error)
	List() []*domain.Device
}
