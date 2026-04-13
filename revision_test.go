package fortimgr

import (
	"context"
	"testing"
	"time"
)

func TestListADOMRevisions(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListADOMRevisions(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListADOMRevisions(context.Background(), "bad adom")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		// Fixture covers both int and string forms of created_time (FMG
		// returns it as a numeric string in the wild) and both locked/
		// unlocked entries.
		client := newTestClient(t, map[string]string{
			"/dvmdb/adom/root/revision": `[
				{
					"oid": 4441,
					"version": 1,
					"name": "initial-import",
					"desc": "Initial ADOM import from backup",
					"created_by": "admin",
					"created_time": 1700000000,
					"locked": 0
				},
				{
					"oid": 4500,
					"version": 42,
					"name": "policy-approval-snapshot",
					"desc": "",
					"created_by": "reviewer",
					"created_time": "1752464974",
					"locked": 1
				},
				{
					"oid": 4600,
					"version": 100,
					"name": "",
					"desc": "",
					"created_by": "",
					"created_time": 0,
					"locked": 0
				}
			]`,
		})

		revs, err := client.ListADOMRevisions(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(revs) != 3 {
			t.Fatalf("len = %d, want 3", len(revs))
		}

		// Entry 0: int created_time
		r0 := revs[0]
		if r0.Version != 1 {
			t.Errorf("Version = %d, want 1", r0.Version)
		}
		if r0.Name != "initial-import" {
			t.Errorf("Name = %q", r0.Name)
		}
		if r0.Desc != "Initial ADOM import from backup" {
			t.Errorf("Desc = %q", r0.Desc)
		}
		if r0.CreatedBy != "admin" {
			t.Errorf("CreatedBy = %q", r0.CreatedBy)
		}
		if !r0.CreatedAt.Equal(time.Unix(1700000000, 0).UTC()) {
			t.Errorf("CreatedAt = %v", r0.CreatedAt)
		}
		if r0.Locked {
			t.Errorf("Locked should be false")
		}

		// Entry 1: string created_time + locked
		r1 := revs[1]
		if r1.Version != 42 {
			t.Errorf("Version = %d", r1.Version)
		}
		if !r1.CreatedAt.Equal(time.Unix(1752464974, 0).UTC()) {
			t.Errorf("CreatedAt = %v, want unix 1752464974 (string-encoded)", r1.CreatedAt)
		}
		if !r1.Locked {
			t.Errorf("Locked should be true when locked=1")
		}

		// Entry 2: zero created_time → zero time.Time, no 1970
		r2 := revs[2]
		if !r2.CreatedAt.IsZero() {
			t.Errorf("CreatedAt should be zero for created_time=0, got %v", r2.CreatedAt)
		}
	})
}
