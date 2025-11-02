package routes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/crypto"
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/persistence"
	"io"
	"net/http"
)

type SignTransactionRequest struct {
	DeviceID string `json:"device_id"`
	Data     string `json:"data"`
}

func (request *SignTransactionRequest) UnmarshalJSON(data []byte) error {
	type Alias SignTransactionRequest // Avoid recursion
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

	return nil
}

type SignTransactionResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

func SignTransaction(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		common.WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	body, _ := io.ReadAll(request.Body)
	defer request.Body.Close()
	var input SignTransactionRequest
	if err := json.Unmarshal(body, &input); err != nil {
		common.WriteErrorResponse(response, http.StatusBadRequest, []string{
			err.Error(),
		})
		return
	}

	db := persistence.GetInstance()
	as := persistence.NewAtomicStorage(db)
	as.Lock(input.DeviceID)
	defer as.Unlock(input.DeviceID)

	device, err := as.Load(input.DeviceID)
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

	data := fmt.Sprintf("%d_%s_%s", device.SignatureCounter, input.Data, device.LastSignature)
	signature, err := algo.Sign(kp.PrivateKey(), []byte(data))
	if err != nil {
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
		return
	}

	base64encodedSignature := base64.StdEncoding.EncodeToString([]byte(signature))
	output := SignTransactionResponse{
		Signature:  base64encodedSignature,
		SignedData: data,
	}

	device.SignatureCounter++
	device.LastSignature = base64encodedSignature

	err = as.Save(device.ID, device)
	if err != nil {
		common.WriteErrorResponse(response, http.StatusInternalServerError, []string{
			"Something is wrong on our side, please try again in a few moments, our development team has been notified",
		})
		// log err to system log, email, prometheus, whatever, skipped for brevity
		return
	}

	common.WriteAPIResponse(response, http.StatusOK, output)
}

func init() {
	common.RegisterRoute("/api/v0/sign_transaction", SignTransaction)
}
