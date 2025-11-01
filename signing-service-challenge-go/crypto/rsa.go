package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
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
	if block == nil {
		return errors.New("Given private key is not a valid PEM encoded key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	kp.Private = privateKey
	kp.Public = &privateKey.PublicKey
	return nil
}

// RSAAlgorithm implements methods to generate RSA key pair as well as to sign and verify data
type RSAAlgorithm struct {
}

// Generates a new RSAKeyPair
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

func (algo *RSAAlgorithm) ConstructKeyPair(priv []byte) (KeyPair, error) {
	kp := &RSAKeyPair{}
	err := kp.Deserialize(priv)
	return kp, err
}

// Sign signs @data with a @priv key, returning its signature
func (algo *RSAAlgorithm) Sign(priv Key, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	rsaPrivateKey, ok := priv.(*rsa.PrivateKey)
	if ok {
		return rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed[:])
	} else {
		return nil, errors.New("Given private key is not a RSA private key")
	}
}

// Verify ensures authenticity of @data given a @signature with a @pub key
func (algo *RSAAlgorithm) Verify(pub Key, data []byte, signature []byte) ([]byte, error) {
	return nil, nil
}

func init() {
	RegisterAlgorithm("rsa", &RSAAlgorithm{})
}
