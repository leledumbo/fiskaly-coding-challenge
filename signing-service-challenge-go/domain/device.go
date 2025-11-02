package domain

type Device struct {
	ID string
	// Name of the algorithm chosen for this device
	Algorithm string
	// PEM encoded private key, this is enough for reconstructing the whole key pair
	PrivateKey []byte
	// Optional label, for UI display
	Label string
	// Tracks number of call to Sign() with the same Algorithm
	SignatureCounter int
	// Signature of the last call to Sign() with this device, or simply base64 encoded device ID initially
	LastSignature string
}
