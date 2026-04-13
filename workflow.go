package fortimgr

import (
	"context"
	"fmt"
)

type apiWorkflowSession struct {
	OID        int    `json:"oid"`
	SessionID  int    `json:"sessionid"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	CreateUser string `json:"create_user"`
	CreateTime any    `json:"create_time"` // numeric string or int
	SubmitUser string `json:"submit_user"`
	SubmitTime any    `json:"submit_time"`
	AuditUser  string `json:"audit_user"`
	AuditTime  any    `json:"audit_time"`
	RevID      int    `json:"revid"`
	State      any    `json:"state"`
	Flags      int    `json:"flags"`
}

// ListWorkflowSessions returns the workflow audit trail for an ADOM —
// every change request FortiManager has recorded, along with its
// creator, submitter, approver, and resulting revision ID. This is the
// primary audit-trail endpoint for compliance review: it answers "who
// asked for change X, who approved it, and when".
//
// Sessions with both SubmittedAt and AuditedAt populated have completed
// the approval chain. Sessions with only CreatedAt set are drafts that
// have not been submitted. The State field is best-effort because
// FortiManager's documentation does not enumerate all state values —
// empirically, sessions in the approval-complete state report
// State == "approved"; unknown ints pass through unchanged.
//
// The RevisionID field joins against ADOMRevision.Version to locate
// the snapshot the session produced.
//
// Pagination is applied transparently (1000 rows per forward request
// by default); override with WithPageSize or observe progress with
// WithPageCallback.
func (c *Client) ListWorkflowSessions(ctx context.Context, adom string, opts ...ListOption) ([]WorkflowSession, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/dvmdb/adom/%s/workflow", adom)
	items, err := getPaged[apiWorkflowSession](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	sessions := make([]WorkflowSession, len(items))
	for i, w := range items {
		sessions[i] = WorkflowSession{
			SessionID:   w.SessionID,
			Name:        w.Name,
			Description: w.Desc,
			State:       mapEnum(toString(w.State), workflowStates),
			CreatedBy:   w.CreateUser,
			CreatedAt:   unixToTime(w.CreateTime),
			SubmittedBy: w.SubmitUser,
			SubmittedAt: unixToTime(w.SubmitTime),
			AuditedBy:   w.AuditUser,
			AuditedAt:   unixToTime(w.AuditTime),
			RevisionID:  w.RevID,
		}
	}
	return sessions, nil
}
