package fortimgr

import (
	"encoding/json"
	"errors"
	"fmt"
)

// flatUIResponse is the JSON envelope from /cgi-bin/module/forward.
type flatUIResponse struct {
	Code int `json:"code"`
	Data struct {
		Result []flatUIResult `json:"result"`
	} `json:"data"`
}

type flatUIResult struct {
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	Data json.RawMessage `json:"data"`
}

// checkResponse validates the FlatUI response envelope.
// Returns the data payload from result[0] on success.
func checkResponse(resp *flatUIResponse) (json.RawMessage, error) {
	if isSessionExpired(resp) {
		return nil, ErrSessionExpired
	}

	if resp.Code != 0 {
		return nil, &APIError{Code: resp.Code, Message: "transport error"}
	}

	if len(resp.Data.Result) == 0 {
		return nil, errors.New("fortimgr: empty response")
	}

	status := resp.Data.Result[0].Status
	if status.Code == -11 {
		return nil, fmt.Errorf("%w: %s", ErrPermission, status.Message)
	}
	if status.Code != 0 {
		return nil, &APIError{Code: status.Code, Message: status.Message}
	}

	return resp.Data.Result[0].Data, nil
}

// isSessionExpired checks if the response indicates a session timeout.
// FortiManager signals this with transport code -6 or result status code -6.
func isSessionExpired(resp *flatUIResponse) bool {
	if resp.Code == -6 {
		return true
	}
	if len(resp.Data.Result) > 0 && resp.Data.Result[0].Status.Code == -6 {
		return true
	}
	return false
}
