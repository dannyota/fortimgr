package fortimgr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestPagination covers the generic getPaged helper via the ListAddresses
// pilot. Subtests using newTestClient exercise the common termination
// paths with small fixtures and small page sizes. The last two subtests
// use a custom httptest server to cover behaviors that need request-body
// inspection (extras merging) or stateful per-call responses (session
// expiry mid-pagination).
func TestPagination(t *testing.T) {
	addrURL := "/pm/config/adom/root/obj/firewall/address"

	// Helper: 5-row synthetic fixture "a" … "e".
	fx5 := `[
		{"name":"a"},{"name":"b"},{"name":"c"},{"name":"d"},{"name":"e"}
	]`

	t.Run("multi_page_assembly", func(t *testing.T) {
		// 5 rows, page size 2 → 3 pages (2+2+1). Terminates on
		// under-full final page (1 < 2).
		client := newTestClient(t, map[string]string{addrURL: fx5})
		addrs, err := client.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 5 {
			t.Fatalf("len = %d, want 5", len(addrs))
		}
		want := []string{"a", "b", "c", "d", "e"}
		for i, w := range want {
			if addrs[i].Name != w {
				t.Errorf("addrs[%d] = %q, want %q", i, addrs[i].Name, w)
			}
		}
	})

	t.Run("exact_page_size_boundary", func(t *testing.T) {
		// 4 rows, page size 2 → 3 pages (2+2+0). Needs the empty
		// follow-up page to terminate after an exact-boundary fetch.
		fx4 := `[{"name":"a"},{"name":"b"},{"name":"c"},{"name":"d"}]`
		client := newTestClient(t, map[string]string{addrURL: fx4})
		addrs, err := client.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 4 {
			t.Fatalf("len = %d, want 4", len(addrs))
		}
	})

	t.Run("single_page_boundary", func(t *testing.T) {
		// 2 rows, page size 2 → 2 pages (2+0). Page 1 is exactly full;
		// page 2 is empty and terminates the loop.
		fx2 := `[{"name":"a"},{"name":"b"}]`
		client := newTestClient(t, map[string]string{addrURL: fx2})
		addrs, err := client.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 2 {
			t.Fatalf("len = %d", len(addrs))
		}
	})

	t.Run("under_full_first_page", func(t *testing.T) {
		// 3 rows, page size 5 → 1 page. First-page short-circuit.
		fx3 := `[{"name":"a"},{"name":"b"},{"name":"c"}]`
		client := newTestClient(t, map[string]string{addrURL: fx3})
		addrs, err := client.ListAddresses(context.Background(), "root", WithPageSize(5))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 3 {
			t.Fatalf("len = %d", len(addrs))
		}
	})

	t.Run("empty_result", func(t *testing.T) {
		client := newTestClient(t, map[string]string{addrURL: `[]`})
		addrs, err := client.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 0 {
			t.Fatalf("len = %d, want 0", len(addrs))
		}
	})

	t.Run("page_callback_fires_correctly", func(t *testing.T) {
		// 5 rows, page size 2. Callback must fire 3 times with
		// cumulative counts (2,1), (4,2), (5,3).
		client := newTestClient(t, map[string]string{addrURL: fx5})

		type call struct{ fetched, page int }
		var calls []call
		cb := func(fetched, page int) {
			calls = append(calls, call{fetched, page})
		}
		_, err := client.ListAddresses(context.Background(), "root",
			WithPageSize(2), WithPageCallback(cb))
		if err != nil {
			t.Fatal(err)
		}
		want := []call{{2, 1}, {4, 2}, {5, 3}}
		if len(calls) != len(want) {
			t.Fatalf("callback calls = %d, want %d", len(calls), len(want))
		}
		for i, w := range want {
			if calls[i] != w {
				t.Errorf("call %d = %+v, want %+v", i, calls[i], w)
			}
		}
	})

	t.Run("default_page_size_no_opts", func(t *testing.T) {
		// 3 rows, no opts → single call with default 1000 page size,
		// under-full short-circuit. No callback, no issues.
		fx := `[{"name":"a"},{"name":"b"},{"name":"c"}]`
		client := newTestClient(t, map[string]string{addrURL: fx})
		addrs, err := client.ListAddresses(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 3 {
			t.Fatalf("len = %d", len(addrs))
		}
	})

	t.Run("pagesize_out_of_range_falls_back", func(t *testing.T) {
		// WithPageSize(-1), WithPageSize(0), WithPageSize(99999) — all
		// should fall back to the default (1000). Verified indirectly:
		// 3-row fixture returns 3 via default page size.
		fx := `[{"name":"a"},{"name":"b"},{"name":"c"}]`
		client := newTestClient(t, map[string]string{addrURL: fx})
		for _, n := range []int{-1, 0, 99999} {
			addrs, err := client.ListAddresses(context.Background(), "root", WithPageSize(n))
			if err != nil {
				t.Fatalf("n=%d: %v", n, err)
			}
			if len(addrs) != 3 {
				t.Errorf("n=%d: len = %d", n, len(addrs))
			}
		}
	})

	t.Run("range_injected_per_page", func(t *testing.T) {
		// Capture every request body on a multi-page ListAddresses fetch
		// and assert each one carries a range parameter. Pilot-level
		// evidence that getPaged actually injects range into the extras
		// map on every iteration. (Extras-merging with fields allowlist
		// gets its own test after Phase D.1 sweeps ListDevices onto
		// getPaged.)
		var sent []map[string]any
		callCount := 0
		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "HTTP_CSRF_TOKEN", Value: "tok", Path: "/"})
			w.Write([]byte(`{}`))
		})
		mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
			callCount++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			params := body["params"].([]any)
			p0 := params[0].(map[string]any)
			sent = append(sent, p0)

			// Serve slices of 5 synthetic rows based on the range param
			rng := p0["range"].([]any)
			offset := int(rng[0].(float64))
			count := int(rng[1].(float64))
			total := 5
			if offset >= total {
				fmt.Fprint(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[]}]}}`)
				return
			}
			end := offset + count
			if end > total {
				end = total
			}
			var rows []string
			for i := offset; i < end; i++ {
				rows = append(rows, fmt.Sprintf(`{"name":"row-%d"}`, i))
			}
			data := "[" + strings.Join(rows, ",") + "]"
			fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":%s}]}}`, data)
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

		addrs, err := c.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 5 {
			t.Fatalf("len = %d, want 5", len(addrs))
		}
		// 5 rows, page size 2 → 3 pages (2+2+1). Terminates on
		// under-full page 3 (1 < 2).
		if callCount != 3 {
			t.Errorf("callCount = %d, want 3", callCount)
		}

		// Every captured request must carry a range parameter.
		want := [][]int{{0, 2}, {2, 2}, {4, 2}}
		for i, p := range sent {
			rng, ok := p["range"].([]any)
			if !ok || len(rng) != 2 {
				t.Errorf("req %d: missing or malformed range: %v", i, p["range"])
				continue
			}
			off := int(rng[0].(float64))
			cnt := int(rng[1].(float64))
			if off != want[i][0] || cnt != want[i][1] {
				t.Errorf("req %d: range=[%d,%d], want %v", i, off, cnt, want[i])
			}
		}
	})

	t.Run("session_expiry_mid_pagination", func(t *testing.T) {
		// Custom stateful server: page 1 returns full fixture; page 2
		// returns -6 Session Expired on first attempt, then the real
		// data on retry (after forwardExtra's automatic re-login).
		//
		// Verifies: getPaged + forwardExtra compose correctly across
		// session boundaries.
		loginCount := 0
		forwardCount := 0
		var bodies []map[string]any

		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
			loginCount++
			http.SetCookie(w, &http.Cookie{Name: "HTTP_CSRF_TOKEN", Value: "tok", Path: "/"})
			w.Write([]byte(`{}`))
		})
		mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
			forwardCount++
			if r.Header.Get("X-CSRFToken") != "tok" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			var req struct {
				Params []map[string]any `json:"params"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			bodies = append(bodies, req.Params[0])

			// Parse range
			offset := 0
			rng := req.Params[0]["range"].([]any)
			offset = int(rng[0].(float64))
			count := int(rng[1].(float64))

			// Page 2 (offset=count) returns session expired ONCE
			if offset > 0 && forwardCount == 2 {
				fmt.Fprint(w, `{"code":0,"data":{"result":[{"status":{"code":-6,"message":"Session expired"},"data":null}]}}`)
				return
			}

			// Generate a page based on offset: 6 total rows, page size = count
			totalRows := 6
			if offset >= totalRows {
				fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[]}]}}`)
				return
			}
			end := offset + count
			if end > totalRows {
				end = totalRows
			}
			var rows []string
			for i := offset; i < end; i++ {
				rows = append(rows, fmt.Sprintf(`{"name":"row-%d"}`, i))
			}
			data := "[" + strings.Join(rows, ",") + "]"
			fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":%s}]}}`, data)
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

		// Page size 3 → page 1 (offset 0) + page 2 (offset 3, expired, retry) + page 3 (offset 6, empty)
		addrs, err := c.ListAddresses(context.Background(), "root", WithPageSize(3))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(addrs) != 6 {
			t.Errorf("len = %d, want 6 (all rows despite page-2 session expiry)", len(addrs))
		}

		// loginCount: 1 initial login + 1 re-login after -6 = 2
		if loginCount != 2 {
			t.Errorf("loginCount = %d, want 2 (initial + re-login)", loginCount)
		}
		// forwardCount: page 1 + page 2 (expired) + page 2 retry + page 3 = 4
		if forwardCount != 4 {
			t.Errorf("forwardCount = %d, want 4", forwardCount)
		}
	})

	t.Run("endpoint_ignores_range_over_full", func(t *testing.T) {
		// Case 1: endpoint returns MORE rows than pageSize on every
		// call. Termination rule 2 (over-full page) stops after the
		// first call.
		callCount := 0
		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "HTTP_CSRF_TOKEN", Value: "tok", Path: "/"})
			w.Write([]byte(`{}`))
		})
		mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
			callCount++
			fmt.Fprint(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[
				{"name":"a"},{"name":"b"},{"name":"c"},{"name":"d"},{"name":"e"}
			]}]}}`)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		c, _ := NewClient(srv.URL, WithCredentials("u", "p"))
		_ = c.Login(context.Background())
		addrs, err := c.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 5 {
			t.Errorf("len = %d, want 5", len(addrs))
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1 (over-full short-circuit)", callCount)
		}
	})

	t.Run("endpoint_ignores_range_exact_size_match", func(t *testing.T) {
		// Case 2: endpoint returns EXACTLY pageSize rows regardless of
		// offset. Over-full detection doesn't fire (5 == 5). Under-full
		// detection doesn't fire (5 == 5). Termination rule 3 — the
		// same-data detection — must catch the infinite loop by
		// byte-comparing page 2 against page 1.
		//
		// Without this detection, a dataset of 2 rows fetched with
		// WithPageSize(2) against /pm/config/adom/{adom}/_package/status
		// loops forever.
		callCount := 0
		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "HTTP_CSRF_TOKEN", Value: "tok", Path: "/"})
			w.Write([]byte(`{}`))
		})
		mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
			callCount++
			// Return exactly 5 rows, regardless of offset in range.
			fmt.Fprint(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[
				{"name":"a"},{"name":"b"},{"name":"c"},{"name":"d"},{"name":"e"}
			]}]}}`)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		c, _ := NewClient(srv.URL, WithCredentials("u", "p"))
		_ = c.Login(context.Background())

		// Page size 5 matches the fixture size exactly. Page 1 returns
		// 5 rows (5 == 5, continue). Page 2 returns identical 5 rows
		// (same-data detection fires, break).
		addrs, err := c.ListAddresses(context.Background(), "root", WithPageSize(5))
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 5 {
			t.Errorf("len = %d, want 5 (no duplicates from page 2)", len(addrs))
		}
		if callCount != 2 {
			t.Errorf("callCount = %d, want 2 (page 1 accepted + page 2 detected as duplicate)", callCount)
		}
	})

	// Sanity: the error-propagation path returns no partial data.
	t.Run("error_discards_accumulated_rows", func(t *testing.T) {
		// Custom server: page 1 ok, page 2 returns permanent -11 permission error.
		callCount := 0
		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "HTTP_CSRF_TOKEN", Value: "tok", Path: "/"})
			w.Write([]byte(`{}`))
		})
		mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				fmt.Fprint(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[{"name":"a"},{"name":"b"}]}]}}`)
				return
			}
			fmt.Fprint(w, `{"code":0,"data":{"result":[{"status":{"code":-11,"message":"no permission"}}]}}`)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		c, _ := NewClient(srv.URL, WithCredentials("u", "p"))
		_ = c.Login(context.Background())

		addrs, err := c.ListAddresses(context.Background(), "root", WithPageSize(2))
		if err == nil {
			t.Fatal("expected error from page 2, got nil")
		}
		if !errors.Is(err, ErrPermission) {
			t.Errorf("err = %v, want ErrPermission", err)
		}
		if addrs != nil {
			t.Errorf("addrs should be nil on error, got %d rows", len(addrs))
		}
	})
}

