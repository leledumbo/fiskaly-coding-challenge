package crypto_test

import (
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRSAAlgorithm(t *testing.T) {
	Convey("Given an RSAAlgorithm", t, func() {
		algo := &crypto.RSAAlgorithm{}

		Convey("It should generate a valid key pair", func() {
			kp, err := algo.GenerateKeyPair()

			So(err, ShouldBeNil)
			So(kp, ShouldNotBeNil)

			rsaKp, ok := kp.(*crypto.RSAKeyPair)
			So(ok, ShouldBeTrue)
			So(rsaKp.Public, ShouldNotBeNil)
			So(rsaKp.Private, ShouldNotBeNil)
		})

		Convey("ConstructKeyPair should rebuild the same key pair from serialized private key", func() {
			kp, _ := algo.GenerateKeyPair()
			_, priv, _ := kp.(*crypto.RSAKeyPair).Serialize()

			reconstructed, err := algo.ConstructKeyPair(priv)
			So(err, ShouldBeNil)
			So(reconstructed, ShouldNotBeNil)

			newKp := reconstructed.(*crypto.RSAKeyPair)
			So(newKp.Private.D.Cmp(kp.(*crypto.RSAKeyPair).Private.D), ShouldEqual, 0)
			So(newKp.Public.N.Cmp(kp.(*crypto.RSAKeyPair).Public.N), ShouldEqual, 0)
		})

		Convey("ConstructKeyPair should fail with invalid PEM", func() {
			reconstructed, err := algo.ConstructKeyPair([]byte("INVALID PEM DATA"))
			So(reconstructed, ShouldNotBeNil) // returns *RSAKeyPair anyway
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not a valid PEM")
		})

		Convey("Serialize and Deserialize should preserve key values", func() {
			kp, _ := algo.GenerateKeyPair()
			pub, priv, err := kp.(*crypto.RSAKeyPair).Serialize()

			So(err, ShouldBeNil)
			So(pub, ShouldNotBeEmpty)
			So(priv, ShouldNotBeEmpty)

			newKp := &crypto.RSAKeyPair{}
			err = newKp.Deserialize(priv)
			So(err, ShouldBeNil)

			So(newKp.Private.D.Cmp(kp.(*crypto.RSAKeyPair).Private.D), ShouldEqual, 0)
			So(newKp.Public.N.Cmp(kp.(*crypto.RSAKeyPair).Public.N), ShouldEqual, 0)
		})

		Convey("Deserialize should fail on invalid PEM", func() {
			kp := &crypto.RSAKeyPair{}
			err := kp.Deserialize([]byte("INVALID PEM DATA"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not a valid PEM")
		})

		Convey("Deserialize should fail on invalid private key bytes", func() {
			kp := &crypto.RSAKeyPair{}
			invalidPem := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA_PRIVATE_KEY",
				Bytes: []byte("INVALID KEY"),
			})
			err := kp.Deserialize(invalidPem)
			So(err, ShouldNotBeNil)
		})

		Convey("Sign and Verify should succeed on valid data", func() {
			kp, _ := algo.GenerateKeyPair()
			data := []byte("fiskaly test data")

			signature, err := algo.Sign(kp.PrivateKey(), data)
			So(err, ShouldBeNil)
			So(signature, ShouldNotBeEmpty)

			err = algo.Verify(kp.PublicKey(), data, signature)
			So(err, ShouldBeNil)
		})

		Convey("Verify should fail on tampered data", func() {
			kp, _ := algo.GenerateKeyPair()
			data := []byte("original data")
			tampered := []byte("different data")

			signature, _ := algo.Sign(kp.PrivateKey(), data)
			err := algo.Verify(kp.PublicKey(), tampered, signature)
			So(err, ShouldNotBeNil)
		})

		Convey("Sign should fail on non-RSA private key", func() {
			_, err := algo.Sign("not-a-key", []byte("data"))
			So(err, ShouldNotBeNil)
			So(errors.Is(err, rsa.ErrMessageTooLong), ShouldBeFalse) // just ensures it's our own error
		})

		Convey("Verify should fail on non-RSA public key", func() {
			err := algo.Verify("not-a-key", []byte("data"), []byte("sig"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not a RSA")
		})
	})
}
