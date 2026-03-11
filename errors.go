package fortimgr

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAuth           = errors.New("fortimgr: authentication failed")
	ErrPermission     = errors.New("fortimgr: no permission for resource")
	ErrCertificate    = errors.New("fortimgr: invalid TLS certificate")
	ErrNotLoggedIn    = errors.New("fortimgr: not logged in")
	ErrInvalidName    = errors.New("fortimgr: invalid ADOM or package name")
	ErrSessionExpired = errors.New("fortimgr: session expired")
)

// APIError represents a non-zero status code from FortiManager.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("fortimgr: API error %d: %s", e.Code, e.Message)
}

// isCertificateError checks if err is a TLS/x509 certificate error.
func isCertificateError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "x509:") ||
		strings.Contains(s, "certificate") ||
		strings.Contains(s, "tls:")
}
