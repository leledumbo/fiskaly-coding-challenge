package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/routes"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/testutil/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

// Helper struct to unmarshal {"data": {...}}
type apiResponse struct {
	Data routes.ListDevicesResponse `json:"data"`
}

func TestListDevices(t *testing.T) {
	Convey("ListDevices endpoint", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockStorage(ctrl)
		persistence.SetInstance(mockDB)

		Convey("returns 405 if method is not GET", func() {
			req := httptest.NewRequest(http.MethodPost, "/signature-device", bytes.NewBuffer(nil))
			rec := httptest.NewRecorder()

			routes.ListDevices(rec, req)

			So(rec.Code, ShouldEqual, http.StatusMethodNotAllowed)

			var errResp common.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			So(err, ShouldBeNil)
			So(errResp.Errors, ShouldContain, http.StatusText(http.StatusMethodNotAllowed))
		})

		Convey("returns empty list if no devices", func() {
			mockDB.EXPECT().List().Return([]*domain.Device{})

			req := httptest.NewRequest(http.MethodGet, "/signature-device", nil)
			rec := httptest.NewRecorder()

			routes.ListDevices(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)

			var resp apiResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Devices, ShouldBeEmpty)
		})

		Convey("returns one device", func() {
			dev := &domain.Device{ID: "dev123", Algorithm: "RSA"}
			mockDB.EXPECT().List().Return([]*domain.Device{dev})

			req := httptest.NewRequest(http.MethodGet, "/signature-device", nil)
			rec := httptest.NewRecorder()

			routes.ListDevices(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)

			var resp apiResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Devices, ShouldHaveLength, 1)
			So(resp.Data.Devices[0].ID, ShouldEqual, "dev123")
			So(resp.Data.Devices[0].Algorithm, ShouldEqual, "RSA")
		})

		Convey("returns multiple devices", func() {
			devices := []*domain.Device{
				{ID: "dev1", Algorithm: "RSA"},
				{ID: "dev2", Algorithm: "ECC"},
			}
			mockDB.EXPECT().List().Return(devices)

			req := httptest.NewRequest(http.MethodGet, "/signature-device", nil)
			rec := httptest.NewRecorder()

			routes.ListDevices(rec, req)

			So(rec.Code, ShouldEqual, http.StatusOK)

			var resp apiResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Devices, ShouldHaveLength, 2)
			So(resp.Data.Devices[0].ID, ShouldEqual, "dev1")
			So(resp.Data.Devices[1].ID, ShouldEqual, "dev2")
		})
	})
}
