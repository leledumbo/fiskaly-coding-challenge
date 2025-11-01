package routes

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"net/http"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Health evaluates the health of the service and writes a standardized response.
func Health(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		common.WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	health := HealthResponse{
		Status:  "pass",
		Version: "v0",
	}

	common.WriteAPIResponse(response, http.StatusOK, health)
}

func init() {
	common.RegisterRoute("/api/v0/health", Health)
}
