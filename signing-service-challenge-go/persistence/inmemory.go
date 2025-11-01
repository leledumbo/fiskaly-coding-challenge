package persistence

import (
	"errors"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
)

type InMemoryDB struct {
	DeviceMap map[string]*domain.Device
}

func (db *InMemoryDB) Save(id string, data *domain.Device) error {
	db.DeviceMap[id] = data
	return nil
}

func (db *InMemoryDB) Load(id string) (*domain.Device, error) {
	if data, ok := db.DeviceMap[id]; ok {
		return data, nil
	} else {
		return nil, errors.New("Device with id " + id + " not found")
	}
}

func (db *InMemoryDB) List() []*domain.Device {
	devices := []*domain.Device{}

	for _, device := range db.DeviceMap {
		devices = append(devices, device)
	}

	return devices
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		DeviceMap: make(map[string]*domain.Device),
	}
}
