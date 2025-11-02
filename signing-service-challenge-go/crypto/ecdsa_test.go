package crypto_test

import (
	"encoding/pem"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestECCAlgorithm(t *testing.T) {
	Convey("Given an ECCAlgorithm", t, func() {
		algo := &crypto.ECCAlgorithm{}

		Convey("It should generate a valid ECC key pair", func() {
			kp, err := algo.GenerateKeyPair()

			So(err, ShouldBeNil)
			So(kp, ShouldNotBeNil)

			eccKp, ok := kp.(*crypto.ECCKeyPair)
			So(ok, ShouldBeTrue)
			So(eccKp.Public, ShouldNotBeNil)
			So(eccKp.Private, ShouldNotBeNil)
		})

		Convey("ConstructKeyPair should rebuild the same key pair from serialized private key", func() {
			kp, _ := algo.GenerateKeyPair()
			_, priv, _ := kp.(*crypto.ECCKeyPair).Serialize()

			reconstructed, err := algo.ConstructKeyPair(priv)
			So(err, ShouldBeNil)
			So(reconstructed, ShouldNotBeNil)

			newKp := reconstructed.(*crypto.ECCKeyPair)
			So(newKp.Private.D.Cmp(kp.(*crypto.ECCKeyPair).Private.D), ShouldEqual, 0)
			So(newKp.Public.X.Cmp(kp.(*crypto.ECCKeyPair).Public.X), ShouldEqual, 0)
			So(newKp.Public.Y.Cmp(kp.(*crypto.ECCKeyPair).Public.Y), ShouldEqual, 0)
		})

		Convey("ConstructKeyPair should fail with invalid PEM", func() {
			reconstructed, err := algo.ConstructKeyPair([]byte("INVALID PEM DATA"))
			So(reconstructed, ShouldNotBeNil) // still returns *ECCKeyPair
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not a valid PEM")
		})

		Convey("Serialize and Deserialize should preserve key values", func() {
			kp, _ := algo.GenerateKeyPair()
			pub, priv, err := kp.(*crypto.ECCKeyPair).Serialize()

			So(err, ShouldBeNil)
			So(pub, ShouldNotBeEmpty)
			So(priv, ShouldNotBeEmpty)

			newKp := &crypto.ECCKeyPair{}
			err = newKp.Deserialize(priv)
			So(err, ShouldBeNil)

			// Verify both private and public keys match numerically
			So(newKp.Private.D.Cmp(kp.(*crypto.ECCKeyPair).Private.D), ShouldEqual, 0)
			So(newKp.Public.X.Cmp(kp.(*crypto.ECCKeyPair).Public.X), ShouldEqual, 0)
			So(newKp.Public.Y.Cmp(kp.(*crypto.ECCKeyPair).Public.Y), ShouldEqual, 0)
		})

		Convey("Deserialize should fail on invalid PEM", func() {
			kp := &crypto.ECCKeyPair{}
			err := kp.Deserialize([]byte("INVALID PEM DATA"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not a valid PEM")
		})

		Convey("Deserialize should fail on invalid private key bytes", func() {
			kp := &crypto.ECCKeyPair{}
			invalidPem := pem.EncodeToMemory(&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: []byte("INVALID KEY"),
			})
			err := kp.Deserialize(invalidPem)
			So(err, ShouldNotBeNil)
		})

		Convey("Sign and Verify should succeed on valid data", func() {
			kp, _ := algo.GenerateKeyPair()
			data := []byte("fiskaly ecc test")

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
			So(err.Error(), ShouldContainSubstring, "Verification failed")
		})

		Convey("Sign should fail on non-ECC private key", func() {
			_, err := algo.Sign("not-a-key", []byte("data"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not an ECC")
		})

		Convey("Verify should fail on non-ECC public key", func() {
			err := algo.Verify("not-a-key", []byte("data"), []byte("sig"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not a ECC")
		})
	})
}
