package routes_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/routes"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/testutil/mocks"
)

func TestCreateSignatureDevice(t *testing.T) {
	Convey("CreateSignatureDevice handler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAlgo := mocks.NewMockAlgorithm(ctrl)
		mockKeyPair := mocks.NewMockKeyPair(ctrl)
		mockDB := mocks.NewMockStorage(ctrl)

		// Register the mocked algorithm
		crypto.RegisterAlgorithm("RSA", mockAlgo)
		persistence.SetInstance(mockDB)

		Convey("returns 405 for non-POST method", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v0/create_signature_device", nil)
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusMethodNotAllowed)
			So(rec.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusMethodNotAllowed))
		})

		Convey("returns 400 if device already exists and update=false", func() {
			mockDB.EXPECT().Load("dev123").Return(&domain.Device{}, nil)

			body := []byte(`{"device_id":"dev123","algorithm":"RSA"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "already exists")
		})

		Convey("returns 400 if algorithm not available", func() {
			mockDB.EXPECT().Load("devX").Return(nil, errors.New("Device with id devOK not found"))
			// No registration for "Nonexistent" algorithm
			body := []byte(`{"device_id":"devX","algorithm":"Nonexistent"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "not available")
		})

		Convey("returns 400 if device_id given but algorithm missing", func() {
			body := []byte(`{"device_id":"devOnly"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Algorithm is required")
		})

		Convey("returns 400 if algorithm given but device_id missing", func() {
			body := []byte(`{"algorithm":"RSA"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Device ID is required")
		})

		Convey("returns 200 on successful creation", func() {
			mockDB.EXPECT().Load("devOK").Return(nil, errors.New("Device with id devOK not found"))
			mockAlgo.EXPECT().GenerateKeyPair().Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().Serialize().Return([]byte("pub"), []byte("priv"), nil)
			mockDB.EXPECT().Save("devOK", gomock.Any()).Return(nil)

			body := []byte(`{"device_id":"devOK","algorithm":"RSA"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)
			So(rec.Body.String(), ShouldContainSubstring, `"data"`)
		})

		Convey("returns 200 on successful creation with label", func() {
			mockDB.EXPECT().Load("devOK").Return(nil, errors.New("Device with id devOK not found"))
			mockAlgo.EXPECT().GenerateKeyPair().Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().Serialize().Return([]byte("pub"), []byte("priv"), nil)
			mockDB.EXPECT().Save("devOK", gomock.Any()).Return(nil)

			body := []byte(`{"device_id":"devOK","algorithm":"RSA","label":"XXX"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)
			So(rec.Body.String(), ShouldContainSubstring, `"data"`)
		})

		Convey("returns 500 if GenerateKeyPair fails", func() {
			mockDB.EXPECT().Load("devFail").Return(nil, errors.New("Device with id devOK not found"))
			mockAlgo.EXPECT().GenerateKeyPair().Return(nil, errors.New("Failed to generate key pair"))

			body := []byte(`{"device_id":"devFail","algorithm":"RSA"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "please try again")
		})

		Convey("returns 500 if keyPair.Serialize fails", func() {
			mockDB.EXPECT().Load("devSerializeErr").Return(nil, errors.New("Device with id devOK not found"))
			mockAlgo.EXPECT().GenerateKeyPair().Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().Serialize().Return(nil, nil, errors.New("serialize fail"))
			// Save should NOT be called because Serialize fails before Save; no Save expectation

			body := []byte(`{"device_id":"devSerializeErr","algorithm":"RSA"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "please try again")
		})

		Convey("returns 500 if db.Save fails", func() {
			mockDB.EXPECT().Load("devSaveErr").Return(nil, errors.New("Device with id devOK not found"))
			mockAlgo.EXPECT().GenerateKeyPair().Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().Serialize().Return([]byte("pub"), []byte("priv"), nil)
			mockDB.EXPECT().Save("devSaveErr", gomock.Any()).Return(errors.New("disk full"))

			body := []byte(`{"device_id":"devSaveErr","algorithm":"RSA"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/create_signature_device", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.CreateSignatureDevice(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "please try again")
		})
	})
}
