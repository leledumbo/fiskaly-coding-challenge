package common

import (
	"encoding/json"
	"io"
	"net/http"
)

// Actual *http.ServeMux instance, singleton, inaccessible outside package
var muxInstance *http.ServeMux

// Mux returns a *http.ServeMux instance lazily, acting as a singleton constructor
func Mux() *http.ServeMux {
	if muxInstance == nil {
		muxInstance = http.NewServeMux()
	}

	return muxInstance
}

// RegisterRoute registers @route to the HTTP handler
func RegisterRoute(route string, handler http.HandlerFunc) {
	Mux().Handle(route, handler)
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data any) {
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}

// ParseJSONRequestBody parses request body for an expected JSON structure, pass pointer to it as the second parameter
func ParseJSONRequestBody(requestBody io.ReadCloser, input interface{}) error {
	body, _ := io.ReadAll(requestBody)
	defer requestBody.Close()

	return json.Unmarshal(body, input)
}
