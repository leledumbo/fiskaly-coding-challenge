package routes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"io"
	"net/http"
)

type VerifySignatureRequest struct {
	DeviceID  string `json:"device_id"`
	Data      string `json:"data"`
	Signature string `json:"signature"`
}

func (request *VerifySignatureRequest) UnmarshalJSON(data []byte) error {
	type Alias VerifySignatureRequest // Avoid recursion
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
	if request.Data == "" {
		return errors.New("Data is required")
	}
	if request.Signature == "" {
		return errors.New("Signature is required")
	}

	return nil
}

type VerifySignatureResponse struct {
	Verified bool   `json:"verified"`
	Reason   string `json:"reason,omitempty"`
}

func VerifySignature(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		common.WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	body, _ := io.ReadAll(request.Body)
	defer request.Body.Close()
	var input VerifySignatureRequest
	if err := json.Unmarshal(body, &input); err != nil {
		common.WriteErrorResponse(response, http.StatusBadRequest, []string{
			err.Error(),
		})
		return
	}

	db := persistence.GetInstance()
	device, err := db.Load(input.DeviceID)
	if err != nil {
		common.WriteErrorResponse(response, http.StatusNotFound, []string{
			err.Error(),
		})
		return
	}

	algo := crypto.GetAlgorithm(device.Algorithm)
	if algo == nil {
		// shouldn't happen, but possible to happen, so we still handle it,
		// however this is not user error and therefore, an internal server error
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
		return
	}

	kp, err := algo.ConstructKeyPair(device.PrivateKey)
	if err != nil {
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
		return
	}

	base64decodedSignature, err := base64.StdEncoding.DecodeString(input.Signature)
	if err != nil {
		common.WriteErrorResponse(response, http.StatusBadRequest, []string{
			err.Error(),
		})
		return
	}

	err = algo.Verify(kp.PublicKey(), []byte(input.Data), base64decodedSignature)
	output := VerifySignatureResponse{
		Verified: err == nil,
	}
	if err != nil {
		output.Reason = err.Error()
	}
	common.WriteAPIResponse(response, http.StatusOK, output)
}

func init() {
	common.RegisterRoute("/api/v0/verify_signature", VerifySignature)
}
