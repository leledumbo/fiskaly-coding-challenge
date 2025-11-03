package routes_test

import (
	"bytes"
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

type verifyAPIResponse struct {
	Data routes.VerifySignatureResponse `json:"data"`
}

func TestVerifySignature(t *testing.T) {
	Convey("VerifySignature endpoint", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockStorage(ctrl)
		persistence.SetInstance(mockDB)

		mockAlgo := mocks.NewMockAlgorithm(ctrl)
		mockKeyPair := mocks.NewMockKeyPair(ctrl)
		mockPubKey := "pub"

		Convey("returns 405 if method not POST", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v0/verify_signature", nil)
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusMethodNotAllowed)
			var errResp common.ErrorResponse
			json.Unmarshal(rec.Body.Bytes(), &errResp)
			So(errResp.Errors, ShouldContain, http.StatusText(http.StatusMethodNotAllowed))
		})

		Convey("returns 400 if JSON is invalid", func() {
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer([]byte("{invalid-json")))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "invalid character")
		})

		Convey("returns 400 if device_id missing", func() {
			body := []byte(`{"data":"payload","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Device ID is required")
		})

		Convey("returns 400 if data missing", func() {
			body := []byte(`{"device_id":"dev123","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Data is required")
		})

		Convey("returns 400 if signature missing", func() {
			body := []byte(`{"device_id":"dev123","data":"payload"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "Signature is required")
		})

		Convey("returns 404 if device not found", func() {
			mockDB.EXPECT().Load("dev123").Return(nil, errors.New("not found"))
			body := []byte(`{"device_id":"dev123","data":"payload","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusNotFound)
			So(rec.Body.String(), ShouldContainSubstring, "not found")
		})

		Convey("returns 500 if algorithm not available", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", nil)

			body := []byte(`{"device_id":"dev123","data":"payload","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "Something is wrong on our side")
		})

		Convey("returns 500 if ConstructKeyPair fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(nil, errors.New("keypair fail"))

			body := []byte(`{"device_id":"dev123","data":"payload","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, "Something is wrong on our side")
		})

		Convey("returns 400 if base64 decode fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mockKeyPair, nil)

			body := []byte(`{"device_id":"dev123","data":"payload","signature":"invalid-base64=="}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
			So(rec.Body.String(), ShouldContainSubstring, "illegal base64 data")
		})

		Convey("returns 200 with verified=false if Verify fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().PublicKey().Return(mockPubKey)
			mockAlgo.EXPECT().Verify(mockPubKey, []byte("payload"), []byte("sig")).Return(errors.New("verify fail"))

			body := []byte(`{"device_id":"dev123","data":"payload","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)
			var resp verifyAPIResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Verified, ShouldBeFalse)
			So(resp.Data.Reason, ShouldEqual, "verify fail")
		})

		Convey("returns 200 on success", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mockKeyPair, nil)
			mockKeyPair.EXPECT().PublicKey().Return(mockPubKey)
			mockAlgo.EXPECT().Verify(mockPubKey, []byte("payload"), []byte("sig")).Return(nil)

			body := []byte(`{"device_id":"dev123","data":"payload","signature":"c2ln"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify_signature", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)

			var resp verifyAPIResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Verified, ShouldBeTrue)
			So(resp.Data.Reason, ShouldBeEmpty)
		})
	})
}
