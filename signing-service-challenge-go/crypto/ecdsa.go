package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// ECCKeyPair is a DTO that holds ECC private and public keys
type ECCKeyPair struct {
	Public  *ecdsa.PublicKey
	Private *ecdsa.PrivateKey
}

func (kp *ECCKeyPair) PublicKey() Key {
	return kp.Public
}

func (kp *ECCKeyPair) PrivateKey() Key {
	return kp.Private
}

func (kp *ECCKeyPair) Serialize() ([]byte, []byte, error) {
	privateKeyBytes, err := x509.MarshalECPrivateKey(kp.Private)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(kp.Public)
	if err != nil {
		return nil, nil, err
	}

	encodedPrivate := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	encodedPublic := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return encodedPublic, encodedPrivate, nil
}

func (kp *ECCKeyPair) Deserialize(privateKeyBytes []byte) error {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return errors.New("Given private key is not a valid PEM encoded key")
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	kp.Private = privateKey
	kp.Public = &privateKey.PublicKey
	return nil
}

// ECCAlgorithm implements methods to generate ECC key pair as well as to sign and verify data
type ECCAlgorithm struct {
}

func (algo *ECCAlgorithm) GenerateKeyPair() (KeyPair, error) {
	// Security has been ignored for the sake of simplicity.
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &ECCKeyPair{
		Public:  &key.PublicKey,
		Private: key,
	}, nil
}

func (algo *ECCAlgorithm) ConstructKeyPair(priv []byte) (KeyPair, error) {
	kp := &ECCKeyPair{}
	err := kp.Deserialize(priv)
	return kp, err
}

func (algo *ECCAlgorithm) Sign(priv Key, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	eccPrivateKey, ok := priv.(*ecdsa.PrivateKey)
	if ok {
		return ecdsa.SignASN1(rand.Reader, eccPrivateKey, hashed[:])
	} else {
		return nil, errors.New("Given private key is not an ECC private key")
	}
}

func (algo *ECCAlgorithm) Verify(pub Key, data []byte, signature []byte) error {
	hashed := sha256.Sum256(data)
	eccPublicKey, ok := pub.(*ecdsa.PublicKey)
	if ok {
		verified := ecdsa.VerifyASN1(eccPublicKey, hashed[:], signature[:])
		if verified {
			return nil
		} else {
			return errors.New("Verification failed")
		}
	} else {
		return errors.New("Given private key is not a ECC private key")
	}
}

func init() {
	RegisterAlgorithm("ecc", &ECCAlgorithm{})
}
