package crypto

type Key any

type KeyPair interface {
	PublicKey() Key
	PrivateKey() Key
	Serialize() ([]byte, []byte, error)
	Deserialize([]byte) error
}

type Algorithm interface {
	GenerateKeyPair() (KeyPair, error)
	Encrypt(pub Key, plaintext []byte) ([]byte, error)
	Decrypt(priv Key, ciphertext []byte) ([]byte, error)
}
