package routes

import (
	"encoding/json"
	"errors"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/domain"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"io"
	"net/http"
)

var db persistence.Storage

type CreateSignatureDeviceRequest struct {
	DeviceID  string  `json:"device_id"`
	Algorithm string  `json:"algorithm"`
	Label     *string `json:"label,omitempty"`
	Update    *bool   `json:"update,omitempty"`
}

func (request *CreateSignatureDeviceRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateSignatureDeviceRequest // Avoid recursion
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(request),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if request.DeviceID == "" {
		return errors.New("Device ID is required")
	}
	if request.Algorithm == "" {
		return errors.New("Algorithm is required")
	}

	return nil
}

type CreateSignatureDeviceResponse struct {
	OK bool
}

// CreateSignatureDevice creates a signature device on the system using user selected algorihm, optionally labeling it for display
func CreateSignatureDevice(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		common.WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	body, _ := io.ReadAll(request.Body)
	defer request.Body.Close()
	var input CreateSignatureDeviceRequest
	if err := json.Unmarshal(body, &input); err != nil {
		common.WriteErrorResponse(response, http.StatusBadRequest, []string{
			err.Error(),
		})
		return
	}

	_, err := db.Load(input.DeviceID)
	if err == nil && (input.Update == nil || !*input.Update) {
		common.WriteErrorResponse(response, http.StatusBadRequest, []string{
			"Device with ID " + input.DeviceID + `already exists, if you want to update, supply "update":true in the request body`,
		})
		return
	}

	algo := crypto.GetAlgorithm(input.Algorithm)
	if algo == nil {
		common.WriteErrorResponse(response, http.StatusBadRequest, []string{
			"Algorithm " + input.Algorithm + " not available",
		})
		return
	}

	keyPair, err := algo.GenerateKeyPair()
	if err != nil {
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
	}

	_, serializedPrivateKey, err := keyPair.Serialize()
	if err != nil {
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
	}

	label := ""
	if input.Label != nil {
		label = *input.Label
	}
	device := domain.Device{
		ID:         input.DeviceID,
		Algorithm:  input.Algorithm,
		PrivateKey: serializedPrivateKey,
		Label:      label,
	}

	err = db.Save(device.ID, &device)
	if err != nil {
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
	}

	output := CreateSignatureDeviceResponse{
		OK: true,
	}
	common.WriteAPIResponse(response, http.StatusOK, output)
}

func init() {
	db = persistence.NewInMemoryDB()

	common.RegisterRoute("/api/v0/create_signature_device", CreateSignatureDevice)
}
