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

func TestList(t *testing.T) {
	Convey("Given an InMemoryDB instance", t, func() {
		db := NewInMemoryDB()

		Convey(`and 3 devices with different IDs`, func() {
			d1 := &domain.Device{ID: "a"}
			d2 := &domain.Device{ID: "b"}
			d3 := &domain.Device{ID: "c"}

			Convey("when they're all saved", func() {
				err1 := db.Save(d1.ID, d1)
				err2 := db.Save(d2.ID, d2)
				err3 := db.Save(d3.ID, d3)

				Convey("there should be no error", func() {
					So(err1, ShouldBeNil)
					So(err2, ShouldBeNil)
					So(err3, ShouldBeNil)
				})

				Convey("and when they're listed", func() {
					ds := db.List()

					Convey("there should be 3 of them", func() {
						So(len(ds), ShouldEqual, 3)
					})
				})
			})
		})
	})
}
