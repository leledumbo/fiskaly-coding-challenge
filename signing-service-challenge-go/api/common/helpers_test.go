package common_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCommonHelpers(t *testing.T) {
	Convey("Mux() and RegisterRoute()", t, func() {
		Convey("should return a singleton instance", func() {
			mux1 := common.Mux()
			mux2 := common.Mux()
			So(mux1, ShouldNotBeNil)
			So(mux2, ShouldNotBeNil)
			So(mux1, ShouldEqual, mux2)
		})

		Convey("should register and serve a route", func() {
			common.RegisterRoute("/ping", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("pong"))
			})
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rec := httptest.NewRecorder()
			common.Mux().ServeHTTP(rec, req)
			So(rec.Code, ShouldEqual, http.StatusOK)
			So(rec.Body.String(), ShouldEqual, "pong")
		})
	})

	Convey("WriteInternalError()", t, func() {
		rec := httptest.NewRecorder()
		common.WriteInternalError(rec)
		So(rec.Code, ShouldEqual, http.StatusInternalServerError)
		So(rec.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusInternalServerError))
	})

	Convey("WriteErrorResponse()", t, func() {
		Convey("should write proper JSON error response", func() {
			rec := httptest.NewRecorder()
			common.WriteErrorResponse(rec, http.StatusBadRequest, []string{"invalid input"})

			So(rec.Code, ShouldEqual, http.StatusBadRequest)

			var resp common.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Errors, ShouldContain, "invalid input")
		})
	})

	Convey("WriteAPIResponse()", t, func() {
		Convey("should write proper structured response", func() {
			rec := httptest.NewRecorder()
			data := map[string]string{"hello": "world"}
			common.WriteAPIResponse(rec, http.StatusOK, data)

			So(rec.Code, ShouldEqual, http.StatusOK)

			var resp common.Response
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.(map[string]interface{})["hello"], ShouldEqual, "world")
		})

		Convey("should fall back to internal error if JSON marshal fails", func() {
			rec := httptest.NewRecorder()
			common.WriteAPIResponse(rec, http.StatusOK, make(chan int))
			So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			So(rec.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusInternalServerError))
		})
	})

	Convey("ParseJSONRequestBody()", t, func() {
		type testStruct struct {
			Name string `json:"name"`
		}

		Convey("should parse valid JSON correctly", func() {
			body := io.NopCloser(bytes.NewBufferString(`{"name":"Mario"}`))
			var out testStruct
			err := common.ParseJSONRequestBody(body, &out)
			So(err, ShouldBeNil)
			So(out.Name, ShouldEqual, "Mario")
		})

		Convey("should return error for invalid JSON", func() {
			body := io.NopCloser(bytes.NewBufferString(`{"name":`))
			var out testStruct
			err := common.ParseJSONRequestBody(body, &out)
			So(err, ShouldNotBeNil)
		})

		Convey("should close the body after reading", func() {
			closed := false
			rc := &mockReadCloser{
				reader: bytes.NewBufferString(`{"name":"test"}`),
				onClose: func() {
					closed = true
				},
			}
			var out testStruct
			_ = common.ParseJSONRequestBody(rc, &out)
			So(closed, ShouldBeTrue)
		})
	})
}

// mockReadCloser helps test body closing behavior
type mockReadCloser struct {
	reader  *bytes.Buffer
	onClose func()
	once    sync.Once
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.once.Do(m.onClose)
	return nil
}
