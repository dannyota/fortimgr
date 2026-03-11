package fortimgr

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCheckResponse(t *testing.T) {
	t.Run("session expired transport", func(t *testing.T) {
		resp := &flatUIResponse{Code: -6}
		_, err := checkResponse(resp)
		if !errors.Is(err, ErrSessionExpired) {
			t.Errorf("err = %v, want ErrSessionExpired", err)
		}
	})

	t.Run("session expired result", func(t *testing.T) {
		resp := &flatUIResponse{}
		resp.Data.Result = []flatUIResult{{}}
		resp.Data.Result[0].Status.Code = -6
		resp.Data.Result[0].Status.Message = "Session expired"
		_, err := checkResponse(resp)
		if !errors.Is(err, ErrSessionExpired) {
			t.Errorf("err = %v, want ErrSessionExpired", err)
		}
	})

	t.Run("transport error", func(t *testing.T) {
		resp := &flatUIResponse{Code: -1}
		_, err := checkResponse(resp)
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.Code != -1 {
			t.Errorf("code = %d, want -1", apiErr.Code)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		resp := &flatUIResponse{}
		resp.Data.Result = nil
		_, err := checkResponse(resp)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("no permission", func(t *testing.T) {
		resp := &flatUIResponse{}
		resp.Data.Result = []flatUIResult{{}}
		resp.Data.Result[0].Status.Code = -11
		resp.Data.Result[0].Status.Message = "No permission"

		_, err := checkResponse(resp)
		if !errors.Is(err, ErrPermission) {
			t.Errorf("err = %v, want ErrPermission", err)
		}
	})

	t.Run("api error", func(t *testing.T) {
		resp := &flatUIResponse{}
		resp.Data.Result = []flatUIResult{{}}
		resp.Data.Result[0].Status.Code = -2
		resp.Data.Result[0].Status.Message = "invalid params"

		_, err := checkResponse(resp)
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.Code != -2 {
			t.Errorf("code = %d, want -2", apiErr.Code)
		}
		if apiErr.Message != "invalid params" {
			t.Errorf("message = %q", apiErr.Message)
		}
	})

	t.Run("success", func(t *testing.T) {
		resp := &flatUIResponse{}
		resp.Data.Result = []flatUIResult{{
			Data: json.RawMessage(`[{"name":"test"}]`),
		}}

		data, err := checkResponse(resp)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `[{"name":"test"}]` {
			t.Errorf("data = %s", data)
		}
	})
}

func TestAPIError(t *testing.T) {
	err := &APIError{Code: -11, Message: "forbidden"}
	got := err.Error()
	want := "fortimgr: API error -11: forbidden"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestIsCertificateError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"x509", errors.New("x509: certificate signed by unknown authority"), true},
		{"cert", errors.New("certificate is not valid"), true},
		{"tls", errors.New("tls: handshake failure"), true},
		{"other", errors.New("connection refused"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCertificateError(tt.err)
			if got != tt.want {
				t.Errorf("isCertificateError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
