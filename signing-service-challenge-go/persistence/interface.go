package persistence

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
)

type Storage interface {
	Save(id string, data domain.SignatureDevice) error
	Load(id string) (domain.SignatureDevice, error)
}
