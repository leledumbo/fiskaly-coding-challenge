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
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/routes"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/testutil/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestVerifySignature(t *testing.T) {
	Convey("VerifySignature endpoint", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockStorage(ctrl)
		persistence.SetInstance(mockDB)

		Convey("returns 405 if method is not POST", func() {
			req := httptest.NewRequest(http.MethodGet, "/verify", nil)
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusMethodNotAllowed)
		})

		Convey("returns 400 if request body is missing required fields", func() {
			body := `{"device_id": "dev123"}`
			req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("returns 404 if device not found", func() {
			mockDB.EXPECT().Load("dev123").Return(nil, errors.New("device not found"))

			body := `{"device_id": "dev123", "data": "abc", "signature": "c2ln"}`
			req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("returns 400 if signature is invalid base64", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)

			body := `{"device_id": "dev123", "data": "abc", "signature": "!!!invalid!!!"}`
			req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("returns 200 verified=true when signature is valid", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)

			mockAlgo := mocks.NewMockAlgorithm(ctrl)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mocks.NewMockKeyPair(ctrl), nil)
			mockAlgo.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			crypto.RegisterAlgorithm("RSA", mockAlgo)

			sig := base64.StdEncoding.EncodeToString([]byte("ok"))
			body := `{"device_id": "dev123", "data": "abc", "signature": "` + sig + `"}`
			req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)
			var resp common.Response
			json.Unmarshal(rec.Body.Bytes(), &resp)
			data, _ := json.Marshal(resp.Data)
			var out routes.VerifySignatureResponse
			json.Unmarshal(data, &out)
			So(out.Verified, ShouldBeTrue)
		})

		Convey("returns 200 verified=false when verification fails", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().Load("dev123").Return(dev, nil)

			mockAlgo := mocks.NewMockAlgorithm(ctrl)
			mockAlgo.EXPECT().ConstructKeyPair(gomock.Any()).Return(mocks.NewMockKeyPair(ctrl), nil)
			mockAlgo.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("bad signature"))
			crypto.RegisterAlgorithm("RSA", mockAlgo)

			sig := base64.StdEncoding.EncodeToString([]byte("ok"))
			body := `{"device_id": "dev123", "data": "abc", "signature": "` + sig + `"}`
			req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(body))
			rec := httptest.NewRecorder()

			routes.VerifySignature(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)
			var resp common.Response
			json.Unmarshal(rec.Body.Bytes(), &resp)
			data, _ := json.Marshal(resp.Data)
			var out routes.VerifySignatureResponse
			json.Unmarshal(data, &out)
			So(out.Verified, ShouldBeFalse)
			So(out.Reason, ShouldEqual, "bad signature")
		})
	})
}
