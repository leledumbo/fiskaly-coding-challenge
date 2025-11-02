package routes

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"net/http"
)

type ListDevicesResponse struct {
	Devices []*domain.Device `json:"devices"`
}

// ListDevices lists all devices on the system, no filter
func ListDevices(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		common.WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	db := persistence.GetInstance()
	output := ListDevicesResponse{
		Devices: db.List(),
	}

	common.WriteAPIResponse(response, http.StatusOK, output)
}

func init() {
	common.RegisterRoute("/api/v0/list_devices", ListDevices)
}
