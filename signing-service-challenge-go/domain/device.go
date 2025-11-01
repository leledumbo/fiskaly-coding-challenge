package domain

type Device struct {
	ID               string
	Algorithm        string
	PrivateKey       []byte
	Label            string
	SignatureCounter int
	LastSignature    string
}
