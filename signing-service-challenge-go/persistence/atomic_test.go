package persistence

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/testutil/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAtomicStorageLocking(t *testing.T) {
	Convey("Given an AtomicStorage wrapping a mocked Storage", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockStorage(ctrl)
		storage := NewAtomicStorage(mockDB)

		Convey("Lock and Unlock should block access to the same ID", func() {
			id := "device-1"
			start := make(chan struct{})
			done := make(chan struct{})

			go func() {
				storage.Lock(id)
				defer storage.Unlock(id)
				close(start)
				time.Sleep(100 * time.Millisecond)
				close(done)
			}()

			<-start // wait until first goroutine holds the lock

			blocked := make(chan struct{})
			go func() {
				storage.Lock(id)
				close(blocked)
				storage.Unlock(id)
			}()

			select {
			case <-done:
				select {
				case <-blocked:
					So(true, ShouldBeTrue) // second goroutine proceeded after unlock
				case <-time.After(50 * time.Millisecond):
					So(false, ShouldBeTrue) // blocked too long
				}
			case <-time.After(200 * time.Millisecond):
				So(false, ShouldBeTrue) // deadlock timeout
			}
		})

		Convey("Locks for different IDs should not block each other", func() {
			id1 := "device-a"
			id2 := "device-b"
			blocked := make(chan struct{})

			storage.Lock(id1)
			defer storage.Unlock(id1)

			go func() {
				storage.Lock(id2)
				close(blocked)
				storage.Unlock(id2)
			}()

			select {
			case <-blocked:
				So(true, ShouldBeTrue) // different IDs can lock concurrently
			case <-time.After(50 * time.Millisecond):
				So(false, ShouldBeTrue)
			}
		})
	})
}

func TestAtomicStorageDelegation(t *testing.T) {
	Convey("Given an AtomicStorage wrapping a mocked Storage", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockStorage(ctrl)
		storage := NewAtomicStorage(mockDB)

		Convey("Load delegates to base storage", func() {
			mockDB.EXPECT().Load("id1").Return(nil, nil).Times(1)
			storage.Load("id1")
		})

		Convey("Save delegates to base storage", func() {
			mockDB.EXPECT().Save("id2", gomock.Any()).Return(nil).Times(1)
			storage.Save("id2", nil)
		})

		Convey("List delegates to base storage", func() {
			mockDB.EXPECT().List().Return(nil).Times(1)
			storage.List()
		})
	})
}
