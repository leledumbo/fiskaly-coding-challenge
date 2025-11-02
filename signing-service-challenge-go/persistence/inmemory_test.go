package persistence

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	Convey("Given an InMemoryDB instance", t, func() {
		db := NewInMemoryDB()

		Convey(`and a device with ID "hello" and other fields default`, func() {
			d := &domain.Device{
				ID: "hello",
			}

			Convey("when the device is saved", func() {
				err := db.Save(d.ID, d)

				Convey("it returns no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("and when the same id is loaded", func() {
					d1, err := db.Load(d.ID)

					Convey("it returns no error", func() {
						So(err, ShouldBeNil)
					})

					Convey("and the same device will be returned", func() {
						So(d1, ShouldResemble, d)
					})
				})

				Convey("but when a different id is loaded", func() {
					_, err := db.Load(d.ID + ", world")

					Convey("it returns an error", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})
		})
	})
}
