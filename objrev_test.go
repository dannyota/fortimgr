package fortimgr

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestListPolicyRevisions(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListPolicyRevisions(context.Background(), "root", "pkg1", 10)
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListPolicyRevisions(context.Background(), "bad adom", "pkg1", 10)
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("invalid pkg", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListPolicyRevisions(context.Background(), "root", "bad pkg", 10)
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/_objrev/pkg/test-pkg/firewall/policy/10": `[
				{
					"act": 1,
					"category": 181,
					"config": "{\"policyid\": 10, \"name\": \"allow-web\", \"action\": 1, \"srcaddr\": [\"net-10.0.0.0/24\"], \"dstaddr\": [\"all\"], \"schedule\": [\"always\"], \"service\": [\"HTTPS\"]}",
					"flags": 0,
					"key": "10",
					"note": "...",
					"oid": 5001,
					"pkg_oid": 3000,
					"timestamp": 1700000000,
					"user": "admin"
				},
				{
					"act": 3,
					"category": 181,
					"config": "{\"policyid\": 10, \"name\": \"allow-web\", \"action\": 1, \"srcaddr\": [\"net-10.0.0.0/24\"], \"dstaddr\": [\"all\"], \"schedule\": [\"biz-hours\"], \"service\": [\"HTTPS\"]}",
					"flags": 0,
					"key": "10",
					"note": "Set Schedule to:\nbiz-hours",
					"oid": 5001,
					"pkg_oid": 3000,
					"timestamp": 1700100000,
					"user": "operator"
				},
				{
					"act": 3,
					"category": 181,
					"config": "{\"policyid\": 10, \"name\": \"allow-web-extended\", \"action\": 1, \"srcaddr\": [\"net-10.0.0.0/24\", \"net-172.16.0.0/16\"], \"dstaddr\": [\"all\"], \"schedule\": [\"biz-hours\"], \"service\": [\"HTTPS\"]}",
					"flags": 0,
					"key": "10",
					"note": "Rename policy to allow-web-extended",
					"oid": 5001,
					"pkg_oid": 3000,
					"timestamp": "1700200000",
					"user": "operator"
				}
			]`,
		})

		revs, err := client.ListPolicyRevisions(context.Background(), "root", "test-pkg", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(revs) != 3 {
			t.Fatalf("len = %d, want 3", len(revs))
		}

		// Entry 0: initial creation (oldest)
		r0 := revs[0]
		if r0.Revision != 1 {
			t.Errorf("Revision = %d, want 1", r0.Revision)
		}
		if r0.Action != "add" {
			t.Errorf("Action = %q, want %q", r0.Action, "add")
		}
		if r0.Note != "..." {
			t.Errorf("Note = %q", r0.Note)
		}
		if r0.User != "admin" {
			t.Errorf("User = %q", r0.User)
		}
		if !r0.Timestamp.Equal(time.Unix(1700000000, 0).UTC()) {
			t.Errorf("Timestamp = %v", r0.Timestamp)
		}
		if r0.PolicyID != 10 {
			t.Errorf("PolicyID = %d, want 10", r0.PolicyID)
		}
		if r0.OID != 5001 {
			t.Errorf("OID = %d", r0.OID)
		}
		if !json.Valid(r0.Config) {
			t.Errorf("Config is not valid JSON")
		}

		// Entry 1: schedule modification
		r1 := revs[1]
		if r1.Revision != 2 {
			t.Errorf("Revision = %d, want 2", r1.Revision)
		}
		if r1.Action != "modify" {
			t.Errorf("Action = %q, want %q", r1.Action, "modify")
		}
		if r1.Note != "Set Schedule to:\nbiz-hours" {
			t.Errorf("Note = %q", r1.Note)
		}
		if r1.User != "operator" {
			t.Errorf("User = %q", r1.User)
		}

		// Entry 2: rename — timestamp as string (FMG inconsistency)
		r2 := revs[2]
		if r2.Revision != 3 {
			t.Errorf("Revision = %d, want 3", r2.Revision)
		}
		if r2.Action != "modify" {
			t.Errorf("Action = %q", r2.Action)
		}
		if r2.Note != "Rename policy to allow-web-extended" {
			t.Errorf("Note = %q", r2.Note)
		}
		if !r2.Timestamp.Equal(time.Unix(1700200000, 0).UTC()) {
			t.Errorf("Timestamp = %v, want unix 1700200000 (string-encoded)", r2.Timestamp)
		}
		// Verify config snapshot reflects the renamed policy
		var cfg map[string]any
		if err := json.Unmarshal(r2.Config, &cfg); err != nil {
			t.Fatalf("Config unmarshal: %v", err)
		}
		if name, _ := cfg["name"].(string); name != "allow-web-extended" {
			t.Errorf("Config name = %q, want %q", name, "allow-web-extended")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/_objrev/pkg/test-pkg/firewall/policy/20": `[
				{
					"act": 1,
					"category": 181,
					"config": "",
					"flags": 0,
					"key": "20",
					"note": "...",
					"oid": 5002,
					"pkg_oid": 3000,
					"timestamp": 1700000000,
					"user": "admin"
				}
			]`,
		})

		revs, err := client.ListPolicyRevisions(context.Background(), "root", "test-pkg", 20)
		if err != nil {
			t.Fatal(err)
		}
		if len(revs) != 1 {
			t.Fatalf("len = %d, want 1", len(revs))
		}
		if revs[0].Config != nil {
			t.Errorf("Config should be nil for empty config string, got %s", revs[0].Config)
		}
	})

	t.Run("unmapped action passes through", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/_objrev/pkg/test-pkg/firewall/policy/30": `[
				{
					"act": 99,
					"category": 181,
					"config": "{}",
					"flags": 0,
					"key": "30",
					"note": "unknown action",
					"oid": 5003,
					"pkg_oid": 3000,
					"timestamp": 1700000000,
					"user": "admin"
				}
			]`,
		})

		revs, err := client.ListPolicyRevisions(context.Background(), "root", "test-pkg", 30)
		if err != nil {
			t.Fatal(err)
		}
		if revs[0].Action != "99" {
			t.Errorf("Action = %q, want %q (raw passthrough)", revs[0].Action, "99")
		}
	})
}

func TestListPolicyRevisionCounts(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListPolicyRevisionCounts(context.Background(), "root", "pkg1")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListPolicyRevisionCounts(context.Background(), "bad adom", "pkg1")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("invalid pkg", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListPolicyRevisionCounts(context.Background(), "root", "bad pkg")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/_objrev/pkg/test-pkg/firewall/policy": `{
				"10": 5,
				"20": 1,
				"30": 42,
				"OID_0": 7
			}`,
		})

		counts, err := client.ListPolicyRevisionCounts(context.Background(), "root", "test-pkg")
		if err != nil {
			t.Fatal(err)
		}

		if len(counts) != 3 {
			t.Fatalf("len = %d, want 3 (OID_0 should be excluded)", len(counts))
		}
		if counts[10] != 5 {
			t.Errorf("counts[10] = %d, want 5", counts[10])
		}
		if counts[20] != 1 {
			t.Errorf("counts[20] = %d, want 1", counts[20])
		}
		if counts[30] != 42 {
			t.Errorf("counts[30] = %d, want 42", counts[30])
		}
		if _, ok := counts[0]; ok {
			t.Errorf("OID_0 should not be in the map")
		}
	})

	t.Run("empty result", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/_objrev/pkg/empty-pkg/firewall/policy": `{}`,
		})

		counts, err := client.ListPolicyRevisionCounts(context.Background(), "root", "empty-pkg")
		if err != nil {
			t.Fatal(err)
		}
		if len(counts) != 0 {
			t.Errorf("len = %d, want 0", len(counts))
		}
	})
}
