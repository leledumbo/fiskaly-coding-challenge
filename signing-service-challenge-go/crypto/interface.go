package crypto

// An alias to show purpose, as there's no way to abstract keys of different algorithms
type Key any

type KeyPair interface {
	// Returns public key of the pair
	PublicKey() Key
	// Returns private key of the pair
	PrivateKey() Key
	// Return PEM encoded public and private key, in that order
	Serialize() ([]byte, []byte, error)
	// Given PEM encoded private key, deserialize and assign resulting public and private keys to self
	Deserialize([]byte) error
}

type Algorithm interface {
	// Generates a new KeyPair, typically random
	GenerateKeyPair() (KeyPair, error)
	// Given PEM encoded private key, return a corresponding KeyPair
	ConstructKeyPair(priv []byte) (KeyPair, error)
	// Sign signs @data with a @priv key, returning its signature
	Sign(priv Key, data []byte) ([]byte, error)
	// Verify ensures authenticity of @data given a @signature with a @pub key
	Verify(pub Key, data []byte, signature []byte) error
}
