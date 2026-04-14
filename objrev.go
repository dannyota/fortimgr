package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type apiPolicyRevision struct {
	Act       any    `json:"act"`
	Category  int    `json:"category"`
	Config    string `json:"config"`
	Flags     int    `json:"flags"`
	Key       string `json:"key"`
	Note      string `json:"note"`
	OID       int    `json:"oid"`
	PkgOID    int    `json:"pkg_oid"`
	Timestamp any    `json:"timestamp"`
	User      string `json:"user"`
}

// ListPolicyRevisions returns the per-policy revision history for a
// single firewall policy within a package — every field-level change
// FortiManager has recorded, ordered oldest-first. Each revision
// includes who made the change, when, what action was taken, a
// human-readable change note describing the modification, and the full
// policy configuration snapshot at that point in time (Config).
//
// The underlying /pm/config/adom/{adom}/_objrev/pkg/{pkg}/firewall/policy/{id}
// endpoint returns all revisions in a single response; pagination does
// not apply.
func (c *Client) ListPolicyRevisions(ctx context.Context, adom, pkg string, policyID int) ([]PolicyRevision, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}
	if !validName(pkg) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, pkg)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/_objrev/pkg/%s/firewall/policy/%d", adom, pkg, policyID)
	items, err := get[apiPolicyRevision](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	revs := make([]PolicyRevision, len(items))
	for i, r := range items {
		var config json.RawMessage
		if r.Config != "" {
			config = json.RawMessage(r.Config)
		}

		pid, _ := strconv.Atoi(r.Key)

		revs[i] = PolicyRevision{
			Revision:  i + 1,
			Action:    mapEnum(toString(r.Act), objrevActions),
			Note:      r.Note,
			User:      r.User,
			Timestamp: unixToTime(r.Timestamp),
			PolicyID:  pid,
			OID:       r.OID,
			Config:    config,
		}
	}
	return revs, nil
}

// ListPolicyRevisionCounts returns the number of revisions recorded for
// each firewall policy in a package. The result maps policy ID to
// revision count. Policies with no recorded revisions are not included.
//
// This is useful for identifying heavily-modified policies (potential
// audit targets) without fetching the full revision history of every
// policy.
func (c *Client) ListPolicyRevisionCounts(ctx context.Context, adom, pkg string) (map[int]int, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}
	if !validName(pkg) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, pkg)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/_objrev/pkg/%s/firewall/policy", adom, pkg)
	data, err := c.forwardExtra(ctx, apiURL, map[string]any{"option": "count"})
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal revision counts: %w", err)
	}

	counts := make(map[int]int, len(raw))
	for k, v := range raw {
		id, err := strconv.Atoi(k)
		if err != nil {
			continue // skip non-numeric keys like "OID_0"
		}
		if n, ok := v.(float64); ok {
			counts[id] = int(n)
		}
	}
	return counts, nil
}
