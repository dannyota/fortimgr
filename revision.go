package fortimgr

import (
	"context"
	"fmt"
)

type apiADOMRevision struct {
	OID         int    `json:"oid"`
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	CreatedBy   string `json:"created_by"`
	CreatedTime any    `json:"created_time"` // int or numeric string
	Locked      any    `json:"locked"`       // int 0/1
}

// ListADOMRevisions returns the revision history for an ADOM — every
// snapshot FortiManager has on file for the ADOM, ordered by version.
// Each revision is created when a change is applied (via workflow
// submission, install preview, or manual revision) and is identified
// by an incrementing Version number. The Version field joins against
// WorkflowSession.RevisionID for per-change-request traceability.
//
// Typical result sizes range from a few hundred to several thousand
// entries on busy ADOMs. Pagination is applied transparently (1000
// rows per forward request by default); override with WithPageSize or
// observe progress with WithPageCallback.
func (c *Client) ListADOMRevisions(ctx context.Context, adom string, opts ...ListOption) ([]ADOMRevision, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/dvmdb/adom/%s/revision", adom)
	items, err := getPaged[apiADOMRevision](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	revs := make([]ADOMRevision, len(items))
	for i, r := range items {
		revs[i] = ADOMRevision{
			Version:   r.Version,
			Name:      r.Name,
			Desc:      r.Desc,
			CreatedBy: r.CreatedBy,
			CreatedAt: unixToTime(r.CreatedTime),
			Locked:    toString(r.Locked) == "1",
		}
	}
	return revs, nil
}
