package routes_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/routes"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/testutil/mocks"
)

type signAPIResponse struct {
	Data routes.SignTransactionResponse `json:"data"`
}

func TestSignTransaction(t *testing.T) {
	Convey("SignTransaction endpoint", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockStorage(ctrl)
		persistence.SetInstance(mockDB)

		mockAlgo := mocks.NewMockAlgorithm(ctrl)
		mockKeyPair := mocks.NewMockKeyPair(ctrl)

		Convey("returns 405 if method not POST", func() {
			req := httptest.NewRequest(http.MethodGet, "/sign", nil)
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusMethodNotAllowed)
			var errResp common.ErrorResponse
			json.Unmarshal(rec.Body.Bytes(), &errResp)
			So(errResp.Errors, ShouldContain, http.StatusText(http.StatusMethodNotAllowed))
		})

		Convey("returns 400 if JSON is invalid", func() {
			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer([]byte("{invalid-json")))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "invalid character")
		})

		Convey("returns 400 if device_id missing", func() {
			body := []byte(`{"data":"payload"}`)
			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Device ID is required")
		})

		Convey("returns 400 if data missing", func() {
			body := []byte(`{"device_id":"dev123"}`)
			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Data is required")
		})

		Convey("returns 404 if device not found", func() {
			mockDB.EXPECT().Load("dev123").Return(nil, errors.New("not found"))
			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer([]byte(`{"device_id":"dev123","data":"payload"}`)))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusNotFound)
			So(rec.Body.String(), ShouldContainSubstring, "not found")
		})

		Convey("returns 500 if algorithm not available", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", nil)

			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer([]byte(`{"device_id":"dev123","data":"payload"}`)))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "Something is wrong on our side")
		})

		Convey("returns 500 if ConstructKeyPair fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(nil, errors.New("keypair fail"))

			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer([]byte(`{"device_id":"dev123","data":"payload"}`)))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "Something is wrong on our side")
		})

		Convey("returns 500 if Sign fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().PrivateKey().Return("priv")
			mockAlgo.EXPECT().Sign("priv", gomock.Any()).Return(nil, errors.New("sign fail"))

			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer([]byte(`{"device_id":"dev123","data":"payload"}`)))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "Something is wrong on our side")
		})

		Convey("returns 500 if Save fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			mockDB.EXPECT().Save("dev123", gomock.Any()).Return(errors.New("save fail"))

			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().PrivateKey().Return("priv")
			mockAlgo.EXPECT().Sign("priv", gomock.Any()).Return([]byte("sig"), nil)

			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer([]byte(`{"device_id":"dev123","data":"payload"}`)))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "Something is wrong on our side")
		})

		Convey("returns 200 on success", func() {
			dev := &domain.Device{
				ID:               "dev123",
				Algorithm:        "RSA",
				SignatureCounter: 0,
				LastSignature:    "",
			}

			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			mockDB.EXPECT().Save("dev123", gomock.Any()).Return(nil)

			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().PrivateKey().Return("priv")
			mockAlgo.EXPECT().Sign("priv", gomock.Any()).Return([]byte("mysignature"), nil)

			body := []byte(`{"device_id":"dev123","data":"payload"}`)
			req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.SignTransaction(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)

			var resp signAPIResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Signature, ShouldEqual, base64.StdEncoding.EncodeToString([]byte("mysignature")))
			So(resp.Data.SignedData, ShouldContainSubstring, "payload")
		})
	})
}
