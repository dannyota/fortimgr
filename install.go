package fortimgr

import (
	"context"
	"fmt"
)

type apiPackageInstallStatus struct {
	Dev    string `json:"dev"`
	Pkg    string `json:"pkg"`
	VDOM   string `json:"vdom"`
	Status string `json:"status"`
	OID    int    `json:"oid"`
}

// ListPackageInstallStatus returns per-device install status for policy
// packages in an ADOM. If pkg is non-empty, results are filtered to that
// package via a server-side filter.
//
// Distinguishes assignment (device is on the scope list — see
// PolicyPackage.Scope) from actual installation (config has been pushed
// and is running on the FortiGate). Use Status == "installed" to verify
// a policy is actually enforcing.
//
// The underlying /pm/config/adom/{adom}/_package/status endpoint does not
// expose revision numbers, install time, or modify state. Callers that
// need that richer view should join against ADOM revision history
// returned by ListADOMRevisions.
//
// Pagination: this method does NOT accept ListOption arguments because
// the underlying /pm/config/adom/{adom}/_package/status endpoint ignores
// the range parameter — FortiManager returns the full dataset in a single
// response regardless of any range or offset the client specifies. This
// is by FortiManager design. Every other List* method in the SDK supports
// WithPageSize / WithPageCallback; this one is a documented exception.
func (c *Client) ListPackageInstallStatus(ctx context.Context, adom, pkg string) ([]PackageInstallStatus, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}
	if pkg != "" && !validName(pkg) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, pkg)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/_package/status", adom)
	var items []apiPackageInstallStatus
	var err error
	if pkg == "" {
		items, err = get[apiPackageInstallStatus](ctx, c, apiURL)
	} else {
		items, err = getExtra[apiPackageInstallStatus](ctx, c, apiURL, map[string]any{
			"filter": []any{"pkg", "==", pkg},
		})
	}
	if err != nil {
		return nil, err
	}

	result := make([]PackageInstallStatus, len(items))
	for i, it := range items {
		result[i] = PackageInstallStatus{
			ADOM:    adom,
			Package: it.Pkg,
			Device:  it.Dev,
			VDOM:    it.VDOM,
			Status:  it.Status,
		}
	}
	return result, nil
}
