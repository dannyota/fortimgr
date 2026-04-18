package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPackageInstallStatus(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListPackageInstallStatus(context.Background(), "root", "")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListPackageInstallStatus(context.Background(), "bad adom", "")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("invalid pkg", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListPackageInstallStatus(context.Background(), "root", "bad pkg")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("unfiltered — all packages", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/_package/status": `[
				{"dev": "fw-01", "pkg": "pkg-a", "vdom": "root", "status": "installed", "oid": 100},
				{"dev": "fw-01", "pkg": "pkg-b", "vdom": "dmz",  "status": "modified",  "oid": 101},
				{"dev": "fw-02", "pkg": "pkg-a", "vdom": "root", "status": "never",     "oid": 102}
			]`,
		})

		rows, err := client.ListPackageInstallStatus(context.Background(), "root", "")
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 3 {
			t.Fatalf("len = %d, want 3", len(rows))
		}
		// Spot-check one of each status value.
		seen := map[string]string{}
		for _, r := range rows {
			seen[r.Device+"/"+r.Package+"/"+r.VDOM] = r.Status
			if r.ADOM != "root" {
				t.Errorf("ADOM = %q, want root", r.ADOM)
			}
		}
		if seen["fw-01/pkg-a/root"] != "installed" {
			t.Errorf("fw-01/pkg-a/root = %q, want installed", seen["fw-01/pkg-a/root"])
		}
		if seen["fw-01/pkg-b/dmz"] != "modified" {
			t.Errorf("fw-01/pkg-b/dmz = %q, want modified", seen["fw-01/pkg-b/dmz"])
		}
		if seen["fw-02/pkg-a/root"] != "never" {
			t.Errorf("fw-02/pkg-a/root = %q, want never", seen["fw-02/pkg-a/root"])
		}
	})

	t.Run("filter by pkg — server-side filter", func(t *testing.T) {
		// Use a custom server so we can assert the outgoing params include
		// the filter clause (testhelper's forward handler doesn't inspect
		// filter; we want to verify it's actually sent).
		var lastBody []byte
		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "HTTP_CSRF_TOKEN", Value: "test-token", Path: "/"})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		})
		mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-CSRFToken") != "test-token" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			lastBody = readAll(t, r)
			// Return only pkg-b rows.
			_, _ = fmt.Fprintln(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[{"dev":"fw-01","pkg":"pkg-b","vdom":"dmz","status":"modified","oid":101}]}]}}`)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		c, err := NewClient(srv.URL, WithCredentials("u", "p"))
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Login(context.Background()); err != nil {
			t.Fatal(err)
		}

		rows, err := c.ListPackageInstallStatus(context.Background(), "root", "pkg-b")
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].Package != "pkg-b" || rows[0].Status != "modified" {
			t.Fatalf("rows = %+v", rows)
		}

		// Assert the sent body carried the filter clause.
		var sent struct {
			Method string `json:"method"`
			Params []struct {
				URL    string `json:"url"`
				Filter []any  `json:"filter"`
			} `json:"params"`
		}
		if err := json.Unmarshal(lastBody, &sent); err != nil {
			t.Fatalf("unmarshal sent body: %v", err)
		}
		if len(sent.Params) != 1 {
			t.Fatalf("params len = %d, want 1", len(sent.Params))
		}
		if sent.Params[0].URL != "/pm/config/adom/root/_package/status" {
			t.Errorf("url = %q", sent.Params[0].URL)
		}
		if len(sent.Params[0].Filter) != 3 {
			t.Fatalf("filter = %v, want [pkg, ==, pkg-b]", sent.Params[0].Filter)
		}
		if sent.Params[0].Filter[0] != "pkg" || sent.Params[0].Filter[1] != "==" || sent.Params[0].Filter[2] != "pkg-b" {
			t.Errorf("filter = %v, want [pkg, ==, pkg-b]", sent.Params[0].Filter)
		}
	})
}

// readAll is a small test helper that reads a request body.
func readAll(t *testing.T, r *http.Request) []byte {
	t.Helper()
	b := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			b = append(b, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return b
}
