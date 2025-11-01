package persistence

import (
	"errors"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
)

type InMemoryDB struct {
	DeviceMap map[string]domain.SignatureDevice
}

func (db *InMemoryDB) Save(id string, data domain.SignatureDevice) error {
	db.DeviceMap[id] = data
	return nil
}

func (db *InMemoryDB) Load(id string) (domain.SignatureDevice, error) {
	if data, ok := db.DeviceMap[id]; ok {
		return data, nil
	} else {
		return nil, errors.New("Device with id " + id + " not found")
	}
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		DeviceMap: make(map[string]domain.SignatureDevice),
	}
}
