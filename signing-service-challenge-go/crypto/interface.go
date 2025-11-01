package crypto

type Key any

type KeyPair interface {
	PublicKey() Key
	PrivateKey() Key
	Serialize() ([]byte, []byte, error)
	// Given private key bytes, deserialize and assign resulting public and private keys to self
	Deserialize([]byte) error
}

type Algorithm interface {
	GenerateKeyPair() (KeyPair, error)
	// Given PEM encoded private key, return a corresponding KeyPair
	ConstructKeyPair(priv []byte) (KeyPair, error)
	Sign(priv Key, data []byte) ([]byte, error)
	Verify(pub Key, data []byte, signature []byte) ([]byte, error)
}
