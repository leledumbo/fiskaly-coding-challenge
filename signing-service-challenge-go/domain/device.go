package domain

type SignatureDevice interface {
	GetID() string
	GetAlgorithm() string
	GetPrivateKey() []byte
	GetLabel() string
}

type Device struct {
	ID               string
	Algorithm        string
	PrivateKey       []byte
	Label            string
	SignatureCounter int
}

func (d *Device) GetID() string {
	return d.ID
}

func (d *Device) GetAlgorithm() string {
	return d.Algorithm
}

func (d *Device) GetPrivateKey() []byte {
	return d.PrivateKey
}

func (d *Device) GetLabel() string {
	return d.Label
}
