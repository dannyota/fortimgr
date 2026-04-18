package fortimgr

import (
	"context"
	"testing"
	"time"
)

func TestListWorkflowSessions(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListWorkflowSessions(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListWorkflowSessions(context.Background(), "bad adom")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		// Fixture covers:
		// - A fully approved session (state=3, all three user/time pairs set)
		// - A draft session (only create_user / create_time; submit/audit empty)
		// - An unknown state (state=99) that should pass through as raw int
		client := newTestClient(t, map[string]string{
			"/dvmdb/adom/root/workflow": `[
				{
					"oid": 100, "sessionid": 1,
					"name": "add firewall rule for app-a",
					"desc": "Workflow session",
					"create_user": "ops-alice", "create_time": "1700000000",
					"submit_user": "ops-alice", "submit_time": "1700000500",
					"audit_user":  "sec-bob",   "audit_time":  "1700001000",
					"revid": 42, "state": 3, "flags": 0
				},
				{
					"oid": 101, "sessionid": 2,
					"name": "draft — wip",
					"desc": "",
					"create_user": "ops-alice", "create_time": "1700500000",
					"submit_user": "", "submit_time": 0,
					"audit_user": "", "audit_time": 0,
					"revid": 0, "state": 0, "flags": 0
				},
				{
					"oid": 102, "sessionid": 3,
					"name": "rejected test",
					"desc": "",
					"create_user": "ops-alice", "create_time": "1700600000",
					"submit_user": "ops-alice", "submit_time": "1700600200",
					"audit_user":  "sec-carol", "audit_time":  "1700600400",
					"revid": 43, "state": 99, "flags": 0
				}
			]`,
		})

		sessions, err := client.ListWorkflowSessions(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(sessions) != 3 {
			t.Fatalf("len = %d, want 3", len(sessions))
		}

		// Approved session
		s0 := sessions[0]
		if s0.SessionID != 1 {
			t.Errorf("SessionID = %d", s0.SessionID)
		}
		if s0.Name != "add firewall rule for app-a" {
			t.Errorf("Name = %q", s0.Name)
		}
		if s0.State != "approved" {
			t.Errorf("State = %q, want approved", s0.State)
		}
		if s0.CreatedBy != "ops-alice" || s0.SubmittedBy != "ops-alice" || s0.AuditedBy != "sec-bob" {
			t.Errorf("user chain = %q/%q/%q", s0.CreatedBy, s0.SubmittedBy, s0.AuditedBy)
		}
		if !s0.CreatedAt.Equal(time.Unix(1700000000, 0).UTC()) {
			t.Errorf("CreatedAt = %v", s0.CreatedAt)
		}
		if !s0.SubmittedAt.Equal(time.Unix(1700000500, 0).UTC()) {
			t.Errorf("SubmittedAt = %v", s0.SubmittedAt)
		}
		if !s0.AuditedAt.Equal(time.Unix(1700001000, 0).UTC()) {
			t.Errorf("AuditedAt = %v", s0.AuditedAt)
		}
		if s0.RevisionID != 42 {
			t.Errorf("RevisionID = %d, want 42", s0.RevisionID)
		}

		// Draft session — zero timestamps, unmapped state (0) passed through
		s1 := sessions[1]
		if s1.Name != "draft — wip" {
			t.Errorf("Name = %q", s1.Name)
		}
		if !s1.SubmittedAt.IsZero() {
			t.Errorf("SubmittedAt should be zero for draft, got %v", s1.SubmittedAt)
		}
		if !s1.AuditedAt.IsZero() {
			t.Errorf("AuditedAt should be zero for draft")
		}
		if s1.SubmittedBy != "" {
			t.Errorf("SubmittedBy = %q, want empty", s1.SubmittedBy)
		}
		if s1.State != "0" {
			t.Errorf("State = %q, want raw passthrough \"0\" (unmapped)", s1.State)
		}
		if s1.RevisionID != 0 {
			t.Errorf("RevisionID = %d, want 0 (no revision for draft)", s1.RevisionID)
		}

		// Unknown state — passes through as raw int string
		s2 := sessions[2]
		if s2.State != "99" {
			t.Errorf("State = %q, want raw passthrough \"99\"", s2.State)
		}
	})
}

func TestListWorkflowLogs(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListWorkflowLogs(context.Background(), "root", 1)
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid args", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListWorkflowLogs(context.Background(), "bad adom", 1)
		if err == nil {
			t.Fatal("want ErrInvalidName for bad ADOM")
		}
		_, err = client.ListWorkflowLogs(context.Background(), "root", 0)
		if err == nil {
			t.Fatal("want ErrInvalidName for zero session")
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/dvmdb/adom/root/workflow/1/wflog": `[
				{"oid":5975,"seq":1,"sessionid":1,"action":5,"user":"ops-alice","time":"1752464088","desc":null,"flags":0},
				{"oid":5976,"seq":2,"sessionid":1,"action":6,"user":"sec-bob","time":1752464099,"desc":"approved","flags":1}
			]`,
		})

		logs, err := client.ListWorkflowLogs(context.Background(), "root", 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(logs) != 2 {
			t.Fatalf("len = %d, want 2", len(logs))
		}
		if logs[0].OID != 5975 || logs[0].Sequence != 1 || logs[0].SessionID != 1 || logs[0].Action != 5 {
			t.Errorf("logs[0] = %+v", logs[0])
		}
		if logs[0].User != "ops-alice" {
			t.Errorf("User = %q", logs[0].User)
		}
		if !logs[0].Timestamp.Equal(time.Unix(1752464088, 0).UTC()) {
			t.Errorf("Timestamp = %v", logs[0].Timestamp)
		}
		if logs[1].Description != "approved" || logs[1].Flags != 1 {
			t.Errorf("logs[1] = %+v", logs[1])
		}
	})
}
