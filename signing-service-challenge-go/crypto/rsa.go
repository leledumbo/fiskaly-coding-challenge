package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// RSAKeyPair is a DTO that holds RSA private and public keys.
type RSAKeyPair struct {
	Public  *rsa.PublicKey
	Private *rsa.PrivateKey
}

func (kp *RSAKeyPair) PublicKey() Key {
	return kp.Public
}

func (kp *RSAKeyPair) PrivateKey() Key {
	return kp.Private
}

func (kp *RSAKeyPair) Serialize() ([]byte, []byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(kp.Private)
	publicKeyBytes := x509.MarshalPKCS1PublicKey(kp.Public)

	encodedPrivate := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA_PRIVATE_KEY",
		Bytes: privateKeyBytes,
	})

	encodePublic := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA_PUBLIC_KEY",
		Bytes: publicKeyBytes,
	})

	return encodePublic, encodedPrivate, nil
}

func (kp *RSAKeyPair) Deserialize(privateKeyBytes []byte) error {
	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	kp.Private = privateKey
	kp.Public = &privateKey.PublicKey
	return nil
}

// RSAAlgorithm implements methods to generate RSA key pair as well as to (en|de)crypt data with
type RSAAlgorithm struct {
}

// Generates a new RSAKeyPair.
func (algo *RSAAlgorithm) GenerateKeyPair() (KeyPair, error) {
	// Security has been ignored for the sake of simplicity.
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return nil, err
	}

	return &RSAKeyPair{
		Public:  &key.PublicKey,
		Private: key,
	}, nil
}

// Encrypt encrypts @plainText with a given @pub key
func (algo *RSAAlgorithm) Encrypt(pub Key, plaintext []byte) ([]byte, error) {
	return nil, nil
}

// Decrypt Decrypts @cipherText with a given @priv key
func (algo *RSAAlgorithm) Decrypt(priv Key, ciphertext []byte) ([]byte, error) {
	return nil, nil
}

func init() {
	RegisterAlgorithm("rsa", &RSAAlgorithm{})
}
