package common

// Response is the generic API response container.
type Response struct {
	Data any `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}
